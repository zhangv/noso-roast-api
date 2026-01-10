package item

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"

	"gorm.io/gorm"
)

type ItemSKU struct {
	models.Model
	ItemID  uint          `json:"item_id"`
	Item    Item          `json:"item" gorm:"foreignKey:ItemID;references:ID"`
	Price   float64       `json:"price" gorm:"type:decimal(10,2)"`
	SkuAtt1 string        `gorm:"type:varchar(200);column:skuatt1" json:"skuatt1"`
	SkuAtt2 string        `gorm:"type:varchar(200);column:skuatt2" json:"skuatt2"`
	SkuAtt3 string        `gorm:"type:varchar(200);column:skuatt3" json:"skuatt3"`
	Status  models.Status `json:"status"`
}

func GetSKUByAtts(itemID string, att1 string, att2 string, att3 string) (*ItemSKU, error) {
	var o ItemSKU
	db := database.GetConn()
	err := db.Where("item_id = ? and skuatt1 = ? and skuatt2 = ? and skuatt3 = ? and status = ?", itemID, att1, att2, att3, models.StatusValid).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func GetSKUByItemID(itemID string) ([]ItemSKU, error) {
	var o []ItemSKU
	db := database.GetConn()
	err := db.Where("item_id = ? and status = ?", itemID, models.StatusValid).Find(&o).Error
	if err != nil {
		return []ItemSKU{}, err
	}
	return o, nil
}

func (o *ItemSKU) GetStocks() ([]ItemStock, error) {
	stocks, err := GetStockByItemSKUID(o.ID)
	if err != nil {
		return []ItemStock{}, err
	}
	return stocks, nil
}

func (o *ItemSKU) InitStock(qty int, unit string) (*[]ItemStock, error) {
	stocks, err := o.GetStocks()
	if err != nil {
		return nil, err
	}
	if len(stocks) == 0 {
		stock := ItemStock{
			ItemID:    o.ItemID,
			ItemSKUID: o.ID,
			Qty:       qty,
			Unit:      unit,
			Status:    models.StatusValid,
		}
		newstock, err := CreateStock(&stock)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, *newstock)
	} else {
		for _, stock := range stocks {
			stock.SetQty(qty)
		}
	}

	return &stocks, nil
}

type SearchSKUForm struct {
	utils.Pager
	ItemID  string        `form:"item_id"`
	Status  models.Status `form:"status"`
	SkuAtt1 string        `form:"skuatt1"`
	SkuAtt2 string        `form:"skuatt2"`
	SkuAtt3 string        `form:"skuatt3"`
	//模式2: 批量查询
	IDs     []string `form:"ids"`
	ItemIDs []string `form:"item_ids"`
}

func SearchSKU(q SearchSKUForm) ([]ItemSKU, int) {
	db := database.GetConn()
	var count int64
	var os []ItemSKU
	var statement = db.Model(&ItemSKU{})
	if q.ItemID != "" {
		statement = statement.Where("item_id = ?", q.ItemID)
	}
	if q.ItemIDs != nil {
		statement = statement.Where("item_id IN (?)", q.ItemIDs)
	}
	if q.IDs != nil {
		statement = statement.Where("id IN (?)", q.IDs)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	if q.SkuAtt1 != "" {
		statement = statement.Where("skuatt1 = ?", q.SkuAtt1)
	}
	if q.SkuAtt2 != "" {
		statement = statement.Where("skuatt2 = ?", q.SkuAtt2)
	}
	if q.SkuAtt3 != "" {
		statement = statement.Where("skuatt3 = ?", q.SkuAtt3)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	statement.Where(&q).Count(&count).Order(sort).Offset(start).Limit(pagesize).Find(&os)
	return os, int(count)
}

func CreateSKU(o *ItemSKU) (*ItemSKU, error) {
	db := database.GetConn()
	if err := db.Create(o).Error; err != nil {
		return nil, err
	}
	return o, nil
}

func UpdateSKU(id int, o *ItemSKU) error {
	db := database.GetConn()
	if err := db.Model(&o).Where("id = ?", id).Updates(o).Error; err != nil {
		return err
	}
	return nil
}

func GetSKUById(id uint) (*ItemSKU, error) {
	var o ItemSKU
	db := database.GetConn()
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func (o *ItemSKU) Delete() error {
	db := database.GetConn()
	if err := db.Delete(&ItemSKU{}, o.ID).Error; err != nil {
		return err
	}
	return nil
}
