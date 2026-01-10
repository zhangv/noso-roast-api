package utils

import "fmt"

const (
	// 重量单位转换常量
	KG_TO_LB = 2.20462262    // 1千克 = 2.20462262磅
	LB_TO_KG = 0.45359237    // 1磅 = 0.45359237千克
	G_TO_LB  = 0.00220462262 // 1克 = 0.00220462262磅
	LB_TO_G  = 453.59237     // 1磅 = 453.59237克
)

// WeightUnit 重量单位
type WeightUnit string

const (
	Kilogram WeightUnit = "kg"
	Gram     WeightUnit = "g"
	Pound    WeightUnit = "lb"
)

// ConvertWeight 转换重量
func ConvertWeight(value float64, from, to WeightUnit) (float64, error) {
	if from == to {
		return value, nil
	}

	// 先将所有单位转换为克
	var grams float64
	switch from {
	case Kilogram:
		grams = value * 1000
	case Gram:
		grams = value
	case Pound:
		grams = value * LB_TO_G
	default:
		return 0, fmt.Errorf("不支持的重量单位: %s", from)
	}

	// 从克转换为目标单位
	switch to {
	case Kilogram:
		return grams / 1000, nil
	case Gram:
		return grams, nil
	case Pound:
		return grams * G_TO_LB, nil
	default:
		return 0, fmt.Errorf("不支持的重量单位: %s", to)
	}
}

// ConvertToPounds 将重量转换为磅
func ConvertToPounds(value float64, from WeightUnit) (float64, error) {
	return ConvertWeight(value, from, Pound)
}

// ConvertFromPounds 从磅转换为其他单位
func ConvertFromPounds(value float64, to WeightUnit) (float64, error) {
	return ConvertWeight(value, Pound, to)
}
