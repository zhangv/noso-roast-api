package page

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/category"
	"hahacking/utils"
	"strings"

	"gorm.io/gorm"
)

type Page struct {
	models.Model
	Code         string             `json:"code"`
	Title        string             `json:"title"`
	Subtitle     string             `json:"subtitle"`
	Summary      string             `json:"summary"` //tldr
	Author       string             `json:"author"`
	PublishDate  string             `json:"publish_date"`
	Type         PageType           `json:"type"`                     //page,post,recipe,cupping
	Content      string             `json:"content" gorm:"type:text"` //could be html
	Status       string             `json:"status"`
	CategoryID   *uint              `json:"category_id"`
	CategoryPath string             `json:"category_path"`
	Category     *category.Category `json:"category" gorm:"foreignKey:CategoryID;references:ID"`
}

type PageType string

const (
	PageTypePage     PageType = "page"
	PageTypeRedirect PageType = "redirect"
	PageTypeForward  PageType = "forward"
)

var PageTypes = map[PageType]string{
	PageTypePage:     "页面",
	PageTypeRedirect: "前端跳转",
	PageTypeForward:  "后端替换",
}

func (o *Page) AfterUpdate(tx *gorm.DB) (err error) {
	// o.CategoryPath = o.Category.Path()
	return nil
}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "PageStatus" {
			result[n] = models.ConvertMapToArray(models.CommonStatuses)
		}
		if n == "PageType" {
			result[n] = models.ConvertMapToArray(PageTypes)
		}
	}
	return result
}

func GetById(id uint) (*Page, error) {
	var o Page
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
func GetByCode(code string) (*Page, error) {
	var o Page
	db := database.Conn
	err := db.Where("code = ?", code).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchForm struct {
	utils.Pager
	Title      string `form:"title"`
	Code       string `form:"code"`
	CategoryID *uint  `form:"category_id"`
	Author     string `form:"author"`
	Status     string `form:"status"`
}

func Search(q SearchForm) ([]Page, int) {
	db := database.Conn
	var os []Page
	var statement = db.Model(&Page{}).Preload("Category")
	if q.Title != "" {
		statement = statement.Where("title LIKE ?", "%"+q.Title+"%")
	}
	if q.Code != "" {
		statement = statement.Where("code like ?", "%"+q.Code+"%")
	}
	if q.CategoryID != nil {
		statement = statement.Where("category_id = ?", *q.CategoryID)
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
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	Code         string `json:"code"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	PublishDate  string `json:"publish_date"`
	Content      string `json:"content"`
	Author       string `json:"author"`
	CategoryID   *uint  `json:"category_id"`
	CategoryPath string `json:"category_path"`
	utils.Pager
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Title == "" {
		result["title"] = "请填写标题"
	}
	if form.Code == "" {
		result["code"] = "请填写编码"
	} else {
		existingBean, err := GetByCode(form.Code) // Changed from form["code"] to form.Code
		if err == nil && existingBean.ID != 0 && existingBean.ID != form.ID {
			result["code"] = "编码已存在"
		}
	}

	return result
}

func Update(o *Page) (*Page, error) {
	db := database.Conn
	tx := db.Begin()

	if o.ID != 0 {
		if err := tx.Model(&Page{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
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

	if err := tx.Delete(&Page{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
