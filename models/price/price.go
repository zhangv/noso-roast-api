package price

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"
	"strings"

	"gorm.io/gorm"
)

type Price struct {
	ID uint `json:"id" form:"id" gorm:"primaryKey"`
	models.Model
	BeanID       uint       `json:"bean_id" gorm:"not null;uniqueIndex:idx_bean_id_pricelevel_id"`
	PriceLevelID uint       `json:"pricelevel_id" gorm:"column:pricelevel_id;not null;uniqueIndex:idx_bean_id_pricelevel_id"`
	PriceLevel   PriceLevel `json:"pricelevel" gorm:"foreignKey:PriceLevelID;references:ID"`
	Price        float64    `json:"price"`
	Status       string     `json:"status"`
}

type SearchForm struct {
	utils.Pager
	BeanID    uint   `form:"bean_id"`
	LevelType string `form:"level_type"`
	Status    string `form:"status"`
}

func Search(q SearchForm) ([]Price, int) {
	db := database.Conn

	var statement = db.Model(&Price{})
	if q.LevelType != "" {
		statement = statement.Joins("JOIN t_price_level ON t_price_level.id = t_price.pricelevel_id").Where("t_price_level.type = ?", q.LevelType)
	}
	if q.BeanID != 0 {
		statement = statement.Where("bean_id = ?", q.BeanID)
	}
	statement.Preload("PriceLevel")

	_, pagesize, start, sort := q.Pager.Paging()

	// if q.Pager.SortBy == "level_name" {
	// 	sort = "t_price_level.name " + q.Pager.SortDesc
	// }
	var count int64
	var os []Price
	statement.
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}
func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "LevelType" {
			result[n] = models.ConvertMapToArray(LevelTypes)
		}
		if n == "Level" {
			// 直接获取全部level的kv map
			levels, _ := SearchLevel(SearchLevelForm{
				Status: "valid",
			})
			result[n] = levels
		}
	}
	return result
}

func GetById(id uint) (*Price, error) {
	var o Price
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
func Update(o *Price) error {
	db := database.Conn

	if o.ID != 0 {
		if err := db.Model(&Price{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return err
		}
	}
	return nil
}

func Delete(id int) error {
	db := database.Conn
	if err := db.Delete(&Price{}, id).Error; err != nil {
		return err
	}
	return nil
}
