package acl

import (
	"hahacking/database/datatypes"
	"hahacking/models"
)

type Role struct {
	models.LegacyModel
	Name        string `json:"name"`
	Code        string `gorm:"type:varchar(200)" json:"code"`
	Status      string `json:"status"`
	Description string `json:"description"`
	//domain在这里用来区分角色可分配的范围，如shop，则只有在为门店分配角色时才使用
	// Domain string `gorm:"type:varchar(100)" json:"domain"`
	Domain  datatypes.Slot `gorm:"type:text" json:"domain"`
	GroupId int            `json:"int"`
}
type Group struct { //组织机构层级
	models.LegacyModel
	Name        string `json:"name"`
	Code        string `json:"code"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	ParentId    int    `json:"int"` //所属组
}
