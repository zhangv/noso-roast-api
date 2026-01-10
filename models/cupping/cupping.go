package cupping

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/user"
	"hahacking/utils"
	"strings"

	"gorm.io/gorm"
)

type Cupping struct {
	models.Model
	Description string           `json:"description"`
	Status      CuppingStatus    `json:"status"`
	StatusL     string           `json:"statusL" gorm:"-"`
	Type        string           `json:"type" gorm:"column:type"`
	Location    string           `json:"location" gorm:"column:location"`
	Method      CuppingMethod    `json:"method" gorm:"column:method"`
	Scores      *[]CuppingScore  `json:"scores" gorm:"foreignKey:CuppingID;references:ID"`
	Samples     *[]CuppingSample `json:"samples" gorm:"foreignKey:CuppingID;references:ID"`
	SampleCount int              `json:"samplecount" gorm:"-"`
	ScoreCount  int              `json:"scorecount" gorm:"-"`
	UserCount   int              `json:"usercount" gorm:"-"`
	UserID      *uint            `json:"user_id"` //创建人
	User        *user.User       `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

func (o *Cupping) AfterFind(tx *gorm.DB) (err error) {
	var count int64
	tx.Model(&CuppingSample{}).Where("cupping_id = ?", o.ID).Count(&count)
	o.SampleCount = int(count)

	var scoreCount int64
	tx.Model(&CuppingScore{}).Where("cupping_id = ?", o.ID).Count(&scoreCount)
	o.ScoreCount = int(scoreCount)
	var userCount int64
	tx.Model(&CuppingScore{}).Where("cupping_id = ?", o.ID).Select("distinct user_id").Count(&userCount)
	o.UserCount = int(userCount)

	return
}

type CuppingStatus string // 杯测状态

const (
	CuppingStatusInvalid CuppingStatus = "invalid"
	CuppingStatusValid   CuppingStatus = "valid"
	CuppingStatusClosed  CuppingStatus = "closed"
)

var CuppingStatuses = map[CuppingStatus]string{
	CuppingStatusInvalid: "无效",
	CuppingStatusValid:   "有效",
	CuppingStatusClosed:  "关闭",
}

type CuppingMethod string // 杯测方法

const (
	CuppingMethodStandard CuppingMethod = "standard"
	CuppingMethodDrip     CuppingMethod = "drip"
	CuppingMethodEspresso CuppingMethod = "espresso"
	CuppingMethodFrench   CuppingMethod = "french"
	CuppingMethodOther    CuppingMethod = "other"
)

var CuppingMethods = map[CuppingMethod]string{
	CuppingMethodStandard: "标准",
	CuppingMethodDrip:     "滴滤",
	CuppingMethodEspresso: "意式",
	CuppingMethodFrench:   "法压",
	CuppingMethodOther:    "其他",
}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "CuppingMethod" {
			result[n] = models.ConvertMapToArray(CuppingMethods)
		}
		if n == "CuppingStatus" {
			result[n] = models.ConvertMapToArray(CuppingStatuses)
		}
		if n == "CuppingScoreStatus" {
			result[n] = models.ConvertMapToArray(models.CommonStatuses)
		}
	}
	return result
}

func GetById(id uint) (*Cupping, error) {
	var o Cupping
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchForm struct {
	utils.Pager
	Name   string `form:"name"`
	BeanID string `form:"beanid"`
	Status string `form:"status"`
}

func Search(q SearchForm) ([]Cupping, int) {
	db := database.Conn
	var os []Cupping
	var statement = db.Model(&Cupping{}).Preload("User").Preload("Samples.Bean").Preload("Samples.BeanBatch").Preload("Samples").Preload("Scores")
	if q.Name != "" {
		statement = statement.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.BeanID != "" {
		statement = statement.Where("bean_id = ?", q.BeanID)
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

func Update(o *Cupping) error {
	db := database.Conn
	tx := db.Begin()

	if o.ID != 0 {
		if err := tx.Model(&Cupping{}).Where("id = ?", o.ID).Updates(map[string]any{
			"description": o.Description,
			"status":      o.Status,
			"type":        o.Type,
			"location":    o.Location,
			"method":      o.Method,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Handle samples
	if o.Samples != nil {
		// Get existing samples
		var existingSamples []CuppingSample
		if err := tx.Where("cupping_id = ?", o.ID).Find(&existingSamples).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Create a map of existing samples by ID
		existingMap := make(map[uint]*CuppingSample)
		for i := range existingSamples {
			existingMap[existingSamples[i].ID] = &existingSamples[i]
		}

		// Process each sample
		for i := range *o.Samples {
			sample := &(*o.Samples)[i]
			sample.CuppingID = o.ID

			if sample.ID != 0 && existingMap[sample.ID] != nil {
				// Update existing sample
				if err := tx.Model(&CuppingSample{}).Where("id = ?", sample.ID).Updates(map[string]interface{}{
					"bean_id":         sample.BeanID,
					"beanbatch_id":    sample.BeanBatchID,
					"beanqty":         sample.BeanQty,
					"grindsize":       sample.GrindSize,
					"watertemp":       sample.WaterTemp,
					"extractionyield": sample.ExtractionYield,
					"method":          sample.Method,
					"triangle":        sample.Triangle,
				}).Error; err != nil {
					tx.Rollback()
					return err
				}
				delete(existingMap, sample.ID)
			} else {
				// Create new sample
				if err := tx.Create(sample).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}

		// Delete samples that are no longer in the list
		for id := range existingMap {
			if err := tx.Delete(&CuppingSample{}, id).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	tx.Commit()
	return nil
}

func Delete(id int) error {
	db := database.GetConn()
	tx := db.Begin()

	// Delete associated samples first
	if err := tx.Where("cupping_id = ?", id).Delete(&CuppingSample{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete associated scores
	if err := tx.Where("cupping_id = ?", id).Delete(&CuppingScore{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the cupping record
	if err := tx.Delete(&Cupping{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
