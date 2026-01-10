package supplier

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"
	"strings"

	"gorm.io/gorm"
)

type Supplier struct {
	ID          uint           `json:"id" form:"id" gorm:"primaryKey"`
	Name        string         `json:"name"`
	Code        string         `json:"code"`
	Type        SupplierType   `json:"type"`
	Country     string         `json:"country"`
	Status      SupplierStatus `json:"status"`
	StatusL     string         `json:"statusL" gorm:"-"`
	Address     string         `json:"address"`
	Contact     string         `json:"contact"`
	Tel         string         `json:"tel"`
	Description string         `json:"description"`
}

type SupplierStatus string

const (
	StatusValid   SupplierStatus = "valid"
	StatusInvalid SupplierStatus = "invalid"
)

type SupplierType string

const (
	SupplierTypeBean       SupplierType = "bean"       //咖啡豆
	SupplierTypeIngredient SupplierType = "ingredient" //食材
	SupplierTypeConsumable SupplierType = "consumable" //耗材
)

var SupplierTypes = map[SupplierType]string{
	SupplierTypeBean:       "咖啡豆",
	SupplierTypeIngredient: "食材",
	SupplierTypeConsumable: "耗材",
}

var SupplierStatuses = models.CommonStatuses

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "SupplierType" {
			result[n] = models.ConvertMapToArray(SupplierTypes)
		}
		if n == "SupplierStatus" {
			result[n] = models.ConvertMapToArray(SupplierStatuses)
		}
	}
	return result
}

func GetById(id uint) (*Supplier, error) {
	var o Supplier
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
func GetByCode(code string) (*Supplier, error) {
	var o Supplier
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
}

func Search(q SearchForm) ([]Supplier, int) {
	db := database.Conn
	var os []Supplier

	var statement = db.Model(&Supplier{})
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

func Update(o *Supplier) error {
	// db := GetConn()
	db := database.Conn
	//TODO 重构为form并validate
	if o.ID != 0 {
		if err := db.Model(&Supplier{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return err
		}
	}
	return nil
}

func (supplier Supplier) Delete() error {
	db := database.Conn
	if err := db.Delete(&Supplier{}, supplier.ID).Error; err != nil {
		return err
	}
	return nil
}
func (supplier Supplier) Remove() error {
	db := database.Conn
	if err := db.Model(&Supplier{}).Where("id = ?", supplier.ID).Updates(map[string]any{"status": "invalid"}).Error; err != nil {
		return err
	}
	return nil
}
