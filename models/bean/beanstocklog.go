package bean

import (
	"errors"
	"hahacking/database"
	"hahacking/models"

	"hahacking/utils"
)

const (
	FieldRawStock     = "RawStock"
	FieldRoastedStock = "RoastedStock"
)

type BeanStockLog struct {
	models.Model
	ID          uint           `json:"id" form:"id" gorm:"primaryKey"`
	BeanID      uint           `json:"bean_id"`
	Bean        *Bean          `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	BeanBatchID uint           `json:"beanbatch_id"` //如果是减库存（销售熟豆，烘焙生豆），必须要有batch信息
	BeanType    BeanType       `json:"bean_type" gorm:"column:bean_type"`
	Qty         int            `json:"qty"`
	FromQty     int            `json:"fromqty" gorm:"column:fromqty"`
	ToQty       int            `json:"toqty" gorm:"column:toqty"`
	Type        StockLogType   `json:"type"`
	Status      StockLogStatus `json:"status"`
	StatusL     string         `json:"statusL" gorm:"-"`
	SrcType     string         `json:"src_type"`
	SrcID       string         `json:"src_id"`
	Memo        string         `json:"memo"`
}
type StockLogStatus string

const (
	StockLogStatusValid   StockLogStatus = "valid"
	StockLogStatusInvalid StockLogStatus = "invalid"
)

type StockLogType string

const (
	StockLogTypeProduce     StockLogType = "produce"
	StockLogTypePurchase    StockLogType = "purchase"
	StockLogTypeSale        StockLogType = "sale"
	StockLogTypeReturn      StockLogType = "return"
	StockLogTypeTransferIn  StockLogType = "transferin"
	StockLogTypeTransferOut StockLogType = "transferout"
	StockLogTypeWaste       StockLogType = "waste"
	StockLogTypeUndo        StockLogType = "undo" //撤销
)

var StockLogTypes = map[StockLogType]string{
	StockLogTypeProduce:     "生产",
	StockLogTypePurchase:    "采购",
	StockLogTypeSale:        "销售",
	StockLogTypeReturn:      "退货",
	StockLogTypeTransferIn:  "转入",
	StockLogTypeTransferOut: "转出",
	StockLogTypeWaste:       "损耗",
	StockLogTypeUndo:        "撤销", //注意撤销可能是入库，也可能是出库
}
var StockLogInTypes = map[StockLogType]string{
	StockLogTypeProduce:    "生产",
	StockLogTypePurchase:   "采购",
	StockLogTypeReturn:     "退货",
	StockLogTypeTransferIn: "转入",
}
var StockLogOutTypes = map[StockLogType]string{
	StockLogTypeSale:        "销售",
	StockLogTypeTransferOut: "转出",
	StockLogTypeWaste:       "损耗",
}

func StockInTypes() []string {
	inTypes := []string{}
	for k := range StockLogInTypes {
		inTypes = append(inTypes, string(k))
	}
	return inTypes
}
func StockOutTypes() []string {
	outTypes := []string{}
	for k := range StockLogOutTypes {
		outTypes = append(outTypes, string(k))
	}
	return outTypes
}

type SearchStockLogForm struct {
	utils.Pager
	BeanID   int    `form:"bean_id"`
	BeanName string `form:"bean_name"`
	BeanType string `form:"bean_type"`
	Date     string `form:"date"`
	Status   string `form:"status"`
	Type     string `form:"type"`
	SrcType  string `form:"src_type"`
	SrcID    string `form:"src_id"`
}

func SearchStockLog(q SearchStockLogForm) ([]BeanStockLog, int) {
	db := database.Conn
	var os []BeanStockLog

	var statement = db.Model(&BeanStockLog{}).Preload("Bean")
	if q.Date != "" {
		statement = statement.Where("created_at LIKE ?", "%"+q.Date+"%")
	}
	if q.BeanID != 0 {
		statement = statement.Where("bean_id = ?", q.BeanID)
	}
	if q.BeanName != "" {
		statement = statement.Joins("Bean").Where("Bean.name LIKE ?", "%"+q.BeanName+"%")
	}
	if q.BeanType != "" {
		statement = statement.Where("bean_type = ?", q.BeanType)
	}
	if q.Type != "" {
		statement = statement.Where("type = ?", q.Type)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	if q.SrcType != "" {
		statement = statement.Where("src_type = ?", q.SrcType)
	}
	if q.SrcID != "" {
		statement = statement.Where("src_id = ?", q.SrcID)
	}
	_, pagesize, start, sort := q.Pager.Paging()

	var count int64
	statement.
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

// qty的正负由外部逻辑决定
func (bean *Bean) Stocking(qty int, beanType BeanType, logType StockLogType, srcType string, srcID string, memo string) (*Bean, error) {
	db := database.Conn

	// 避免重复入库
	if srcType != "user" { //用户操作不检查重复
		var exist BeanStockLog
		db.Model(&BeanStockLog{}).Where("type = ? AND src_type = ? AND src_id = ? AND bean_id = ? AND bean_type = ?", logType, srcType, srcID, bean.ID, beanType).First(&exist)
		if exist.ID > 0 {
			return nil, errors.New("重复库存修改记录")
		}
	}

	fromQty := 0
	toQty := 0

	// Get fresh data after locking
	var freshBean Bean
	if err := db.First(&freshBean, bean.ID).Error; err != nil {
		// tx.Rollback()
		return nil, err
	}

	switch beanType {
	case BeanTypeGreen:
		fromQty = freshBean.RawStock
		toQty = freshBean.RawStock + qty

		if err := db.Model(&bean).Update("rawstock", toQty).Error; err != nil {
			// tx.Rollback()
			return nil, err
		}
		// log.Printf("xxxxxxxxxxxxfromQty: %+v", fromQty)
	case BeanTypeRoasted:
		fromQty = freshBean.RoastedStock
		toQty = freshBean.RoastedStock + qty

		if err := db.Model(&bean).Update("roastedstock", toQty).Error; err != nil {
			// tx.Rollback()
			return nil, err
		}
	default:
		return nil, errors.New("不支持的豆子类型")
	}

	log := BeanStockLog{
		BeanID:   bean.ID,
		BeanType: beanType,
		Qty:      qty,
		FromQty:  fromQty,
		ToQty:    toQty,
		Type:     logType,
		Status:   StockLogStatusValid,
		SrcType:  srcType,
		SrcID:    srcID,
		Memo:     memo,
	}
	if err := db.Create(&log).Error; err != nil {
		// tx.Rollback()
		return nil, err
	}

	// if err := tx.Commit().Error; err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }

	return bean, nil
}
