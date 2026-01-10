package user

import (
	"crypto/sha256"
	"encoding/hex"
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/acl"
	"hahacking/models/price"
	"hahacking/utils"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	models.Model
	Nickname string  `json:"nickname"`
	Email    *string `json:"email" gorm:"uniqueIndex"`
	Mobile   *string `json:"mobile" gorm:"uniqueIndex;type:varchar(20);"`
	//前端输入+sha256后的串 + sha256+salt的串
	Password     string            `json:"-"`
	Type         string            `json:"type"`
	Status       string            `json:"status"`
	RegisterTime string            `gorm:"column:registertime" json:"registertime"`
	PriceLevel   *price.PriceLevel `json:"pricelevel" gorm:"foreignKey:PriceLevelID;references:ID"`
	PriceLevelID *uint             `json:"pricelevel_id" gorm:"column:pricelevel_id"`
}
type UserStatus string

const (
	UserStatusValid   UserStatus = "valid"
	UserStatusInvalid UserStatus = "invalid"
)

var UserStatuses = map[UserStatus]string{
	UserStatusValid:   "正常",
	UserStatusInvalid: "禁用",
}

type UserType string

const (
	UserTypeAdmin UserType = "admin"
	UserTypeUser  UserType = "user"
)

var UserTypes = map[UserType]string{
	UserTypeAdmin: "管理员",
	UserTypeUser:  "用户",
}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "UserStatus" {
			result[n] = models.ConvertMapToArray(UserStatuses)
		}
		if n == "UserType" {
			result[n] = models.ConvertMapToArray(UserTypes)
		}
	}
	return result
}

func GetById(id uint) (*User, error) {
	var user User
	db := database.GetConn()
	err := db.Select("*").Where("id = ?", id).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if user.ID > 0 {
		return &user, nil
	}

	return nil, nil
}

type SearchForm struct {
	utils.Pager
	NickName string `form:"nickname"`
	Email    string `form:"email"`
	Mobile   string `form:"mobile"`
	Status   string `form:"status"`
}

func Search(q SearchForm) ([]User, int) {
	db := database.Conn
	var os []User

	var statement = db.Model(&User{})
	if q.NickName != "" {
		statement = statement.Where("nickname LIKE ?", "%"+q.NickName+"%")
	}
	if q.Email != "" {
		statement = statement.Where("email LIKE ?", "%"+q.Email+"%")
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	} else {
		statement = statement.Where("status != ?", "invalid")
	}
	if q.Mobile != "" {
		statement = statement.Where("mobile LIKE ?", "%"+q.Mobile+"%")
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

func Update(o *User) error {
	db := database.Conn
	if o.ID != 0 {
		if err := db.Model(&User{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return err
		}
	}
	return nil
}

func GetByNickname(nick string) (*User, error) {
	var user User
	db := database.GetConn()
	err := db.Select("*").Where("nickname = ?", nick).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if user.ID > 0 {
		return &user, nil
	}

	return nil, nil
}
func GetByMobile(mobile string) (*User, error) {
	var user User
	db := database.GetConn()
	err := db.Select("*").Where("mobile = ?", mobile).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if user.ID > 0 {
		return &user, nil
	}

	return nil, nil
}

func GetByEmail(email string) (*User, error) {
	var user User
	db := database.GetConn()
	err := db.Select("*").Where("email = ?", email).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if user.ID > 0 {
		return &user, nil
	}

	return nil, nil
}
func (u *User) GetRoles() []*acl.Role {
	db := database.GetConn()
	var roles []*acl.Role
	db.Model(&u).Association("Roles").Find(&roles)
	return roles
}

/*
前端传入的是明文+sha256的串
数据库存储的是明文+sha256+salt+sha256的串
*/
func CreatePassword(pass string) string {
	var salt = "hahacking"
	sum := sha256.Sum256([]byte(pass + salt))
	// 将字节数组转换为十六进制字符串
	return hex.EncodeToString(sum[:])
}

func (user User) Delete() error {
	db := database.Conn
	if err := db.Delete(&User{}, user.ID).Error; err != nil {
		return err
	}
	return nil
}
func (user User) Remove() error {
	db := database.Conn
	if err := db.Model(&User{}).Where("id = ?", user.ID).Updates(map[string]any{"status": "invalid"}).Error; err != nil {
		return err
	}
	return nil
}
