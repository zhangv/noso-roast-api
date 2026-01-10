package bean

import (
	"errors"
	"hahacking/database"
	"hahacking/database/datatypes"
	"hahacking/models"
	"hahacking/utils"
	"strconv"
)

type BeanBatch struct {
	models.Model
	BeanID     uint                  `json:"bean_id"`
	Bean       *Bean                 `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	Seq        int                   `json:"seq"`
	Date       string                `json:"date"`
	BeanType   BeanType              `json:"bean_type"`
	Status     BatchStatus           `json:"status"`
	StatusL    string                `json:"statusL" gorm:"-"`
	TotalPrice float64               `json:"totalprice" gorm:"column:totalprice;type:decimal(10,2)"`
	TotalQty   int                   `json:"totalqty" gorm:"column:totalqty;default:0;not null;type:int"`
	CurrentQty int                   `json:"currentqty" gorm:"column:currentqty;default:0;not null;type:int"` //批次在更新前的数量
	Memo       string                `json:"memo"`
	Type       StockLogType          `json:"type"`
	Receipts   datatypes.StringSlice `json:"receipts" gorm:"type:text"` //采购凭证图片
}

// 记录批次数量变化：采购生/熟豆+、烘焙生豆-、销售熟豆订单-、撤销销售熟豆订单+
type BeanBatchLog struct {
	models.Model
	BeanBatchID uint       `json:"beanbatch_id" gorm:"column:beanbatch_id"`
	BeanBatch   *BeanBatch `json:"beanbatch" gorm:"foreignKey:BeanBatchID;references:ID"`
	BeanID      uint       `json:"bean_id" gorm:"column:bean_id"`
	Bean        *Bean      `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	BeanType    BeanType   `json:"bean_type" gorm:"column:bean_type"`
	Qty         int        `json:"qty" gorm:"column:qty"`
	RefType     string     `json:"ref_type"` //来源类型
	RefID       string     `json:"ref_id"`   //来源ID
}

type BatchStatus string

const (
	BatchStatusInvalid BatchStatus = "invalid" // 无效
	BatchStatusValid   BatchStatus = "valid"   // 有效
)

var BatchStatuses = map[BatchStatus]string{
	BatchStatusInvalid: "无效",
	BatchStatusValid:   "有效",
}

type BatchType string

const (
	BatchTypePurchase BatchType = "purchase"
	BatchTypeProduce  BatchType = "produce"
	BatchTypeSale     BatchType = "sale"
	BatchTypeReturn   BatchType = "return"
	BatchTypeTransfer BatchType = "transfer"
)

var BatchTypes = map[BatchType]string{
	BatchTypePurchase: "采购",
	BatchTypeProduce:  "生产",
	BatchTypeSale:     "销售",
	BatchTypeReturn:   "退货",
	BatchTypeTransfer: "转移",
}

type SearchBatchForm struct {
	utils.Pager
	BeanID   int       `form:"bean_id"`
	Date     string    `form:"date"`
	Status   string    `form:"status"`
	Type     BatchType `form:"type"`
	BeanType string    `form:"bean_type"`
}

func GetBatchById(id uint) (*BeanBatch, error) {
	db := database.Conn
	var o BeanBatch
	if err := db.Model(&BeanBatch{}).Where("id = ?", id).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func SearchBatch(q SearchBatchForm) ([]BeanBatch, int) {
	db := database.Conn
	var os []BeanBatch
	var statement = db.Model(&BeanBatch{}).Preload("Bean")
	if q.Date != "" {
		statement = statement.Where("date LIKE ?", "%"+q.Date+"%")
	}
	if q.BeanID != 0 {
		statement = statement.Where("bean_id = ?", q.BeanID)
	}
	if q.Type != "" {
		statement = statement.Where("type = ?", q.Type)
	}
	if q.BeanType != "" {
		statement = statement.Where("bean_type = ?", q.BeanType)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
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

func UpdateBatch(o *BeanBatch) (*BeanBatch, error) {
	db := database.Conn

	if o.ID != 0 {
		if err := db.Model(&BeanBatch{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return nil, err
		}
	} else {
		o.Status = BatchStatusValid
		o.CurrentQty = o.TotalQty
		if err := db.Create(&o).Error; err != nil {
			return nil, err
		}
	}

	bean, err := GetById(o.BeanID)
	if err != nil {
		return nil, err
	}
	delta := o.TotalQty
	memo := "批次入库"
	if contains(StockOutTypes(), string(o.Type)) {
		delta = -delta
		memo = "批次出库"
	}

	bean.Stocking(delta, o.BeanType, o.Type, "beanbatch", strconv.Itoa(int(o.ID)), memo)
	return o, nil
}

// 撤销批次 - 需要修改库存回去并且记录日志
func (batch *BeanBatch) Undo() error {
	db := database.Conn
	bean, err := GetById(batch.BeanID)
	if err != nil {
		return err
	}
	bean.Stocking(-batch.TotalQty, batch.BeanType, StockLogTypeUndo, "beanbatch", strconv.Itoa(int(batch.ID)), "批次撤销"+strconv.Itoa(int(batch.ID)))
	return db.Model(&BeanBatch{}).Where("id = ?", batch.ID).Update("status", BatchStatusInvalid).Error
}

func (batch *BeanBatch) Log(qty int, refType string, refID string, memo string) error {
	db := database.Conn
	// TODO 应该增加个字段
	if refType != "user" { //用户操作不检查重复 - 因为同一个人可能会多次操作入库
		var exist BeanBatchLog
		db.Model(&BeanBatchLog{}).
			Where("ref_type = ? AND ref_id = ? AND beanbatch_id = ? AND bean_type = ?",
				refType, refID, batch.ID, batch.BeanType).First(&exist)
		if exist.ID > 0 {
			return errors.New("重复记录")
		}
	}
	db.Create(&BeanBatchLog{
		BeanBatchID: batch.ID,
		BeanType:    batch.BeanType,
		BeanID:      batch.BeanID,
		Qty:         qty,
		RefType:     refType,
		RefID:       refID,
	})
	batch.CurrentQty += qty
	db.Save(batch)
	return nil
}

type SearchBatchLogForm struct {
	utils.Pager
	BeanBatchID uint     `form:"beanbatch_id"`
	BeanID      uint     `form:"bean_id"`
	BeanType    BeanType `form:"bean_type"`
	RefType     string   `form:"ref_type"`
	RefID       string   `form:"ref_id"`
}

func SearchBatchLog(q SearchBatchLogForm) ([]BeanBatchLog, int) {
	db := database.Conn
	var os []BeanBatchLog
	statement := db.Model(&BeanBatchLog{}).Preload("BeanBatch").Preload("Bean")
	if q.BeanBatchID != 0 {
		statement = statement.Where("beanbatch_id = ?", q.BeanBatchID)
	}
	if q.BeanID != 0 {
		statement = statement.Where("bean_id = ?", q.BeanID)
	}
	if q.BeanType != "" {
		statement = statement.Where("bean_type = ?", q.BeanType)
	}
	if q.RefType != "" {
		statement = statement.Where("ref_type = ?", q.RefType)
	}
	if q.RefID != "" {
		statement = statement.Where("ref_id = ?", q.RefID)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	statement.
		Order(sort).
		Offset(start).
		Limit(pagesize).Find(&os)
	var count int64
	statement.Count(&count)
	return os, int(count)
}
