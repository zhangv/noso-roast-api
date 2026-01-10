package item

import (
	"errors"
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"

	"gorm.io/gorm"
)

type ItemStock struct {
	models.Model
	ItemID    uint          `json:"item_id"`
	Item      Item          `json:"item" gorm:"foreignKey:ItemID;references:ID"`
	ItemSKUID uint          `json:"item_sku_id" gorm:"column:itemsku_id"`
	ItemSKU   ItemSKU       `json:"item_sku" gorm:"foreignKey:ItemSKUID;references:ID"`
	Qty       int           `json:"qty"`
	Unit      string        `json:"unit" gorm:"column:unit"`
	Status    models.Status `json:"status"`
}

func GetStockByItemID(itemID string) ([]ItemStock, error) {
	var o []ItemStock
	db := database.GetConn()
	err := db.Where("item_id = ? and status = ?", itemID, models.StatusValid).Find(&o).Error
	if err != nil {
		return []ItemStock{}, err
	}
	return o, nil
}

func GetStockByItemSKUID(itemSKUID uint) ([]ItemStock, error) {
	var o []ItemStock
	db := database.GetConn()
	err := db.Where("itemSKU_id = ? and status = ?", itemSKUID, models.StatusValid).Find(&o).Error
	if err != nil {
		return []ItemStock{}, err
	}
	return o, nil
}

type SearchStockForm struct {
	utils.Pager
	ItemID    string        `form:"item_id"`
	Status    models.Status `form:"status"`
	ItemSKUID string        `form:"itemSKU_id"`
}

func SearchStock(q SearchStockForm) ([]ItemStock, int) {
	db := database.GetConn()
	var count int64
	var os []ItemStock
	var statement = db.Model(&ItemStock{})
	if q.ItemID != "" {
		statement = statement.Where("item_id = ?", q.ItemID)
	}

	if q.ItemSKUID != "" {
		statement = statement.Where("itemSKU_id = ?", q.ItemSKUID)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}

	_, pagesize, start, sort := q.Pager.Paging()
	statement.Where(&q).Count(&count).Order(sort).Offset(start).Limit(pagesize).Find(&os)
	return os, int(count)
}

func CreateStock(o *ItemStock) (*ItemStock, error) {
	db := database.GetConn()
	if err := db.Create(o).Error; err != nil {
		return nil, err
	}
	return o, nil
}

func UpdateStock(id int, o *ItemStock) error {
	db := database.GetConn()
	if err := db.Model(&o).Where("id = ?", id).Updates(o).Error; err != nil {
		return err
	}
	return nil
}

func GetStockById(id uint) (*ItemStock, error) {
	var o ItemStock
	db := database.GetConn()
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func (o *ItemStock) Delta(n int) int {
	db := database.GetConn()
	tx := db.Begin()
	err := tx.Model(&o).Where("id = ?", o.ID).Update("qty", o.Qty-n).Error
	if err != nil {
		tx.Rollback()
		return n
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return n
	}
	o.Qty = o.Qty - n
	return o.Qty
}
func (o *ItemStock) Incr(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("n must be greater than 0")
	}
	return o.Delta(n), nil
}
func (o *ItemStock) Decr(n int) (int, error) {
	if n > 0 {
		return 0, errors.New("n must be greater than 0")
	}
	return o.Delta(n), nil
}
func (o *ItemStock) SetQty(n int) int {
	db := database.GetConn()
	tx := db.Begin()
	err := tx.Model(&o).Where("id = ?", o.ID).Update("qty", n).Error
	if err != nil {
		tx.Rollback()
		return n
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return n
	}
	o.Qty = n
	return o.Qty
}

func (o *ItemStock) Delete() error {
	db := database.GetConn()
	if err := db.Delete(&ItemStock{}, o.ID).Error; err != nil {
		return err
	}
	return nil
}
