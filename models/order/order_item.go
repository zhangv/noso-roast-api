package order

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/bean"

	"gorm.io/gorm"
)

type OrderItem struct {
	models.Model
	OrderID     uint            `json:"order_id"`
	BeanID      uint            `json:"bean_id"`
	Bean        *bean.Bean      `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	BeanName    string          `json:"bean_name"`
	BeanBatchID uint            `json:"beanbatch_id" gorm:"null"`
	BeanBatch   *bean.BeanBatch `json:"beanbatch" gorm:"foreignKey:BeanBatchID;references:ID"`
	Qty         int             `json:"qty"`
	Price       float64         `json:"price" gorm:"type:decimal(10,2)"`
	Amt         float64         `json:"amt" gorm:"type:decimal(10,2)"`
}

func GetItemById(id uint) (*OrderItem, error) {
	var o OrderItem
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
