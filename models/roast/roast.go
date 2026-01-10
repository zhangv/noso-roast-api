package roast

import (
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/bean"
	"hahacking/utils"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Roast struct {
	models.Model
	Name         string         `json:"name"`
	Date         string         `json:"date"` //烘焙日期
	InQty        int            `json:"inqty" gorm:"column:inqty"`
	OutQty       int            `json:"outqty" gorm:"column:outqty"`
	Degree       RoastDegree    `json:"degree"`
	BeanID       uint           `json:"bean_id" gorm:"column:bean_id"`
	Bean         bean.Bean      `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	Status       RoastStatus    `json:"status"`
	StatusL      string         `json:"statusL" gorm:"-"`
	StartTime    string         `json:"starttime" gorm:"column:starttime"`
	CompleteTime string         `json:"completetime" gorm:"column:completetime"`
	Agtron       int            `json:"agtron" gorm:"column:agtron"`
	AgtronGround int            `json:"agtron_ground" gorm:"column:agtron_ground"`
	LossRate     float64        `json:"lossrate" gorm:"column:lossrate"`
	CancelTime   string         `json:"canceltime" gorm:"column:canceltime"`
	Items        []RoastItem    `json:"items" gorm:"foreignKey:RoastID;references:ID"`
	BeanBatchID  uint           `json:"beanbatch_id" gorm:"column:beanbatch_id;default:null"` //熟豆批次ID
	BeanBatch    bean.BeanBatch `json:"beanbatch" gorm:"foreignKey:BeanBatchID;references:ID"`
}
type RoastItem struct {
	models.Model
	BeanBatchID uint `json:"beanbatch_id" gorm:"column:beanbatch_id;default:null"` //生豆批次ID
	// BeanBatch   *bean.BeanBatch `json:"beanbatch" gorm:"foreignKey:BeanBatchID;references:ID"`
	RoastID uint `json:"roast_id" gorm:"column:roast_id"`
	// Roast       *Roast          `json:"roast" gorm:"foreignKey:RoastID;references:ID"`
	Qty int `json:"qty"`
}

type RoastStatus string

const (
	RoastStatusStart    RoastStatus = "start"
	RoastStatusComplete RoastStatus = "complete"
	RoastStatusCancel   RoastStatus = "cancel"
	RoastStatusDelete   RoastStatus = "delete"
)

var RoastStatuses = map[RoastStatus]string{
	RoastStatusStart:    "开始",
	RoastStatusComplete: "完成",
	RoastStatusCancel:   "取消",
	RoastStatusDelete:   "删除",
}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "RoastStatus" {
			result[n] = models.ConvertMapToArray(RoastStatuses)
		}
		if n == "RoastDegree" {
			result[n] = models.ConvertMapToArray(RoastDegrees)
		}
	}
	return result
}

func GetById(id uint) (*Roast, error) {
	var o Roast
	db := database.GetConn()
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func (o *Roast) GetItems() ([]RoastItem, error) {
	var items []RoastItem
	db := database.GetConn()
	err := db.Where("roast_id = ?", o.ID).Find(&items).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return items, nil
}

type SearchForm struct {
	utils.Pager
	Name   string `form:"name"`
	BeanID string `form:"bean_id"`
	Status string `form:"status"`
}

func Search(q SearchForm) ([]Roast, int) {
	db := database.Conn
	var os []Roast

	var statement = db.Model(&Roast{}).Preload("Items").Preload("Bean").Preload("BeanBatch")
	if q.BeanID != "" {
		statement = statement.Where("bean_id = ?", q.BeanID)
	}
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

func Update(o *Roast) error {
	db := database.Conn
	tx := db.Begin()
	if o.ID != 0 {
		if err := tx.Model(&Roast{}).Where("id = ?", o.ID).Updates(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	for i := range o.Items {
		o.Items[i].RoastID = o.ID // 确保外键正确
		if err := tx.Save(&o.Items[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

type StartForm struct {
	Date   string      `json:"date" form:"date"`
	Name   string      `json:"name"`
	BeanID uint        `json:"bean_id"`
	InQty  int         `json:"inqty"`
	Items  []RoastItem `json:"items"`
}

func (form *StartForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Name == "" {
		result["name"] = "请填写名称"
	}
	if form.Date == "" {
		result["date"] = "请填写日期"
	}

	if len(form.Items) == 0 {
		result["items"] = "请选择生豆和批次"
	}

	return result
}

func Start(form StartForm) (*Roast, error) {
	db := database.Conn
	tx := db.Begin()
	roast := Roast{
		Name:      form.Name,
		Date:      form.Date,
		InQty:     form.InQty,
		BeanID:    form.BeanID,
		Status:    RoastStatusStart,
		StartTime: time.Now().Format("2006-01-02 15:04"),
	}
	if err := tx.Create(&roast).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, item := range form.Items {
		item := RoastItem{
			RoastID:     roast.ID,
			BeanBatchID: item.BeanBatchID,
			Qty:         item.Qty,
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	tx.Commit()
	return &roast, nil
}

type CompleteForm struct {
	ID           uint    `json:"id" form:"id"`
	OutQty       int     `json:"outqty" form:"outqty"`
	Agtron       int     `json:"agtron" form:"agtron"`
	AgtronGround int     `json:"agtron_ground" form:"agtron_ground"`
	Degree       string  `json:"degree" form:"degree"`
	LossRate     float64 `json:"lossrate" form:"lossrate"`
}

func (form *CompleteForm) Validate() map[string]string {
	var result = make(map[string]string)
	roast, err := GetById(form.ID)
	if err != nil {
		result["id"] = "烘焙不存在"
	}
	if roast.Status != RoastStatusStart {
		result["id"] = "烘焙未开始"
	}
	if form.OutQty <= 0 {
		result["outqty"] = "请填写熟豆重量"
	}
	if form.OutQty > roast.InQty {
		result["outqty"] = "熟豆重量不能大于生豆重量(" + strconv.Itoa(roast.InQty) + "g)"
	}
	if form.Agtron == 0 {
		result["agtron"] = "请填写豆色值"
	}
	if form.AgtronGround == 0 {
		result["agtron_ground"] = "请填写粉色值"
	}
	if form.Degree == "" {
		result["degree"] = "请选择烘焙度"
	}
	return result
}
func (roast *Roast) Complete(form CompleteForm) error {
	db := database.Conn
	//修改生豆批次
	items, err := roast.GetItems()
	if err != nil {
		return err
	}
	for _, item := range items {
		beanBatch, err := bean.GetBatchById(item.BeanBatchID)
		if err != nil {
			return err
		}

		beanBatch.Log(-item.Qty, "roast", strconv.FormatUint(uint64(roast.ID), 10),
			"烘焙完成"+strconv.FormatUint(uint64(roast.ID), 10))
	}

	// 生豆出库日志
	b, err := bean.GetById(roast.BeanID)
	if err != nil {
		// tx.Rollback()
		return err
	}

	if _, err := b.Stocking(int(roast.InQty)*-1, bean.BeanTypeGreen, bean.StockLogTypeProduce,
		"Roast", strconv.FormatUint(uint64(roast.ID), 10), "烘焙生豆出库"+strconv.FormatUint(uint64(roast.ID), 10)); err != nil {
		// tx.Rollback()
		return err
	}

	outQty := form.OutQty
	// 熟豆批次入库
	roastedBatch := bean.BeanBatch{
		BeanID:     roast.BeanID,
		BeanType:   bean.BeanTypeRoasted,
		TotalQty:   outQty,
		CurrentQty: outQty,
		Type:       bean.StockLogTypeProduce,
		Status:     bean.BatchStatusValid,
		Date:       roast.Date,
	}
	if err := db.Create(&roastedBatch).Error; err != nil {
		// tx.Rollback()
		return err
	}

	//更新烘焙完成
	if err := db.Model(&roast).Updates(
		Roast{
			Status:       RoastStatusComplete,
			OutQty:       outQty,
			CompleteTime: time.Now().Format("2006-01-02 15:04"),
			Agtron:       form.Agtron,
			AgtronGround: form.AgtronGround,
			LossRate:     form.LossRate,
			Degree:       RoastDegree(form.Degree),
			BeanBatchID:  roastedBatch.ID,
		}).Error; err != nil {
		// tx.Rollback()
		return err
	}
	//更新烘焙profile
	profile, err := GetProfileByRoastId(roast.ID)
	if err != nil {
		return err
	}
	profile.Agtron = form.Agtron
	profile.AgtronGround = form.AgtronGround
	profile.Degree = RoastDegree(form.Degree)
	profile.LossRate = form.LossRate
	profile.BeanID = roast.BeanID
	UpdateProfile(profile)

	//熟豆入库日志
	if _, err := b.Stocking(outQty, bean.BeanTypeRoasted, bean.StockLogTypeProduce,
		"Roast", strconv.FormatUint(uint64(roast.ID), 10), "烘焙熟豆入库"+strconv.FormatUint(uint64(roast.ID), 10)); err != nil {
		// tx.Rollback()
		return err
	}

	// tx.Commit()
	return nil
}

func (o *Roast) OnComplete(data CompleteForm, tx *gorm.DB) error {

	return nil
}

func Cancel(id int) error {
	db := database.Conn
	if err := db.Model(&Roast{}).Where("id = ?", id).Updates(Roast{Status: "cancel", CancelTime: time.Now().Format("2006-01-02 15:04")}).Error; err != nil {
		return err
	}
	return nil
}
func Delete(id int) error {
	db := database.Conn
	if err := db.Delete(&Roast{}, id).Error; err != nil {
		return err
	}
	return nil
}
