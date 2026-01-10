package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ExchangeRate 存储汇率信息
type ExchangeRate struct {
	Rate        float64   `json:"rate"`
	LastUpdated time.Time `json:"last_updated"`
}

var (
	// 缓存汇率数据
	rateCache = make(map[string]*ExchangeRate)
	// 汇率缓存过期时间（1小时）
	cacheExpiration = time.Hour
)

// GetExchangeRate 获取汇率
func GetExchangeRate(from, to string) (float64, error) {
	cacheKey := fmt.Sprintf("%s_%s", from, to)

	// 检查缓存
	if cached, exists := rateCache[cacheKey]; exists {
		if time.Since(cached.LastUpdated) < cacheExpiration {
			return cached.Rate, nil
		}
	}

	// 从API获取最新汇率
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", from)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("获取汇率失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %v", err)
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("解析响应失败: %v", err)
	}

	rate, exists := result.Rates[to]
	if !exists {
		return 0, fmt.Errorf("未找到汇率: %s -> %s", from, to)
	}

	// 更新缓存
	rateCache[cacheKey] = &ExchangeRate{
		Rate:        rate,
		LastUpdated: time.Now(),
	}

	return rate, nil
}

// ConvertPrice 转换价格
func ConvertPrice(price float64, from, to string) (float64, error) {
	if from == to {
		return price, nil
	}

	rate, err := GetExchangeRate(from, to)
	if err != nil {
		return 0, err
	}

	return price * rate, nil
}

// ConvertToGBP 将价格转换为英镑
func ConvertToGBP(price float64, from string) (float64, error) {
	return ConvertPrice(price, from, "GBP")
}
