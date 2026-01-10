package shop

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"
)

type ShopAccountLog struct {
	models.Model
	ShopAccountID uint         `json:"shopaccount_id" gorm:"column:shopaccount_id"`
	ShopAccount   *ShopAccount `json:"shopaccount" gorm:"foreignKey:ShopAccountID;references:ID"`
	Amount        float64      `json:"amount" gorm:"type:decimal(10,2)"` //可以为负值
	From          float64      `json:"from" gorm:"type:decimal(10,2)"`
	To            float64      `json:"to" gorm:"type:decimal(10,2)"`
	Remark        string       `json:"remark"`
	SrcType       string       `json:"src_type"`
	SrcID         string       `json:"src_id"`
	SN            string       `json:"sn"`
	Status        string       `json:"status" `
}
type SearchAccountLogForm struct {
	utils.Pager
	ShopAccountID uint   `form:"shopaccount_id"`
	SrcType       string `form:"src_type"`
	SrcID         string `form:"src_id"`
	Status        string `form:"status"`
	Preloads      string `form:"preloads"`
}

func SearchAccountLog(q SearchAccountLogForm) ([]ShopAccountLog, int) {
	db := database.Conn
	var os []ShopAccountLog

	var statement = db.Model(&ShopAccountLog{})
	if q.Preloads != "" {
		statement = statement.Preload(q.Preloads)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	if q.ShopAccountID != 0 {
		statement = statement.Where("shopaccount_id = ?", q.ShopAccountID)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	var count int64
	statement.Where(&ShopAccountLog{}).
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

func GetAccountLogById(id int) (*ShopAccountLog, error) {
	db := database.Conn
	var o ShopAccountLog
	db.Where("id = ?", id).First(&o)
	return &o, nil
}
