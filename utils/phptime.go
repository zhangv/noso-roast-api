package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PHPTimeFormatRelativeFixed 解析相对时间并格式化，支持固定日期格式
// 例如: -1 year + Y-01-01 = 2024-01-01
func PHPTimeFormatRelativeFixed(relativeTime string, format string) (string, error) {
	// 先解析相对时间
	t, err := ParseRelativeTime(relativeTime)
	if err != nil {
		return "", err
	}

	// 对于 "today" 和 "yesterday"，即使格式包含固定日期，也使用普通格式化
	timeStr := strings.TrimSpace(strings.ToLower(relativeTime))
	if timeStr == "today" || timeStr == "yesterday" {
		return PHPTimeFormat(t, format), nil
	}

	// 检查格式是否包含固定日期（如 01-01, 12-31 等）
	if containsFixedDate(format) {
		// 检查是否是月份相对时间（对于月份相对时间，固定日期格式使用当前年份）
		isMonthRelative := isMonthRelativeTime(relativeTime)
		return applyFixedDateToTime(t, format, isMonthRelative), nil
	}

	// 如果不包含固定日期，使用普通格式化
	return PHPTimeFormat(t, format), nil
}

// isMonthRelativeTime 检查是否是月份相对时间
func isMonthRelativeTime(relativeTime string) bool {
	timeStr := strings.TrimSpace(strings.ToLower(relativeTime))
	monthKeywords := []string{"month", "months"}
	for _, keyword := range monthKeywords {
		if strings.Contains(timeStr, keyword) {
			return true
		}
	}
	return false
}

// containsFixedDate 检查格式是否包含固定日期
func containsFixedDate(format string) bool {
	// 使用正则表达式匹配任意固定日期格式
	// 匹配模式：数字-数字 或 数字/数字 或 数字月数字日
	fixedDateRegex := regexp.MustCompile(`\d{1,2}[-/]\d{1,2}|\d{1,2}月\d{1,2}日|\d{1,2}\.\d{1,2}`)
	return fixedDateRegex.MatchString(format)
}

// extractFixedDate 从格式中提取固定日期信息
func extractFixedDate(format string) (month, day int, separator string, monthStr, dayStr string, err error) {
	// 匹配不同的日期格式
	patterns := []struct {
		regex      *regexp.Regexp
		separator  string
		monthIndex int
		dayIndex   int
	}{
		{regexp.MustCompile(`(\d{1,2})-(\d{1,2})`), "-", 1, 2},  // 01-01
		{regexp.MustCompile(`(\d{1,2})/(\d{1,2})`), "/", 1, 2},  // 01/01
		{regexp.MustCompile(`(\d{1,2})月(\d{1,2})日`), "月", 1, 2}, // 01月01日
		{regexp.MustCompile(`(\d{1,2})\.(\d{1,2})`), ".", 1, 2}, // 01.01
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(format)
		if len(matches) >= 3 {
			month, err1 := strconv.Atoi(matches[pattern.monthIndex])
			day, err2 := strconv.Atoi(matches[pattern.dayIndex])
			if err1 == nil && err2 == nil && month >= 1 && month <= 12 && day >= 1 && day <= 31 {
				return month, day, pattern.separator, matches[pattern.monthIndex], matches[pattern.dayIndex], nil
			}
		}
	}

	return 0, 0, "", "", "", fmt.Errorf("无法解析固定日期格式: %s", format)
}

// applyFixedDateToTime 将固定日期应用到时间上
// isMonthRelative: 如果是月份相对时间，使用当前年份；否则使用计算后的年份
func applyFixedDateToTime(t time.Time, format string, isMonthRelative bool) string {
	// 提取年份
	year := t.Year()
	// 如果是月份相对时间，使用当前年份而不是计算后的年份
	if isMonthRelative {
		year = time.Now().Year()
	}

	// 提取固定日期信息
	_, _, _, monthStr, dayStr, err := extractFixedDate(format)
	if err != nil {
		// 如果无法解析固定日期，回退到普通格式化
		return PHPTimeFormat(t, format)
	}

	// 替换年份占位符
	result := format
	result = strings.ReplaceAll(result, "YYYY", fmt.Sprintf("%d", year))
	result = strings.ReplaceAll(result, "YY", fmt.Sprintf("%02d", year%100))
	result = strings.ReplaceAll(result, "Y", fmt.Sprintf("%d", year))

	// 使用原始的月日字符串，保持格式一致

	// 替换单字符格式
	result = strings.ReplaceAll(result, "m", monthStr)
	result = strings.ReplaceAll(result, "d", dayStr)

	// 替换多字符格式
	result = strings.ReplaceAll(result, "MM", monthStr)
	result = strings.ReplaceAll(result, "DD", dayStr)

	// 替换时间占位符为默认值
	result = strings.ReplaceAll(result, "H", "00")
	result = strings.ReplaceAll(result, "h", "12")
	result = strings.ReplaceAll(result, "i", "00")
	result = strings.ReplaceAll(result, "s", "00")
	result = strings.ReplaceAll(result, "HH", "00")
	result = strings.ReplaceAll(result, "mm", "00")
	result = strings.ReplaceAll(result, "ss", "00")

	return result
}

// PHPTimeFormat 基于PHP时间占位符的Go格式化函数
// 支持PHP风格的时间格式占位符，如 Y-m-d H:i:s, YYYY-MM-DD 等
func PHPTimeFormat(t time.Time, format string) string {
	goFormat := convertPHPFormatToGo(format)
	return t.Format(goFormat)
}

// PHPTimeFormatNow 使用当前时间格式化
func PHPTimeFormatNow(format string) string {
	return PHPTimeFormat(time.Now(), format)
}

