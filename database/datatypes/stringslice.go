package datatypes

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	// 示例使用 SQLite 数据库
)

// --- 自定义类型 StringSlice ---

// StringSlice 是 []string 的自定义类型，用于数据库的 JSON 存储
type StringSlice []string

// Value 方法将 StringSlice 转换为 JSON 字符串，以保存到数据库
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil // 如果切片为空，保存为 NULL
	}
	// 将 []string 编码为 JSON 字节数组
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(jsonBytes), nil // 返回 JSON 字符串
}

// Scan 方法从数据库中读取 JSON 字符串，并反序列化回 StringSlice
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil // 如果数据库值为 NULL，则将其设为空切片
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprintf("unsupported type for StringSlice scan: %T", value))
	}

	// 将 JSON 字节数组解码为 []string
	if len(bytes) == 0 { // 处理空字符串情况
		*s = nil
		return nil
	}

	err := json.Unmarshal(bytes, s)
	if err != nil {
		return err
	}
	return nil
}
