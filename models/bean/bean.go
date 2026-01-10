package bean

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/price"
	"hahacking/models/supplier"
	"hahacking/models/tag"
	"hahacking/utils"
	"math"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Bean struct {
	ID   uint   `json:"id" form:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Code string `json:"code"`
	//这个字段应该放到batch里
	Type    BeanType `json:"type"`
	Country string   `json:"country"`
	Status  string   `json:"status"`
	// StatusL           string        `json:"statusL" gorm:"-"`
	Processor    string `json:"processor"`
	Provider     string `json:"provider"`
	Variety      string `json:"variety"`
	Process      string `json:"process"`
	OtherProcess string `json:"otherprocess" gorm:"column:otherprocess"`
	// ProcessL          string        `json:"processL" gorm:"-"`
	Moisture          float64       `json:"moisture"`
	Density           float64       `json:"density"` //堆积密度 单位：g/cm3
	Mesh              string        `json:"mesh"`    //目数 14/18/20/24/30/40/60/80/100/120/140/170/200/230/270/325/400/500/630/800/1000/1250/1500/2000
	RawStock          int           `json:"rawstock" gorm:"column:rawstock"`
	RawStockAlert     int           `json:"rawstockalert" gorm:"column:rawstockalert;default:2000"` //生豆库存预警值
	RoastedStock      int           `json:"roastedstock" gorm:"column:roastedstock"`
	RoastedStockAlert int           `json:"roastedstockalert" gorm:"column:roastedstockalert;default:500"` //熟豆库存预警值
	PurchasePrice     float64       `json:"purchaseprice" gorm:"column:purchaseprice;type:decimal(10,2)"`
	BPrice            float64       `json:"bprice" gorm:"column:bprice;type:decimal(10,2)"`
	CPrice            float64       `json:"cprice" gorm:"column:cprice;type:decimal(10,2)"`
	Owner             string        `json:"owner"`
	Description       string        `json:"description"`
	Prices            []price.Price `json:"prices" gorm:"-"`
	Tags              []tag.Tag     `json:"tags" gorm:"-"`
	Cover             string        `json:"cover"` //封面图片url

	//标签 始终报错
	// Taggings []tag.Tagging `gorm:"polymorphic:Target"`
	// Tags     []tag.Tag     `gorm:"many2many:t_taggings;joinForeignKey:TargetID;joinReferences:TagID;polymorphic:Target;polymorphicValue:bean"`
}

type BeanType string

var BeanTypes = map[BeanType]string{
	BeanTypeGreen:   "生豆",
	BeanTypeRoasted: "熟豆",
}

const (
	BeanTypeGreen   BeanType = "green"
	BeanTypeRoasted BeanType = "roasted"
	BeanTypeBlended BeanType = "blended"
	BeanTypeGround  BeanType = "ground"
	BeanTypeExtract BeanType = "extract"
	BeanTypeCapsule BeanType = "capsule"
	BeanTypePod     BeanType = "pod"
	BeanTypeOther   BeanType = "other"
)

type BeanStatus string

const (
	BeanStatusValid   BeanStatus = "valid"
	BeanStatusInvalid BeanStatus = "invalid"
)

var BeanStatuses = map[BeanStatus]string{
	BeanStatusValid:   "有效",
	BeanStatusInvalid: "无效",
}

func (b Bean) IsAuditable() {}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "BeanStatus" {
			result[n] = models.ConvertMapToArray(BeanStatuses)
		}
		if n == "BeanType" {
			result[n] = models.ConvertMapToArray(BeanTypes)
		}
		if n == "BatchStatus" {
			result[n] = models.ConvertMapToArray(BatchStatuses)
		}
		if n == "BatchType" {
			result[n] = models.ConvertMapToArray(BatchTypes)
		}
		if n == "StockLogType" {
			result[n] = models.ConvertMapToArray(StockLogTypes)
		}
		if n == "StockLogInType" {
			result[n] = models.ConvertMapToArray(StockLogInTypes)
		}
		if n == "StockLogOutType" {
			result[n] = models.ConvertMapToArray(StockLogOutTypes)
		}
	}
	return result
}

func (s *Bean) AfterFind(tx *gorm.DB) (err error) {
	// s.StatusL = translateDeliveryType(s.Status)
	// s.ProcessL = translateDeliveryType(s.Process)
	return
}
func (s *Bean) AfterUpdate(tx *gorm.DB) (err error) {
	var exprovider supplier.Supplier
	err = tx.Where("name = ?", s.Provider).First(&exprovider).Error
	if err != nil && err != gorm.ErrRecordNotFound { //已经存在该名称的供应商记录
	} else {
		if exprovider.ID == 0 {
			supplier.Update(&supplier.Supplier{
				Name:        s.Provider,
				Status:      supplier.StatusValid,
				Type:        supplier.SupplierTypeBean,
				Description: "生豆供应商（" + s.Name + ")",
			})
		} else {
			if exprovider.Status != supplier.StatusValid {
				exprovider.Status = supplier.StatusValid
				tx.Save(&exprovider)
			}
		}
	}
	return nil
}
func translateDeliveryType(deliveryType string) string {
	switch deliveryType {
	case "express":
		return "快递"
	case "pickup":
		return "自提"
	default:
		return deliveryType
	}
}

// 获取当前 豆子种类数量、生豆数量/价值、熟豆数量价值
func GetTotal() (map[string]any, error) {
	db := database.Conn

	// 定义结果结构体
	type TotalResult struct {
		Count              int64   `json:"count"`
		GreenTotal         int64   `json:"green_total"`
		RoastedTotal       int64   `json:"roasted_total"`
		GreenTotalAmount   float64 `json:"green_total_amount"`
		RoastedTotalAmount float64 `json:"roasted_total_amount"`
		GreenCount         int64   `json:"green_count"`   // rawstock > 0 的个数
		RoastedCount       int64   `json:"roasted_count"` // roastedstock > 0 的个数
	}

	var result TotalResult

	// 使用单个SQL查询获取所有统计数据
	err := db.Model(&Bean{}).
		Where("status = ?", BeanStatusValid).
		Select(`
			COUNT(*) as count,
			COALESCE(SUM(IFNULL(rawstock, 0)), 0) as green_total,
			COALESCE(SUM(IFNULL(roastedstock, 0)), 0) as roasted_total,
			COALESCE(SUM(IFNULL(purchaseprice, 0) * IFNULL(rawstock, 0) /1000), 0) as green_total_amount,
			COALESCE(SUM(IFNULL(cprice, 0) * IFNULL(roastedstock, 0) /1000), 0) as roasted_total_amount,
			COUNT(CASE WHEN IFNULL(rawstock, 0) > 0 THEN 1 END) as green_count,
			COUNT(CASE WHEN IFNULL(roastedstock, 0) > 0 THEN 1 END) as roasted_count
		`).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	greenTotalAmount, _ := decimal.NewFromFloat(result.GreenTotalAmount).Round(2).Float64()
	roastedTotalAmount, _ := decimal.NewFromFloat(result.RoastedTotalAmount).Round(2).Float64()
	return map[string]any{
		"total":              result.Count,
		"greenTotal":         result.GreenTotal,
		"roastedTotal":       result.RoastedTotal,
		"greenTotalAmount":   greenTotalAmount,
		"roastedTotalAmount": roastedTotalAmount,
		"greenCount":         result.GreenCount,
		"roastedCount":       result.RoastedCount,
	}, nil
}
func GetById(id uint) (*Bean, error) {
	var o Bean
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
func GetByCode(code string) (*Bean, error) {
	var o Bean
	db := database.Conn
	err := db.Where("code = ?", code).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchForm struct {
	utils.Pager
	Name   string `form:"name"`
	Code   string `form:"code"`
	Status string `form:"status"`
	Owner  string `form:"owner"`
}

func Search(q SearchForm) ([]Bean, int) {
	db := database.Conn
	var os []Bean

	var statement = db.Model(&Bean{})
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Code != "" {
		statement = statement.Where("code LIKE ?", "%"+q.Code+"%")
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	} else {
		statement = statement.Where("status != ?", "invalid")
	}
	if q.Owner != "" {
		statement = statement.Where("owner = ?", q.Owner)
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

func Update(o *Bean) (*Bean, error) {
	db := database.Conn
	//TODO 重构为form并validate
	if o.ID != 0 {
		if err := db.Model(&Bean{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return nil, err
		}
	}
	o.GeneratePrices(false)
	return o, nil
}

func (bean Bean) GeneratePrices(overwrite bool) (err error) {
	db := database.Conn
	f := price.SearchLevelForm{Status: "valid"}
	allLevels, _ := price.SearchLevel(f)
	for _, level := range allLevels {
		base := bean.CPrice //统一为零售基础价为基底
		priceValue := base * level.Discount
		//取整
		priceValue = math.Round(priceValue)
		newprice := price.Price{BeanID: bean.ID, PriceLevelID: level.ID, Price: priceValue, Status: "valid"}
		var existingPrice price.Price
		if err := db.Where("bean_id = ? AND pricelevel_id = ?", bean.ID, level.ID).First(&existingPrice).Error; err == nil {
			// Update the existing price
			if overwrite {
				existingPrice.Price = priceValue
				if err := db.Save(&existingPrice).Error; err != nil {
					return err
				}
			}
			continue
		} else if err := db.Create(&newprice).Error; err != nil {
			return err
		}
	}

	return nil
}
func GetAllPriceList(leveltype price.LevelType) ([]price.Price, error) {
	var prices []price.Price
	form := price.SearchForm{
		Status: "valid",
	}
	if leveltype != "" {
		form.LevelType = string(leveltype)
	}
	prices, _ = price.Search(form)
	return prices, nil
}
func (bean Bean) GetPriceList(leveltype price.LevelType) ([]price.Price, error) {
	var prices []price.Price
	form := price.SearchForm{
		Status: "valid",
		BeanID: bean.ID,
	}
	if leveltype != "" {
		form.LevelType = string(leveltype)
	}
	prices, _ = price.Search(form)
	return prices, nil
}

func GetPrices(ids []uint, levelid uint) (map[uint]float64, error) {
	var prices []price.Price
	db := database.Conn
	var statement = db.Model(&price.Price{})
	statement.Where("bean_id IN (?)", ids)
	statement.Where("pricelevel_id", levelid)
	err := statement.Find(&prices).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	var mapping = make(map[uint]float64)
	for _, price := range prices {
		mapping[price.BeanID] = price.Price
	}
	return mapping, nil
}

func (bean Bean) Delete() error {
	db := database.Conn
	if err := db.Delete(&Bean{}, bean.ID).Error; err != nil {
		return err
	}
	return nil
}
func (bean Bean) Remove() error {
	db := database.Conn
	if err := db.Model(&Bean{}).Where("id = ?", bean.ID).Updates(map[string]any{"status": "invalid"}).Error; err != nil {
		return err
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
