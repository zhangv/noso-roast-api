package item

import (
	"hahacking/database"
	"hahacking/database/datatypes"
	"hahacking/models"
	"hahacking/utils"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Item struct {
	models.Model
	Name         string         `json:"name"`
	Code         string         `json:"code"`
	CategoryPath string         `json:"categorypath" gorm:"column:categorypath"`
	CategoryID   string         `json:"category_id"`
	ParentID     string         `json:"parent_id"` //deprecated
	Price        float64        `json:"price" gorm:"type:decimal(10,2)"`
	Description  string         `json:"description" gorm:"column:description;type:text"`
	Status       ItemStatus     `json:"status"`
	ShopID       string         `json:"shop_id"`
	SKUAtt1      string         `gorm:"type:varchar(200);column:skuatt1" json:"skuatt1"`
	SKUAtt2      string         `gorm:"type:varchar(200);column:skuatt2" json:"skuatt2"`
	SKUAtt3      string         `gorm:"type:varchar(200);column:skuatt3" json:"skuatt3"`
	SKUAttValue1 string         `gorm:"type:varchar(1000);column:skuAttValue1" json:"skuattvalue1"`
	SKUAttValue2 string         `gorm:"type:varchar(1000);column:skuAttValue2" json:"skuattvalue2"`
	SKUAttValue3 string         `gorm:"type:varchar(1000);column:skuAttValue3" json:"skuattvalue3"`
	Pics         pq.StringArray `gorm:"type:text" json:"pics"`
	Tags         pq.StringArray `gorm:"type:text" json:"tags"`
	Sequence     string         `json:"sequence"`
	Stockable    ItemStockable  `json:"stockable"`                                  //是否限制库存
	MenuTags     datatypes.Slot `gorm:"type:text;column:menuTags" json:"menu_tags"` //门店菜单标签

	//以下字段原本在attribute中
	EnName  string `json:"enname" gorm:"column:enname"`
	Country string `json:"country" gorm:"column:country"` //国家iso code

}

type ItemStatus string

const (
	ItemStatusCreate  ItemStatus = "create"
	ItemStatusPublish ItemStatus = "publish"
	ItemStatusPending ItemStatus = "pending"
)

var ItemStatuses = map[ItemStatus]string{
	ItemStatusCreate:  "创建",
	ItemStatusPublish: "上架",
	ItemStatusPending: "下架",
}

type ItemStockable string

const (
	ItemStockableYes ItemStockable = "y"
	ItemStockableNo  ItemStockable = "n"
)

var ItemStockables = map[ItemStockable]string{
	ItemStockableYes: "是",
	ItemStockableNo:  "否",
}

type SearchForm struct {
	utils.Pager
	Name       string `form:"name"`
	Code       string `form:"code"`
	Status     string `form:"status"`
	CategoryID string `form:"category_id"`
	Tags       string `form:"tags"`
	MenuTags   string `form:"menu_tags"`
}

func Search(q SearchForm) ([]Item, int) {
	db := database.GetConn()
	var count int64
	var os []Item
	_, pagesize, start, sort := q.Pager.Paging()
	var statement = db.Model(&Item{})
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Code != "" {
		statement = statement.Where("code = ?", q.Code)
	}
	if q.CategoryID != "" {
		statement = statement.Where("category_id = ?", q.CategoryID)
	}
	if q.Tags != "" {
		statement = statement.Where("tags LIKE ?", "%"+q.Tags+"%")
	}
	if q.MenuTags != "" {
		statement = statement.Where("menu_tags LIKE ?", "%"+q.MenuTags+"%")
	}
	statement.Where(&q).Count(&count).Order(sort).Offset(start).Limit(pagesize).Find(&os)
	return os, int(count)
}

func Create(o *Item) (*Item, error) {
	db := database.GetConn()
	if err := db.Create(o).Error; err != nil {
		return nil, err
	}
	return o, nil
}

func Update(id int, o *Item) error {
	db := database.GetConn()
	if err := db.Model(&o).Where("id = ?", id).Updates(o).Error; err != nil {
		return err
	}
	return nil
}

func GetById(id uint) (*Item, error) {
	var o Item
	db := database.GetConn()
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func (o *Item) Delete() error {
	db := database.GetConn()
	if err := db.Delete(&Item{}, o.ID).Error; err != nil {
		return err
	}
	return nil
}
