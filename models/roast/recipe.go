package roast

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/bean"
	"hahacking/utils"

	"gorm.io/gorm"
)

type RoastRecipe struct {
	models.Model
	BeanID         uint          `json:"bean_id" gorm:"column:bean_id"`
	Bean           bean.Bean     `json:"bean" gorm:"foreignKey:BeanID"`
	RoastProfileID uint          `json:"roastprofile_id" gorm:"column:roastprofile_id"`
	Status         models.Status `json:"status"`
	StatusL        string        `json:"statusL" gorm:"-"`
	StartTemp      float64       `json:"starttemp" gorm:"column:starttemp"`
	TpTemp         float64       `json:"tptemp" gorm:"column:tptemp"`
	TpTime         uint          `json:"tptime" gorm:"column:tptime"`
	YellowTemp     float64       `json:"yellowtemp" gorm:"column:yellowtemp"`
	YellowTime     uint          `json:"yellowtime" gorm:"column:yellowtime"`
	FcTemp         float64       `json:"fctemp" gorm:"column:fctemp"`
	FcTime         uint          `json:"fctime" gorm:"column:fctime"`
	ScTemp         float64       `json:"sctemp" gorm:"column:sctemp"`
	ScTime         uint          `json:"sctime" gorm:"column:sctime"`
	TotalTime      uint          `json:"totaltime" gorm:"column:totaltime"`
	Description    string        `json:"description" gorm:"column:description;type:text"`
	//以下为同款豆的区分参数（即，不同的烘焙机、烘焙师、入豆量、烘焙度可能不同）
	BeanQty      int         `json:"beanqty" gorm:"column:beanqty"`             //入豆量
	Degree       RoastDegree `json:"degree" gorm:"column:degree"`               //烘焙度
	Machine      string      `json:"machine" gorm:"column:machine"`             //烘焙机
	Roaster      string      `json:"roaster" gorm:"column:roaster"`             //烘焙师
	Agtron       int         `json:"agtron" gorm:"column:agtron"`               //豆色值
	AgtronGround int         `json:"agtron_ground" gorm:"column:agtron_ground"` //粉色值
	LossRate     float64     `json:"lossrate" gorm:"column:lossrate"`
}

func (o RoastRecipe) IsAuditable() {}

func GetRecipeById(id int) (*RoastRecipe, error) {
	var o RoastRecipe
	db := database.GetConn()
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func SaveRecipeByRoastId(roastId uint) (rec *RoastRecipe, err error) {
	roast, err := GetById(roastId)
	if err != nil {
		return nil, err
	}
	profile, _ := GetProfileByRoastId(roast.ID)
	// exrecipe, _ := SearchRecipe(SearchRecipeForm{
	// 	BeanID:  strconv.Itoa(int(profile.BeanID)),
	// 	Degree:  string(profile.Degree),
	// 	BeanQty: float64(roast.InQty),
	// 	Status:  string(models.StatusValid),
	// })
	// if len(exrecipe) > 0 {
	// 	err = errors.New("豆（" + exrecipe[0].Bean.Name + "）已存在(烘焙度" + string(exrecipe[0].Degree) + "，入豆量为" + strconv.Itoa(int(exrecipe[0].BeanQty)) + "克)的配方")
	// 	return nil, err
	// }
	var recipe RoastRecipe

	recipe.BeanID = profile.BeanID
	recipe.Degree = profile.Degree
	recipe.BeanQty = int(roast.InQty)
	recipe.Status = models.StatusValid
	recipe.RoastProfileID = profile.ID

	recipe.Agtron = profile.Agtron
	recipe.AgtronGround = profile.AgtronGround
	recipe.LossRate = profile.LossRate
	recipe.Description = profile.Memo
	recipe.StartTemp = profile.StartTemp
	recipe.TpTemp = profile.TpTemp
	recipe.TpTime = profile.TpTime
	recipe.YellowTemp = profile.YellowTemp
	recipe.YellowTime = profile.YellowTime
	recipe.FcTemp = profile.FcTemp
	recipe.FcTime = profile.FcTime
	recipe.ScTemp = profile.ScTemp
	recipe.ScTime = profile.ScTime
	recipe.TotalTime = profile.TotalTime
	err = UpdateRecipe(&recipe)
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}
func UpdateRecipe(o *RoastRecipe) error {
	db := database.Conn

	if o.ID != 0 {
		if err := db.Model(&RoastRecipe{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			return err
		}
	} else {
		if err := db.Create(&o).Error; err != nil {
			return err
		}
	}
	return nil
}

type SearchRecipeForm struct {
	utils.Pager
	Name    string  `form:"name"`
	BeanID  string  `form:"bean_id"`
	Status  string  `form:"status"`
	Degree  string  `form:"degree"`
	BeanQty float64 `form:"beanqty"`
}

func SearchRecipe(q SearchRecipeForm) ([]RoastRecipe, int) {
	db := database.Conn
	var os []RoastRecipe

	var statement = db.Model(&RoastRecipe{}).Preload("Bean")
	if q.BeanID != "" {
		statement = statement.Where("bean_id = ?", q.BeanID)
	}
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	if q.Degree != "" {
		statement = statement.Where("degree = ?", q.Degree)
	}
	if q.BeanQty != 0 {
		statement = statement.Where("beanqty = ?", q.BeanQty)
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
