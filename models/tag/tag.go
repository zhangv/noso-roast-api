package tag

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/utils"
)

type TagGroup struct {
	models.Model
	Name string `json:"name" gorm:"uniqueIndex;not null"` // 如 "产地"、"风味"、"使用方式"
	Tags []Tag  `json:"tags" gorm:"foreignKey:GroupID"`
}

type Tag struct {
	models.Model
	Name      string    `json:"name" gorm:"not null"`
	GroupID   *uint     `json:"group_id" gorm:"index;null"`
	Group     *TagGroup `json:"group" gorm:"foreignKey:GroupID;references:ID"`
	GroupName string    `json:"group_name" gorm:"-"`
}

type Tagging struct {
	models.Model
	TagID      uint   `json:"tag_id" gorm:"index;not null"`
	TargetID   string `json:"target_id" gorm:"index;not null"`           // 被打标签的对象 ID
	TargetType string `json:"target_type" gorm:"index;size:50;not null"` // 被打标签的模型名，比如 "beans"
}

type AddForm struct {
	Name       string `json:"name"`
	GroupID    uint   `json:"group_id"`
	GroupName  string `json:"group_name"` //如果提供标签组名称则创建
	Status     string `json:"status"`
	TargetID   string `json:"target_id"`
	TargetType string `json:"target_type"`
}

func (form *AddForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写标签"
	}

	return result
}
func Add(form AddForm) (Tag, error) {
	db := database.Conn
	if form.GroupName != "" {
		grp := TagGroup{Name: form.GroupName}
		db.FirstOrCreate(&grp, TagGroup{Name: form.GroupName})
		form.GroupID = grp.ID
	}
	tag := Tag{Name: form.Name, GroupID: &form.GroupID}
	existingTag := Tag{}
	db.Where("name = ? and group_id = ?", form.Name, form.GroupID).First(&existingTag)
	if existingTag.ID > 0 {
		tag.ID = existingTag.ID
	} else {
		db.Create(&tag)
	}
	tagging := Tagging{
		TagID:      tag.ID,
		TargetID:   form.TargetID,
		TargetType: form.TargetType,
	}
	db.FirstOrCreate(&tagging, tagging)
	return tag, nil
}

type SearchForm struct {
	utils.Pager
	Name    string `form:"name"`
	GroupID uint   `form:"group_id"`
}

func Search(q SearchForm) ([]Tag, int) {
	db := database.Conn
	var os []Tag

	var statement = db.Model(&Tag{}).Select("tags.*, tag_groups.name as group_name")
	statement = statement.Joins("LEFT JOIN tag_groups ON tags.group_id = tag_groups.id")
	if q.Name != "" {
		statement = statement.Where("tags.name LIKE ?", "%"+q.Name+"%")
	}
	if q.GroupID > 0 {
		statement = statement.Where("tags.group_id = ?", q.GroupID)
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

func Delete(id int) error {
	db := database.Conn
	if err := db.Delete(&Tag{}, id).Error; err != nil {
		return err
	}
	return nil
}

// func GetById(id uint) (*Tag, error) {
// 	var o Tag
// 	db := database.Conn
// 	err := db.Where("id = ?", id).First(&o).Error
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return nil, err
// 	}
// 	return &o, nil
// }

func GetByTarget(targetID string, targetType string) ([]Tag, error) {
	db := database.Conn
	var os []Tag
	statement := db.Model(&Tag{}).Preload("Group").Select("distinct t_tag.*")
	statement = statement.Joins("JOIN t_tagging ON t_tagging.tag_id = t_tag.id")
	// statement = statement.Joins("JOIN t_tag_group ON t_tag.group_id = t_tag_group.id")
	if targetID != "" {
		statement.Where("t_tagging.target_id = ? ", targetID)
	}
	statement.Where("t_tagging.target_type = ?", targetType)
	statement.Where("t_tagging.deleted_at IS NULL")
	err := statement.Find(&os).Error
	return os, err
}

func GetGroupByTarget(targetid string, targettype string) ([]TagGroup, error) {
	db := database.Conn
	var os []TagGroup
	statement := db.Model(&TagGroup{}).Preload("Tags").Select("distinct t_tag_group.*") //Distinct("t_tag_group.name")
	statement = statement.Joins("JOIN t_tag ON t_tag_group.id = t_tag.group_id")
	statement = statement.Joins("JOIN t_tagging ON t_tag.id = t_tagging.tag_id")
	if targetid != "" {
		statement.Where("t_tagging.target_id = ? ", targetid)
	}
	statement.Where("t_tagging.target_type = ?", targettype)
	statement.Find(&os)
	return os, nil
}
func AddTag(tagname string, groupname string, targetid string, targettype string) (Tag, error) {
	exgroup := SearchGroup(SearchGroupForm{Names: []string{groupname}})
	var group TagGroup
	if len(exgroup) == 0 {
		group, _ = AddGroup(groupname)
	} else {
		group = exgroup[0]
	}
	tag, _ := Add(AddForm{Name: tagname, GroupID: group.ID, TargetID: targetid, TargetType: targettype})
	return tag, nil
}

func RemoveTag(tagid string, targetid string, targettype string) error {
	db := database.Conn
	return db.Delete(&Tagging{}, "tag_id = ? and target_id = ? and target_type = ?", tagid, targetid, targettype).Error
}
