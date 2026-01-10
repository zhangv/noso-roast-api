package order

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"

	"gorm.io/gorm"
)

type OrderPayment struct {
	models.Model
	OrderID uint               `json:"order_id"`
	Status  models.Status      `json:"status" `
	Items   []OrderPaymentItem `json:"items" gorm:"foreignKey:OrderPaymentID;references:ID"`
}
type OrderPaymentItem struct {
	ID uint `json:"id"`
	models.Model
	OrderID        uint             `json:"order_id"`
	OrderPaymentID uint             `json:"orderpayment_id" gorm:"column:orderpayment_id"`
	Type           OrderPaymentType `json:"type"`
	Amt            float64          `json:"amt"`
	Qty            float64          `json:"qty"`
	Memo           string           `json:"memo"`
	Ids            string           `json:"ids"`
	Status         models.Status    `json:"status" `
}

type OrderPaymentType string

const (
	OrderPaymentTypeCash        OrderPaymentType = "cash"        // 现金
	OrderPaymentTypeAlipay      OrderPaymentType = "alipay"      // 支付宝
	OrderPaymentTypeWeChat      OrderPaymentType = "wechat"      // 微信
	OrderPaymentTypeBank        OrderPaymentType = "bank"        // 银行转账
	OrderPaymentTypeShopAccount OrderPaymentType = "shopaccount" // 挂账
	OrderPaymentTypeCredit      OrderPaymentType = "credit"      // 积分
	OrderPaymentTypeGiftCard    OrderPaymentType = "giftcard"
	OrderPaymentTypeOther       OrderPaymentType = "other" // 其他
)

func GetPaymentById(id uint) (*OrderPayment, error) {
	var o OrderPayment
	db := database.Conn
	err := db.Where("id = ?", id).Preload("Items").First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchPaymentForm struct {
	utils.Pager
	OrderID uint          `form:"order_id"`
	Status  models.Status `form:"status"`
}

func SearchPayment(q SearchPaymentForm) ([]OrderPayment, int) {
	db := database.Conn
	var os []OrderPayment
	var count int64
	var statement = db.Model(&OrderPayment{})
	if q.OrderID != 0 {
		statement = statement.Where("order_id = ?", q.OrderID)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	statement.Where(&OrderPayment{}).Preload("Items").
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

func (p *OrderPayment) GetItems() ([]OrderPaymentItem, error) {
	var items []OrderPaymentItem
	db := database.Conn
	err := db.Where("orderpayment_id = ?", p.ID).Find(&items).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return items, nil
}

func GetPaymentItemById(id uint) (*OrderPaymentItem, error) {
	var item OrderPaymentItem
	db := database.Conn
	err := db.Where("id = ?", id).First(&item).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &item, nil
}
