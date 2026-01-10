package category

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"
	"strings"

	"gorm.io/gorm"
)

type Category struct {
	models.Model
	Name     string     `json:"name"`
	Code     string     `json:"code"`
	Type     string     `json:"type"` //page,post,recipe,cupping
	Status   string     `json:"status"`
	ParentID *uint      `json:"parent_id" gorm:"null"`
	Parent   *Category  `json:"parent" gorm:"foreignKey:ParentID;references:ID"`
	Children []Category `json:"children" gorm:"-"`
}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "CategoryStatus" {
			result[n] = models.ConvertMapToArray(models.CommonStatuses)
		}
	}
	return result
}

func GetById(id uint) (*Category, error) {
	var o Category
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func (c *Category) Tree(depth int) ([]Category, error) {
	db := database.Conn
	var children []Category
	err := db.Where("parent_id = ?", c.ID).Find(&children).Error
	if err != nil {
		return nil, err
	}
	// 递归
	if depth != 1 {
		for i := range children {
			subChildren, err := children[i].Tree(depth - 1)
			if err != nil {
				return nil, err
			}
			children[i].Children = subChildren
		}
	}
	return children, nil
}

type SearchForm struct {
	utils.Pager
	Name   string `form:"name"`
	BeanID string `form:"beanid"`
	Status string `form:"status"`
}

func Search(q SearchForm) ([]Category, int) {
	db := database.Conn
	var os []Category
	var statement = db.Model(&Category{}).Preload("Parent")
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
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

type UpdateForm struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	PublishDate string `json:"publish_date"`
	Content     string `json:"content"`
	Author      string `json:"author"`
	ParentID    *uint  `json:"parent_id"`
	utils.Pager
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写名称"
	}
	if form.Code == "" {
		result["code"] = "请填写编码"
	}
	if form.ParentID == nil {
		result["parent_id"] = "请选择父分类"
	}
	if form.Status == "" {
		result["status"] = "请选择状态"
	}
	return result
}
func Update(o *Category) (*Category, error) {
	db := database.Conn
	tx := db.Begin()
	if o.ID != 0 {
		if err := tx.Model(&Category{}).Where("id = ?", o.ID).Updates(map[string]any{
			"name":      o.Name,
			"code":      o.Code,
			"type":      o.Type,
			"status":    o.Status,
			"parent_id": o.ParentID,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()
	return o, nil
}

func Delete(id int) error {
	db := database.GetConn()
	tx := db.Begin()

	if err := tx.Delete(&Category{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
