package cupping

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/bean"
	"hahacking/utils"

	"gorm.io/gorm"
)

type CuppingSample struct {
	models.Model
	Memo            string         `json:"memo" gorm:"column:memo"` //备注
	CuppingID       uint           `json:"cupping_id" gorm:"column:cupping_id"`
	Cupping         Cupping        `json:"cupping" gorm:"foreignKey:CuppingID;references:ID"`
	RoastID         uint           `json:"roast_id" gorm:"column:roast_id"`
	BeanID          uint           `json:"bean_id" gorm:"column:bean_id"`
	Bean            bean.Bean      `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	BeanBatchID     uint           `json:"beanbatch_id" gorm:"column:beanbatch_id;default:null"`
	BeanBatch       bean.BeanBatch `json:"beanbatch" gorm:"foreignKey:BeanBatchID;references:ID"`
	BeanQty         uint           `json:"beanqty" gorm:"column:beanqty"`
	ExtractionYield string         `json:"extractionyield" gorm:"column:extractionyield"`
	GrindSize       string         `json:"grindsize" gorm:"column:grindsize"` //研磨度
	WaterTemp       float64        `json:"watertemp" gorm:"column:watertemp"` //水温
	Method          CuppingMethod  `json:"method" gorm:"column:method"`       //样本杯测方法 - 适用于一次杯测中，同样的豆，不同的杯测方法
	Triangle        string         `json:"triangle" gorm:"column:triangle"`   //三角杯测中的特殊样本 - 标记
	Scores          []CuppingScore `json:"scores" gorm:"foreignKey:CuppingSampleID;references:ID"`
	// Score           map[string]float64 `json:"score" gorm:"column:-"` //根据全部当前的评分计算出的平均分
	UserID *uint `json:"user_id"` //创建人
}

// AfterFind calculates the total score after loading from database
func (s *CuppingSample) AfterFind(tx *gorm.DB) error {

	return nil
}

// SearchSampleForm defines the search criteria for cupping samples
type SearchSampleForm struct {
	utils.Pager
	CuppingID uint   `form:"cupping_id"`
	BeanID    uint   `form:"bean_id"`
	RoastID   uint   `form:"roast_id"`
	Method    string `form:"method"`
}

// SearchSample searches for cupping samples based on the provided criteria
func SearchSample(q SearchSampleForm) ([]CuppingSample, int) {
	db := database.GetConn()
	var os []CuppingSample
	var statement = db.Model(&CuppingSample{}).
		Preload("Bean").
		Preload("BeanBatch").
		Preload("Scores").
		Preload("Cupping")

	if q.CuppingID != 0 {
		statement = statement.Where("cupping_id = ?", q.CuppingID)
	}
	if q.BeanID != 0 {
		statement = statement.Where("bean_id = ?", q.BeanID)
	}
	if q.RoastID != 0 {
		statement = statement.Where("roast_id = ?", q.RoastID)
	}
	if q.Method != "" {
		statement = statement.Where("method = ?", q.Method)
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

// GetSampleById retrieves a cupping sample by its ID
func GetSampleById(id uint) (*CuppingSample, error) {
	var o CuppingSample
	db := database.GetConn()
	err := db.Where("id = ?", id).
		Preload("Bean").
		Preload("BeanBatch").
		Preload("Scores").
		Preload("Cupping").
		First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

// UpdateSample updates or creates a cupping sample
func UpdateSample(o *CuppingSample) error {
	db := database.Conn
	tx := db.Begin()

	if o.ID != 0 {
		if err := tx.Model(&CuppingSample{}).Where("id = ?", o.ID).Updates(o).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Handle scores if provided
	// if o.Scores != nil {
	// 	for _, score := range o.Scores {
	// 		score.CuppingSampleID = o.ID
	// 		if err := tx.Save(&score).Error; err != nil {
	// 			tx.Rollback()
	// 			return err
	// 		}
	// 	}
	// }

	tx.Commit()
	return nil
}

func (s *CuppingSample) Delete() error {
	db := database.GetConn()
	return db.Delete(s).Error
}

func (s *CuppingSample) Load() error {
	db := database.GetConn()
	return db.Preload("Bean").Preload("BeanBatch").Preload("Scores").Preload("Cupping").First(s).Error
}
