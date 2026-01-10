package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPHPTimeFormatRelativeFixed 测试固定日期格式的相对时间
func TestPHPTimeFormatRelativeFixed(t *testing.T) {
	fmt.Println("=== 测试固定日期格式的相对时间 ===")

	now := time.Now()
	currentYear := now.Year()

	testCases := []struct {
		relativeTime string
		format       string
		expected     string
	}{
		{"-1 years", "Y-01-01", fmt.Sprintf("%d-01-01", currentYear-1)},
		{"-1 year", "YYYY-01-01", fmt.Sprintf("%d-01-01", currentYear-1)},
		{"-1 year", "Y-12-31", fmt.Sprintf("%d-12-31", currentYear-1)},
		{"-1 year", "YYYY-12-31", fmt.Sprintf("%d-12-31", currentYear-1)},
		{"+1 year", "Y-01-01", fmt.Sprintf("%d-01-01", currentYear+1)},
		{"+1 year", "YYYY-01-01", fmt.Sprintf("%d-01-01", currentYear+1)},
		{"+1 year", "Y-12-31", fmt.Sprintf("%d-12-31", currentYear+1)},
		{"+1 year", "YYYY-12-31", fmt.Sprintf("%d-12-31", currentYear+1)},
		{"-2 years", "Y-01-01", fmt.Sprintf("%d-01-01", currentYear-2)},
		{"-2 years", "Y-12-30", fmt.Sprintf("%d-12-30", currentYear-2)},
		{"+2 years", "Y-01-01", fmt.Sprintf("%d-01-01", currentYear+2)},
		{"+2 years", "Y-12-31", fmt.Sprintf("%d-12-31", currentYear+2)},

		// 注意：对于月份相对时间，固定日期格式仍然保持固定日期（使用当前年份）
		{"-1 month", "Y-01-01", fmt.Sprintf("%d-01-01", currentYear)},
		{"-1 month", "Y-12-31", fmt.Sprintf("%d-12-31", currentYear)},
		{"+1 month", "Y-01-01", fmt.Sprintf("%d-01-01", currentYear)},
		{"+1 month", "Y-12-31", fmt.Sprintf("%d-12-31", currentYear)},
		// today 和 yesterday 使用普通格式化，期望值设为空字符串，在下面单独验证
		{"today", "Y-01-01", ""},     // 今天不是固定日期，使用普通格式化
		{"yesterday", "Y-01-01", ""}, // 昨天不是固定日期，使用普通格式化
	}

	for _, tc := range testCases {
		result, err := PHPTimeFormatRelativeFixed(tc.relativeTime, tc.format)
		if err != nil {
			fmt.Printf("相对时间: %-15s, 格式: %-15s -> 错误: %s\n", tc.relativeTime, tc.format, err)
		} else {
			fmt.Printf("相对时间: %-15s, 格式: %-15s -> 结果: %s\n", tc.relativeTime, tc.format, result)
			if tc.expected != "" {
				assert.Equal(t, tc.expected, result)
			}
		}
	}

	// 单独验证 today 和 yesterday 使用普通格式化
	todayResult, _ := PHPTimeFormatRelativeFixed("today", "Y-m-d")
	todayExpected := time.Now().Format("2006-01-02")
	assert.Equal(t, todayExpected, todayResult)

	yesterdayResult, _ := PHPTimeFormatRelativeFixed("yesterday", "Y-m-d")
	yesterdayExpected := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	assert.Equal(t, yesterdayExpected, yesterdayResult)

	tmp, _ := PHPTimeFormatRelativeFixed("-1 years", "Y-01-01")
	toyear := time.Now().AddDate(-1, 0, 0).Format("2006")
	assert.Equal(t, toyear+"-01-01", tmp)
	tmp, _ = PHPTimeFormatRelativeFixed("-0 years", "Y-m-d")
	today := time.Now().Format("2006-01-02")
	assert.Equal(t, today, tmp)

	tmp, _ = PHPTimeFormatRelativeFixed("+0 year", "Y-m-d H:i:s")
	today = time.Now().Format("2006-01-02 15:04:05")
	assert.Equal(t, today, tmp)
}
