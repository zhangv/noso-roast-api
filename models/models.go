package models

import (
	"time"

	"gorm.io/gorm"
)

type Status string

const (
	StatusValid   Status = "valid"
	StatusInvalid Status = "invalid"
	StatusDeleted Status = "deleted"
)

var CommonStatuses = map[Status]string{
	StatusValid:   "有效",
	StatusInvalid: "无效",
	StatusDeleted: "删除",
}

var Conn *gorm.DB

// Entity 表示任何包含基础Model字段的实体
type Entity interface {
	*Model | *LegacyModel
}

// Repository 通用CRUD仓储接口
type Repository[T Entity] interface {
	GetById(id uint) (T, error)
	Create(entity T) error
	Update(entity T) error
	Delete(id uint) error
	Search(query interface{}) ([]T, int)
}

// BaseRepository 基础实现
type BaseRepository[T Entity] struct {
	db *gorm.DB
}

func NewBaseRepository[T Entity](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) GetById(id uint) (T, error) {
	var entity T
	err := r.db.First(&entity, id).Error
	return entity, err
}

type Model struct {
	// gorm.Model
	ID uint `gorm:"primaryKey;autoIncrement" json:"id" form:"id"`
	// ID database.UUID `gorm:"primary_key;;default:(UUID_TO_BIN(UUID()));"`ref：https://medium.com/wesionary-team/implementation-of-uuid-and-binary-16-in-gorm-v2-1c329c352c91

	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
	// Attribute string `gorm:"column:attribute;type:text" json:"attribute"`
}

func ConvertMapToArray[K ~string](m map[K]string) []any {
	var result []any
	for k, v := range m {
		result = append(result, map[string]string{"title": v, "value": string(k)})
	}
	return result
}

/*
*

	func (m *Model) BeforeSave(tx *gorm.DB) (err error) {
		var attributes map[string]string
		if err := json.Unmarshal([]byte(m.Attribute), &attributes); err != nil {
			attributes = make(map[string]string)
		}

		modelValue := reflect.ValueOf(m).Elem()
		modelType := modelValue.Type()

		for i := 0; i < modelType.NumField(); i++ {
			field := modelType.Field(i)
			fieldValue := modelValue.Field(i)

			if fieldValue.IsZero() {
				continue
			}

			fieldName := field.Tag.Get("json")
			if fieldName == "" {
				fieldName = field.Name
			}

			if _, exists := attributes[fieldName]; !exists {
				attributes[fieldName] = fmt.Sprintf("%v", fieldValue.Interface())
			}
		}

		attributeBytes, err := json.Marshal(attributes)
		if err != nil {
			return err
		}

		m.Attribute = string(attributeBytes)
		return nil
	}

	func (m *Model) Prop(key string) (string, error) {
		//TODO 县检查model有没有这个属性，然后再看attribute
		var attributes map[string]string
		if err := json.Unmarshal([]byte(m.Attribute), &attributes); err != nil {
			return "", err
		}
		value, exists := attributes[key]
		if !exists {
			return "", errors.New("key not found in attributes")
		}
		return value, nil
	}

	*
*/
type LegacyModel struct {
	ID           int `gorm:"primaryKey" json:"id"`
	CreatedTime  int `gorm:"column:createdTime;type:int;size:32" json:"createdtime"`
	ModifiedTime int `gorm:"column:modifiedTime;type:int;size:32" json:"modifiedtime"`
}
