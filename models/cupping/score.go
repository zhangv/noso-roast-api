package cupping

import (
	"database/sql/driver"
	"encoding/json"
	"hahacking/database"
	"hahacking/database/datatypes"
	"hahacking/models"
	"hahacking/models/bean"
	"hahacking/utils"

	"log"

	"gorm.io/gorm"
)

type CuppingScore struct {
	models.Model
	UserID          uint       `json:"user_id"`
	BeanID          *uint      `json:"bean_id" gorm:"column:bean_id;null"`
	Bean            *bean.Bean `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	CuppingID       uint       `json:"cupping_id" gorm:"column:cupping_id"`
	Cupping         *Cupping   `json:"cupping" gorm:"foreignKey:CuppingID;references:ID"`
	CuppingSampleID uint       `json:"cuppingsample_id" gorm:"column:cuppingsample_id"`

	CuppingScoreItem

	Total        float64       `json:"total" gorm:"total"`                 // 总分 //获取汇总时候用到，其实不存储
	Count        int           `json:"count" gorm:"count"`                 // 评分次数 //获取汇总时候会用到，平时不用
	DefectsTotal float64       `json:"defects_total" gorm:"defects_total"` // 瑕疵总分
	Status       models.Status `json:"status"`
}
type CuppingScoreItem struct { //评分项
	Aroma         float64        `json:"aroma" gorm:"not null"`      // 干香
	Fragrance     float64        `json:"fragrance" gorm:"not null"`  // 湿香
	AromaTags     datatypes.Slot `json:"aroma_tags" gorm:"not null"` // 香气标签
	AromaNotes    string         `json:"aroma_notes"`                // 香气备注
	Flavor        float64        `json:"flavor" gorm:"not null"`     // 风味（强度）
	Aftertaste    float64        `json:"aftertaste" gorm:"not null"` // 回甘
	FlavorTags    datatypes.Slot `json:"flavor_tags" gorm:""`        // 风味标签
	FlavorNotes   string         `json:"flavor_notes"`               // 风味备注
	Acidity       float64        `json:"acidity" gorm:"not null"`    // 酸质
	Sweetness     float64        `json:"sweetness" gorm:"not null"`  // 甜度
	Body          float64        `json:"body" gorm:"not null"`       // 口感
	BodyTags      datatypes.Slot `json:"body_tags"`                  // 口感标签
	PrimaryTastes datatypes.Slot `json:"primary_tastes"`             // 主要风味
	// 以下3个字新的cva表里已经没有了
	// Balance    float64 `json:"balance" gorm:"not null"`    // 平衡感
	// Cleancup   float64 `json:"cleancup" gorm:"not null"`   // 干净度
	// Uniformity float64 `json:"uniformity" gorm:"not null"` // 一致性
	Overall float64  `json:"overall" gorm:"not null"`                      // 整体评分
	Defects *Defects `json:"defects" gorm:"column:defects;type:json;null"` // 瑕疵
	Notes   string   `json:"notes"`
}

type Defects []Defect
type Defect struct {
	Name  string `json:"name"`
	Level string `json:"level"`
	Desc  string `json:"desc"`
}

func (d Defects) Total() float64 {
	total := 0
	for _, defect := range d {
		if defect.Level == "quark" {
			total += 2
		}
		if defect.Level == "full" {
			total += 4
		}
	}
	return float64(total)
}

func (d *Defects) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		log.Printf("Failed to unmarshal Defects value: %v", value)
		return nil

	}
	return json.Unmarshal(bytes, d)
}

func (d Defects) Value() (driver.Value, error) {
	return json.Marshal(d)
}
func (o *CuppingScore) BeforeSave(tx *gorm.DB) (err error) {
	o.Total = o.Aroma + o.Fragrance +
		o.Flavor + o.Acidity + o.Body + o.Aftertaste + o.Sweetness + o.Overall
	if o.Defects != nil {
		o.DefectsTotal = o.Defects.Total()
	}
	return
}

func (o *CuppingScore) AfterFind(tx *gorm.DB) (err error) {
	//改为在保存时更新 2025-05-14
	// o.Total = o.Fragrance + o.Flavor + o.Acidity + o.Body + o.Balance + o.Aftertaste + o.Overall + o.Cleancup + o.Uniformity + o.Sweetness
	// o.DefectsTotal = o.Defects.Total()

	return
}

// SearchScoreForm defines the search criteria for cupping scores
type SearchScoreForm struct {
	utils.Pager
	UserID    uint          `form:"user_id"`
	CuppingID uint          `form:"cupping_id"`
	SampleID  uint          `form:"sample_id"`
	BeanID    uint          `form:"bean_id"`
	Status    models.Status `form:"status"`
}

// SearchScore searches for cupping scores based on the provided criteria
func SearchScore(q SearchScoreForm) ([]CuppingScore, int) {
	db := database.GetConn()
	var os []CuppingScore
	var statement = db.Model(&CuppingScore{}).Preload("Bean").Preload("Cupping")

	if q.UserID != 0 {
		statement = statement.Where("user_id = ?", q.UserID)
	}
	if q.CuppingID != 0 {
		statement = statement.Where("cupping_id = ?", q.CuppingID)
	}
	if q.SampleID != 0 {
		statement = statement.Where("cuppingsample_id = ?", q.SampleID)
	}
	if q.BeanID != 0 {
		statement = statement.Where("bean_id = ?", q.BeanID)
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

// GetScoreById retrieves a cupping score by its ID
func GetScoreById(id uint) (*CuppingScore, error) {
	var o CuppingScore
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

// 获取样本的得分汇总
func GetScoresBySampleId(sampleId uint) ([]CuppingScore, error) {

	var scores []CuppingScore
	db := database.Conn
	err := db.Where("cuppingsample_id = ?", sampleId).Find(&scores).Error
	if err != nil {
		return nil, err
	}
	return scores, nil
}

// UpdateScore updates or creates a cupping score
func UpdateScore(o *CuppingScore) error {
	db := database.GetConn()
	tx := db.Begin()

	if o.ID != 0 {
		// Update existing score
		if err := tx.Model(&CuppingScore{}).Where("id = ?", o.ID).Updates(map[string]any{
			"user_id":          o.UserID,
			"cupping_id":       o.CuppingID,
			"cuppingsample_id": o.CuppingSampleID,
			"aroma":            o.Aroma,
			"aroma_tags":       o.AromaTags,
			"aroma_notes":      o.AromaNotes,
			"fragrance":        o.Fragrance,
			"flavor":           o.Flavor,
			"flavor_tags":      o.FlavorTags,
			"flavor_notes":     o.FlavorNotes,
			"acidity":          o.Acidity,
			"body":             o.Body,
			"body_tags":        o.BodyTags,
			"aftertaste":       o.Aftertaste,
			"overall":          o.Overall,
			"sweetness":        o.Sweetness,
			// "cleancup":         o.Cleancup,
			// "uniformity":       o.Uniformity,
			"defects": o.Defects,
			"notes":   o.Notes,
			"status":  o.Status,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 创建新记录时，排除已废弃的字段（balance, cleancup, uniformity）
		if err := tx.Omit("balance", "cleancup", "uniformity").Create(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// DeleteScore deletes a cupping score and updates the sample's average score
func DeleteScore(id uint) error {
	db := database.GetConn()
	tx := db.Begin()

	// Get the score first to know the sample ID
	var score CuppingScore
	if err := tx.First(&score, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the score
	if err := tx.Delete(&CuppingScore{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

type CuppingScoreSummary struct {
	CuppingScore
	Count int `json:"count"`
	// Total float64 `json:"total"`
	// Defects  int  `json:"defects"` //是否要汇总全部瑕疵（合并全部json）
	DefectsTotal float64 `json:"defects_total"`
}

func GetSampleScoreSummary(sampleID uint) (*CuppingScoreSummary, error) {
	db := database.GetConn()
	var summary CuppingScoreSummary
	rawSQL := getSummaryRawSQL("cuppingsample_id")
	err := db.Raw(rawSQL, sampleID).Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func GetBeanScoreSummary(beanID uint) (*CuppingScoreSummary, error) {
	db := database.GetConn()
	var summary CuppingScoreSummary
	rawSQL := getSummaryRawSQL("bean_id")
	err := db.Raw(rawSQL, beanID).Scan(&summary).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func getSummaryRawSQL(groupBy string) string {
	sql := `SELECT 
			` + groupBy + `,
			COUNT(*) as count,
			ROUND(AVG(aroma), 1) as aroma,
			ROUND(AVG(fragrance), 1) as fragrance,
			ROUND(AVG(flavor), 1) as flavor,
			ROUND(AVG(acidity), 1) as acidity,
			ROUND(AVG(body), 1) as body,
			ROUND(AVG(aftertaste), 1) as aftertaste,
			ROUND(AVG(sweetness), 1) as sweetness,
			ROUND(AVG(overall), 1) as overall,
			ROUND(AVG(aroma + fragrance + flavor + acidity + body + 
					 aftertaste + sweetness + overall), 1) as total,
			ROUND(AVG(defects_total), 1) as defects_total
		FROM t_cupping_score 
		WHERE ` + groupBy + ` = ? GROUP BY ` + groupBy

	// 	SELECT
	// 		cuppingsample_id ,
	// 		COUNT(*) as count,
	// 		ROUND(AVG(aroma), 1) as aroma,
	// 		ROUND(AVG(fragrance), 1) as fragrance,
	// 		ROUND(AVG(flavor), 1) as flavor,
	// 		ROUND(AVG(acidity), 1) as acidity,
	// 		ROUND(AVG(body), 1) as body,
	// 		ROUND(AVG(aftertaste), 1) as aftertaste,
	// 		ROUND(AVG(sweetness), 1) as sweetness,
	// 		ROUND(AVG(overall), 1) as overall,
	// 		ROUND(AVG(aroma + fragrance + flavor + acidity + body +
	// 				 aftertaste + sweetness + overall), 1) as total,
	// 		ROUND(AVG(defects_total), 1) as defects_total
	// 	FROM t_cupping_score
	// 	WHERE cuppingsample_id = ?
	// 	GROUP BY cuppingsample_id
	return sql
}
