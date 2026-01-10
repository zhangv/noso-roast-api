package tag

import (
	"hahacking/database"

	"gorm.io/gorm"
)

type SearchGroupForm struct {
	Names []string `json:"names"`
}

func SearchGroup(q SearchGroupForm) []TagGroup {
	db := database.Conn
	var os []TagGroup

	var statement = db.Model(&TagGroup{}).Preload("Tags")
	statement.Where("name in ?", q.Names)
	statement.Find(&os)

	return os
}

func AddGroup(name string) (TagGroup, error) {
	db := database.Conn
	var tg TagGroup
	tg.Name = name
	err := db.Create(&tg).Error
	if err != nil {
		return tg, err
	}
	return tg, nil
}

//	func AddTagToGroup(tagID uint, groupID uint) error {
//		db := database.GetConn()
//		err := db.Create(&Tagging{TagID: tagID, TagGroupID: groupID}).Error
//		if err != nil {
//			return err
//		}
//		return nil
//	}
func (tg *TagGroup) Delete() error {
	db := database.Conn
	if err := db.Delete(&TagGroup{}, tg.ID).Error; err != nil {
		return err
	}
	return nil
}

func GetGroupById(id uint) (*TagGroup, error) {
	var o TagGroup
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}
