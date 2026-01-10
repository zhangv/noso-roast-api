package favorite

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/category"
	"hahacking/models/item"
	"hahacking/models/report"
	"hahacking/models/user"
	"hahacking/utils"
	"strconv"

	"gorm.io/gorm"
)

/*
收藏内容
*/
type Favorite struct {
	models.Model
	ItemID string        `json:"item_id" gorm:"column:item_id"` // 目标ID
	Type   string        `json:"type"`                          // 类型
	Status models.Status `json:"status"`                        // 状态
	UserID *uint         `json:"user_id" gorm:"not null"`
	User   *user.User    `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

func (o *Favorite) Target() string {
	return "todo"
}

func UserAdd(typ string, itemID string, userID uint) Favorite {
	db := database.Conn
	f := Favorite{
		ItemID: itemID,
		Type:   typ,
		Status: models.StatusValid,
		UserID: &userID,
	}
	db.FirstOrCreate(&f)
	return f
}
func UserRemove(typ string, itemID string, userID uint) error {
	db := database.Conn
	err := db.Delete(&Favorite{}, "type = ? AND item_id = ? AND user_id = ?", typ, itemID, userID).Error
	if err != nil {
		return err
	}
	return nil
}

func GetById(id uint) (*Favorite, error) {
	var o Favorite
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchForm struct {
	utils.Pager
	ItemID string `form:"item_id"`
	UserID string `form:"user_id"`
	Type   string `form:"type"`
	Status string `form:"status"`
}

func Search(q SearchForm) ([]interface{}, int) {
	db := database.Conn
	var os []Favorite
	var statement = db.Model(&Favorite{}).Preload("User")
	if q.ItemID != "" {
		statement = statement.Where("item_id = ?", q.ItemID)
	}
	if q.UserID != "" {
		statement = statement.Where("user_id = ?", q.UserID)
	}
	if q.Type != "" {
		statement = statement.Where("type = ?", q.Type)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	} else {
		statement = statement.Where("status = ?", models.StatusValid)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	var count int64
	statement.
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)

	objs := make([]interface{}, len(os))
	for i, o := range os {
		switch o.Type {
		case "Report", "report":
			id, _ := strconv.ParseUint(o.ItemID, 10, 64)
			objs[i], _ = report.GetById(uint(id))

		case "Category", "category":
			id, _ := strconv.ParseUint(o.ItemID, 10, 64)
			objs[i], _ = category.GetById(uint(id))

		case "Item", "item":
			id, _ := strconv.ParseUint(o.ItemID, 10, 64)
			objs[i], _ = item.GetById(uint(id))

		default:
			objs[i] = o
		}
	}
	return objs, int(count)
}

func Update(o *Favorite) error {
	db := database.Conn
	tx := db.Begin()

	if o.ID != 0 {
		if err := tx.Model(&Favorite{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func Delete(id int) error {
	db := database.Conn
	tx := db.Begin()
	if err := tx.Delete(&Favorite{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
