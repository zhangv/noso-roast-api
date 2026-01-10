package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// ConvertStructFields 自动转换 jsonMap 并映射到结构体
func ConvertStructFields(jsonMap map[string]interface{}, form interface{}) error {
	formValue := reflect.ValueOf(form)
	if formValue.Kind() != reflect.Ptr || formValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("form 必须是指向结构体的指针")
	}

	formValue = formValue.Elem()
	formType := formValue.Type()

	for i := 0; i < formType.NumField(); i++ {
		field := formType.Field(i)
		fieldValue := formValue.Field(i)
		jsonKey := field.Tag.Get("json")

		if jsonKey == "" {
			jsonKey = field.Name
		}

		// 处理 JSON 中没有该字段的情况
		jsonVal, exists := jsonMap[jsonKey]
		if !exists {
			continue
		}

		// 进行类型转换
		if err := convertAndSet(fieldValue, jsonVal); err != nil {
			return fmt.Errorf("字段 %s 转换失败: %v", jsonKey, err)
		}
	}

	return nil
}

// convertAndSet 处理不同数据类型的转换
func convertAndSet(fieldValue reflect.Value, jsonVal interface{}) error {
	if !fieldValue.CanSet() {
		return fmt.Errorf("字段不可设置")
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(fmt.Sprintf("%v", jsonVal))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fieldValue, jsonVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUint(fieldValue, jsonVal)

	case reflect.Float32, reflect.Float64:
		return setFloat(fieldValue, jsonVal)

	case reflect.Bool:
		return setBool(fieldValue, jsonVal)

	default:
		// 尝试 JSON 反序列化
		jsonBytes, err := json.Marshal(jsonVal)
		if err != nil {
			return fmt.Errorf("JSON 序列化失败: %v", err)
		}
		return json.Unmarshal(jsonBytes, fieldValue.Addr().Interface())
	}
	return nil
}

// setInt 处理整型转换
func setInt(fieldValue reflect.Value, jsonVal interface{}) error {
	switch v := jsonVal.(type) {
	case float64:
		fieldValue.SetInt(int64(v))
	case string:
		if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
			fieldValue.SetInt(intVal)
		} else {
			return err
		}
	default:
		return fmt.Errorf("无法转换为 int: %v", v)
	}
	return nil
}

// setUint 处理无符号整型转换
func setUint(fieldValue reflect.Value, jsonVal interface{}) error {
	switch v := jsonVal.(type) {
	case float64:
		fieldValue.SetUint(uint64(v))
	case string:
		if uintVal, err := strconv.ParseUint(v, 10, 64); err == nil {
			fieldValue.SetUint(uintVal)
		} else {
			return err
		}
	default:
		return fmt.Errorf("无法转换为 uint: %v", v)
	}
	return nil
}

// setFloat 处理浮点数转换
func setFloat(fieldValue reflect.Value, jsonVal interface{}) error {
	switch v := jsonVal.(type) {
	case float64:
		fieldValue.SetFloat(v)
	case string:
		if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
			fieldValue.SetFloat(floatVal)
		} else {
			return err
		}
	default:
		return fmt.Errorf("无法转换为 float: %v", v)
	}
	return nil
}

// setBool 处理布尔类型转换
func setBool(fieldValue reflect.Value, jsonVal interface{}) error {
	switch v := jsonVal.(type) {
	case bool:
		fieldValue.SetBool(v)
	case string:
		if boolVal, err := strconv.ParseBool(v); err == nil {
			fieldValue.SetBool(boolVal)
		} else {
			return err
		}
	default:
		return fmt.Errorf("无法转换为 bool: %v", v)
	}
	return nil
}

func IsEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func IsMobile(mobile string) bool {
	mobileRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return mobileRegex.MatchString(mobile)
}
