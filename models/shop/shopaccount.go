package shop

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"

	"gorm.io/gorm"
)

type ShopAccount struct {
	models.Model
	Name    string          `json:"name"`
	Type    ShopAccountType `json:"type"`
	Code    string          `json:"code"`
	Balance float64         `json:"balance" gorm:"type:decimal(10,2)"`
	Status  models.Status   `json:"status" `
	ShopID  uint            `json:"shop_id" `
}
type ShopAccountType string

const (
	ShopAccountTypeCash = "cash"
	ShopAccountTypeAR   = "ar" //应收
	ShopAccountTypeAP   = "ap" //应付
)

var ShopAccountTypes = map[ShopAccountType]string{
	ShopAccountTypeCash: "现金",
	ShopAccountTypeAR:   "应收",
	ShopAccountTypeAP:   "应付",
}

func (account *ShopAccount) Balancing(amount float64, srctype string, srcid string) (float64, error) {
	db := database.Conn
	log := ShopAccountLog{
		ShopAccountID: account.ID,
		Amount:        amount,
		From:          account.Balance,
		To:            account.Balance + amount,
		SrcType:       srctype,
		SrcID:         srcid,
	}
	tx := db.Begin()
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	// Update the balance
	newBalance := account.Balance + amount
	if err := tx.Model(&ShopAccount{}).Where("id = ?", account.ID).Update("balance", newBalance).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	tx.Commit()
	account.Balance = newBalance
	return newBalance, nil
}
func GetAccountById(id int) (*ShopAccount, error) {
	var o ShopAccount
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchAccountForm struct {
	utils.Pager
	Name     string   `form:"name"`
	NameLike string   `form:"name_like"`
	Code     string   `form:"code"`
	CodeLike string   `form:"code_like"`
	Status   string   `form:"status"`
	Statuses []string `form:"statuses"`
	Preloads string   `form:"preloads"`
}

func SearchAccount(q SearchAccountForm) ([]ShopAccount, int) {
	db := database.Conn
	var os []ShopAccount

	var statement = db.Model(&ShopAccount{})
	if q.Name != "" {
		statement = statement.Where("name = ?", q.Name)
	}
	if q.NameLike != "" {
		statement = statement.Where("name LIKE ?", "%"+q.NameLike+"%")
	}
	if q.Code != "" {
		statement = statement.Where("code = ?", q.Code)
	}
	if q.CodeLike != "" {
		statement = statement.Where("code LIKE ?", "%"+q.CodeLike+"%")
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	if len(q.Statuses) > 0 {
		statement = statement.Where("status IN ?", q.Statuses)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	var count int64
	statement.Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

func UpdateAccount(o *ShopAccount) (*ShopAccount, error) {
	db := database.Conn
	if o.ID != 0 {
		if err := db.Model(&ShopAccount{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return nil, err
		}
	}
	return o, nil
}
func DeleteAccount(id int) error {
	db := database.Conn
	if err := db.Delete(&ShopAccount{}, id).Error; err != nil {
		return err
	}
	return nil
}