// PHPTimeFormatUnix 使用Unix时间戳格式化
func PHPTimeFormatUnix(timestamp int64, format string) string {
	t := time.Unix(timestamp, 0)
	return PHPTimeFormat(t, format)
}

// PHPTimeFormatString 解析时间字符串并格式化
func PHPTimeFormatString(timeStr, format string) (string, error) {
	// 尝试多种时间格式解析
	timeFormats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
	}

	var t time.Time
	var err error

	for _, timeFormat := range timeFormats {
		t, err = time.Parse(timeFormat, timeStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("无法解析时间字符串: %s", timeStr)
	}

	return PHPTimeFormat(t, format), nil
}

// PHPTimeFormatRelative 解析相对时间并格式化
func PHPTimeFormatRelative(relativeTime string, format string) (string, error) {
	t, err := ParseRelativeTime(relativeTime)
	if err != nil {
		return "", err
	}
	return PHPTimeFormat(t, format), nil
}

// parseRelativeTime 解析相对时间字符串
func ParseRelativeTime(timeStr string) (time.Time, error) {
	// 移除多余的空格
	timeStr = strings.TrimSpace(timeStr)

	// 处理常见的相对时间格式
	switch strings.ToLower(timeStr) {
	case "+1 day":
		return time.Now().AddDate(0, 0, 1), nil
	case "-1 day":
		return time.Now().AddDate(0, 0, -1), nil
	case "+1 week":
		return time.Now().AddDate(0, 0, 7), nil
	case "-1 week":
		return time.Now().AddDate(0, 0, -7), nil
	case "+1 month":
		return time.Now().AddDate(0, 1, 0), nil
	case "-1 month":
		return time.Now().AddDate(0, -1, 0), nil
	case "+1 year":
		return time.Now().AddDate(1, 0, 0), nil
	case "-1 year":
		return time.Now().AddDate(-1, 0, 0), nil
	case "0 days":
		return time.Now(), nil
	case "-0 days":
		return time.Now(), nil
	case "today":
		return time.Now(), nil
	case "yesterday":
		return time.Now().AddDate(0, 0, -1), nil
	default:
		// 尝试解析数字+单位的格式
		re := regexp.MustCompile(`^([+-]?\d+)\s*(day|days|week|weeks|month|months|year|years)$`)
		matches := re.FindStringSubmatch(timeStr)
		if len(matches) > 2 {
			num, err := strconv.Atoi(matches[1])
			if err != nil {
				return time.Time{}, err
			}

			switch matches[2] {
			case "day", "days":
				return time.Now().AddDate(0, 0, num), nil
			case "week", "weeks":
				return time.Now().AddDate(0, 0, num*7), nil
			case "month", "months":
				return time.Now().AddDate(0, num, 0), nil
			case "year", "years":
				return time.Now().AddDate(num, 0, 0), nil
			}
		}

		// 尝试直接解析时间格式
		if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
			return t, nil
		}
		if t, err := time.Parse("2006-01-02", timeStr); err == nil {
			return t, nil
		}

		return time.Time{}, fmt.Errorf("unable to parse time string: %s", timeStr)
	}
}

// convertPHPFormatToGo 将PHP日期格式转换为Go格式
func convertPHPFormatToGo(phpFormat string) string {
	// 常见的PHP到Go格式映射
	formatMap := map[string]string{
		"Y-m-d H:i:s": "2006-01-02 15:04:05",
		"Y-m-d":       "2006-01-02",
		"Y-m":         "2006-01",
		"H:i:s":       "15:04:05",
		"Y":           "2006",
		"m":           "01",
		"d":           "02",
		"H":           "15",
		"i":           "04",
		"s":           "05",
	}

	if goFormat, exists := formatMap[phpFormat]; exists {
		return goFormat
	}

	// 智能处理包含字面量字符的格式
	return convertCustomFormat(phpFormat)
}

// convertCustomFormat 智能转换自定义格式，区分格式占位符和字面量字符
func convertCustomFormat(format string) string {
	// 多字符格式占位符到Go格式的映射（按长度降序排列）
	multiCharFormats := []struct {
		pattern  string
		goFormat string
	}{
		{"YYYY", "2006"}, // 4位年份
		{"YY", "06"},     // 2位年份
		{"MM", "01"},     // 2位月份
		{"DD", "02"},     // 2位日期
		{"HH", "15"},     // 24小时制
		{"mm", "04"},     // 分钟
		{"ss", "05"},     // 秒
	}

	// 单字符格式占位符到Go格式的映射
	singleCharFormats := map[string]string{
		"Y": "2006", // 4位年份
		"y": "06",   // 2位年份
		"m": "01",   // 月份
		"d": "02",   // 日期
		"H": "15",   // 24小时制
		"h": "03",   // 12小时制
		"i": "04",   // 分钟
		"s": "05",   // 秒
		"A": "PM",   // AM/PM
		"a": "pm",   // am/pm
	}

	result := ""
	runes := []rune(format)
	i := 0
	for i < len(runes) {
		// 先尝试匹配多字符格式占位符
		matched := false
		for _, multiFormat := range multiCharFormats {
			patternRunes := []rune(multiFormat.pattern)
			if i+len(patternRunes) <= len(runes) {
				matchedPattern := true
				for j := 0; j < len(patternRunes); j++ {
					if runes[i+j] != patternRunes[j] {
						matchedPattern = false
						break
					}
				}
				if matchedPattern {
					result += multiFormat.goFormat
					i += len(patternRunes)
					matched = true
					break
				}
			}
		}

		if matched {
			continue
		}

		// 再尝试匹配单字符格式占位符
		char := string(runes[i])
		if goFormat, exists := singleCharFormats[char]; exists {
			result += goFormat
		} else {
			// 字面量字符，直接添加
			result += char
		}
		i++
	}

	return result
}
