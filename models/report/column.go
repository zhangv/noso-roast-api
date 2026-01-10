package report

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"

	"gorm.io/gorm"
)

type Column struct {
	models.Model
	Name        string        `json:"name"`
	DisplayName string        `json:"displayname" gorm:"column:displayname"`
	Sort        string        `json:"sort" gorm:"column:sort"`
	Status      models.Status `json:"status" gorm:"column:status"`
	Description string        `json:"description" gorm:"column:description"`
	ReportID    uint          `json:"report_id"`
	Report      Report        `json:"report" gorm:"foreignKey:ReportID;references:ID"`
	Link        string        `json:"link"`
}

func (r *Column) TableName() string {
	return "t_reportcolumn"
}
func (r *Column) ValidateForm() map[string]string {
	errs := make(map[string]string)
	if r.Name == "" {
		errs["name"] = "名称不能为空"
	}
	return errs
}

func GetColumnById(id uint) (*Column, error) {
	var o Column
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchColumnForm struct {
	utils.Pager
	Name   string `form:"name"`
	Status string `form:"status"`
}

func SearchColumn(q SearchColumnForm) ([]Column, int) {
	db := database.Conn
	var os []Column
	var statement = db.Model(&Column{})
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
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
	return os, int(count)
}

func UpdateColumn(o *Column) error {
	db := database.Conn
	tx := db.Begin()

	if o.ID != 0 {
		if err := tx.Model(&Column{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
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

func StatusColumn(id int, to string) error {
	db := database.GetConn()
	tx := db.Begin()

	if err := tx.Model(&Column{}).Where("id = ?", id).Update("status", to).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func DeleteColumn(id int) error {
	db := database.GetConn()
	tx := db.Begin()

	if err := tx.Delete(&Column{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
