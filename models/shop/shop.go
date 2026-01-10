package shop

import (
	"hahacking/database"
	"hahacking/models"
	"strings"

	"gorm.io/gorm"
)

type Shop struct {
	// models.LegacyModel
	models.Model
	Name          string `json:"name"`
	Type          string `json:"type"`
	Contact       string `json:"contact"`
	Tel           string `json:"tel"`
	Pics          string `json:"pics"`
	Status        string `json:"status" `
	DeliveryType  string `json:"deliverytype" `
	DeliveryTypeL string `json:"deliverytypeL" gorm:"-"`
	Address       string `json:"address"`
	Description   string `json:"description"`
	Owner         string `json:"owner"`
}

func (s *Shop) AfterFind(tx *gorm.DB) (err error) {
	s.DeliveryTypeL = translateDeliveryType(s.DeliveryType)
	return
}

func translateDeliveryType(deliveryType string) string {
	switch deliveryType {
	case "express":
		return "快递"
	case "pickup":
		return "自提"
	default:
		return "未知"
	}
}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "ShopStatus" {
			result[n] = models.ConvertMapToArray(models.CommonStatuses)
		}
		if n == "ShopAccountType" {
			result[n] = models.ConvertMapToArray(ShopAccountTypes)
		}
		if n == "ShopAccountStatus" {
			result[n] = models.ConvertMapToArray(models.CommonStatuses)
		}
	}
	return result
}
func GetById(id int) (*Shop, error) {
	var o Shop
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func Search(q map[string]interface{}) ([]Shop, int) {
	db := database.Conn
	var os []Shop
	var pageno int
	if val, ok := q["pageNo"]; ok {
		pageno = val.(int)

	} else {
		pageno = 1 // default value
	}
	delete(q, "pageNo")
	var pagesize int
	if val, ok := q["pageSize"]; ok {
		pagesize = val.(int)
	} else {
		pagesize = 10 // default value
	}
	delete(q, "pageSize")
	start := (pageno - 1) * pagesize
	// var statement = db.Table("t_shop") //.Model(&Shop{})
	var statement = db.Model(&Shop{})
	if name, ok := q["name"]; ok {
		statement = statement.Where("name LIKE ?", "%"+name.(string)+"%")
		delete(q, "name")
	}
	var count int64
	statement.Where(q).
		Count(&count).
		Order("id").
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

func Update(o *Shop) error {
	db := database.Conn

	if o.ID != 0 {
		if err := db.Model(&Shop{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
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
	err := db.Delete(&Shop{}, id).Error
	return err
}
