package price

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"

	"gorm.io/gorm"
)

type PriceLevel struct {
	ID uint `json:"id" form:"id" gorm:"primaryKey"`
	models.Model
	Name        string    `json:"name"`
	Type        LevelType `json:"type"`
	Discount    float64   `json:"discount" gorm:"column:discount"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
}
type LevelType string

const (
	LevelTypeB LevelType = "B" // 批发价
	LevelTypeC LevelType = "C" // 零售价
)

var LevelTypes = map[LevelType]string{
	LevelTypeB: "批发价",
	LevelTypeC: "零售价",
}

func GetLevelById(id uint) (*PriceLevel, error) {
	var o PriceLevel
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
func GetLevelByCode(code string) (*PriceLevel, error) {
	var o PriceLevel
	db := database.Conn
	err := db.Where("code = ?", code).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchLevelForm struct {
	utils.Pager
	Name   string `form:"name"`
	Code   string `form:"code"`
	Type   string `form:"type"`
	Status string `form:"status"`
}

func SearchLevel(q SearchLevelForm) ([]PriceLevel, int) { // Update return type to []PriceLevel
	db := database.Conn
	var os []PriceLevel                     // Update to []PriceLevel
	var statement = db.Model(&PriceLevel{}) // Update model to PriceLevel
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Type != "" {
		statement = statement.Where("type = ?", q.Type)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	var count int64
	statement.Where(&PriceLevel{}).
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

func UpdateLevel(o *PriceLevel) error {
	db := database.Conn

	if o.ID != 0 {
		if err := db.Model(&PriceLevel{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return err
		}
	}
	return nil
}

func DeleteLevel(id int) error {
	db := database.Conn
	if err := db.Delete(&PriceLevel{}, id).Error; err != nil {
		return err
	}
	return nil
}
