package roast

import (
	"hahacking/database"
	"hahacking/models"

	"gorm.io/gorm"
)

type RoastProfile struct {
	models.Model
	UserID     uint    `json:"user_id" gorm:"column:user_id"`
	RoastID    uint    `json:"roast_id" gorm:"column:roast_id"`
	BeanID     uint    `json:"bean_id" gorm:"column:bean_id"`
	Status     string  `json:"status"`
	StatusL    string  `json:"statusL" gorm:"-"`
	Moisture   float64 `json:"moisture"`
	StartTemp  float64 `json:"starttemp" gorm:"column:starttemp"`
	TpTemp     float64 `json:"tptemp" gorm:"column:tptemp"`
	TpTime     uint    `json:"tptime" gorm:"column:tptime"`
	YellowTemp float64 `json:"yellowtemp" gorm:"column:yellowtemp"`
	YellowTime uint    `json:"yellowtime" gorm:"column:yellowtime"`
	FcTemp     float64 `json:"fctemp" gorm:"column:fctemp"`
	FcTime     uint    `json:"fctime" gorm:"column:fctime"`
	ScTemp     float64 `json:"sctemp" gorm:"column:sctemp"`
	ScTime     uint    `json:"sctime" gorm:"column:sctime"`
	TotalTime  uint    `json:"totaltime" gorm:"column:totaltime"`
	Memo       string  `json:"memo" gorm:"column:memo"`
	//成品参数 - 完成时更新
	Agtron       int         `json:"agtron" gorm:"column:agtron"`
	AgtronGround int         `json:"agtron_ground" gorm:"column:agtron_ground"`
	Degree       RoastDegree `json:"degree"` //烘焙度
	LossRate     float64     `json:"lossrate" gorm:"column:lossrate"`
}

type RoastDegree string

const (
	RoastDegreeLight      RoastDegree = "light"      //Cinnamon Roast, New England Roast
	RoastDegreeMedium     RoastDegree = "medium"     //City Roast, American Roast
	RoastDegreeMediumDark RoastDegree = "mediumdark" //Full City Roast, Vienna Roast
	RoastDegreeDark       RoastDegree = "dark"       //Italian Roast, French Roast
	RoastDegreeExtraDark  RoastDegree = "extradark"  //Spanish Roast
)

var RoastDegrees = map[RoastDegree]string{
	RoastDegreeLight:      "浅烘",
	RoastDegreeMedium:     "中烘",
	RoastDegreeMediumDark: "中深烘",
	RoastDegreeDark:       "深烘",
	RoastDegreeExtraDark:  "极深烘",
}

func GetProfileById(id int) (*RoastProfile, error) {
	var o RoastProfile
	db := database.GetConn()
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func GetProfileByRoastId(id uint) (*RoastProfile, error) {
	var o RoastProfile
	db := database.GetConn()
	err := db.Where("roast_id = ?", id).First(&o).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			roast, _ := GetById(id)
			o.RoastID = id
			o.BeanID = roast.BeanID
			if err = db.Create(&o).Error; err != nil {
				return nil, err
			} else {
				return &o, nil
			}
		}
	} else {
		return &o, nil
	}
	return &o, nil
}

func UpdateProfile(o *RoastProfile) error {
	db := database.Conn

	if o.ID != 0 {
		if err := db.Model(&RoastProfile{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return err
		}
	}
	return nil
}
