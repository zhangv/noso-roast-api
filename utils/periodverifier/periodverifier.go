package periodverifier

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeFieldMatcher 定义时间字段的匹配规则
type TimeFieldMatcher struct {
	IsAny    bool  // 是否为 '*'
	Start    int   // 范围开始 (也用于单个数值)
	End      int   // 范围结束 (也用于单个数值)
	Specific []int // 具体列表
}

// Matches 判断给定的值是否符合匹配规则
func (tfm *TimeFieldMatcher) Matches(value int) bool {
	if tfm.IsAny {
		return true
	}
	if len(tfm.Specific) > 0 { // 优先匹配列表
		for _, s := range tfm.Specific {
			if s == value {
				return true
			}
		}
		return false
	}
	// 匹配范围或单个数值 (单个数值时 Start == End)
	return value >= tfm.Start && value <= tfm.End
}

// PeriodVerifier 类结构
type PeriodVerifier struct {
	Year        TimeFieldMatcher
	Month       TimeFieldMatcher
	Day         TimeFieldMatcher
	Weekday     TimeFieldMatcher // 0-6 (周日-周六)
	Hour        TimeFieldMatcher
	Minute      TimeFieldMatcher
	Second      TimeFieldMatcher
	Occurrences int // -1 表示无限次，其他值表示最大触发次数
}

// NewPeriodVerifierFromExpression 解析表达式并返回 PeriodVerifier 实例
func NewPeriodVerifierFromExpression(expression string) (*PeriodVerifier, error) {
	parts := strings.SplitN(expression, ",", 2)
	periodStr := parts[0]
	occurrences := -1 // 默认无限次

	if len(parts) == 2 {
		occ, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("无效的次数: %s", parts[1])
		}
		occurrences = occ
	}

	fields := strings.Split(periodStr, ":")

	// 期望 7 个字段: 年:月:日:星期:时:分:秒
	if len(fields) != 7 {
		return nil, fmt.Errorf("表达式需要 7 个时间字段 (年:月:日:星期:时:分:秒), 实际 %d 个: %s", len(fields), periodStr)
	}

	pv := &PeriodVerifier{Occurrences: occurrences}
	var err error

	// 解析年 (例如 2025, *) - 范围可根据需要调整
	pv.Year, err = parseTimeField(fields[0], 1900, 2100)
	if err != nil {
		return nil, fmt.Errorf("解析年份失败: %w", err)
	}
	// 解析月 (1-12)
	pv.Month, err = parseTimeField(fields[1], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("解析月份失败: %w", err)
	}
	// 解析日 (1-31)
	pv.Day, err = parseTimeField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("解析日期失败: %w", err)
	}
	// 解析星期 (0-6, 周日-周六)
	pv.Weekday, err = parseTimeField(fields[3], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("解析星期失败: %w", err)
	}
	// 解析时 (0-23)
	pv.Hour, err = parseTimeField(fields[4], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("解析小时失败: %w", err)
	}
	// 解析分 (0-59)
	pv.Minute, err = parseTimeField(fields[5], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("解析分钟失败: %w", err)
	}
	// 解析秒 (0-59)
	pv.Second, err = parseTimeField(fields[6], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("解析秒失败: %w", err)
	}

	return pv, nil
}

// parseTimeField 辅助函数，解析单个时间字段的规则
func parseTimeField(fieldStr string, min, max int) (TimeFieldMatcher, error) {
	tfm := TimeFieldMatcher{}
	if fieldStr == "*" {
		tfm.IsAny = true
		return tfm, nil
	}

	if strings.Contains(fieldStr, "-") {
		parts := strings.Split(fieldStr, "-")
		if len(parts) != 2 {
			return tfm, fmt.Errorf("无效的范围格式 '%s', 期望 'start-end'", fieldStr)
		}
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return tfm, fmt.Errorf("无效的范围开始值: %s", parts[0])
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return tfm, fmt.Errorf("无效的范围结束值: %s", parts[1])
		}
		if start < min || start > max || end < min || end > max || start > end {
			return tfm, fmt.Errorf("范围值 '%s' 无效, 期望 %d-%d 且 start <= end", fieldStr, min, max)
		}
		tfm.Start = start
		tfm.End = end
		return tfm, nil
	}

	if strings.Contains(fieldStr, ",") {
		specificVals := strings.Split(fieldStr, ",")
		for _, valStr := range specificVals {
			val, err := strconv.Atoi(valStr)
			if err != nil {
				return tfm, fmt.Errorf("无效的列表值: %s", valStr)
			}
			if val < min || val > max {
				return tfm, fmt.Errorf("列表值 '%d' 无效, 期望 %d-%d", val, min, max)
			}
			tfm.Specific = append(tfm.Specific, val)
		}
		return tfm, nil
	}

	// 单个数值
	val, err := strconv.Atoi(fieldStr)
	if err != nil {
		return tfm, fmt.Errorf("无效的单个数值: %s", fieldStr)
	}
	if val < min || val > max {
		return tfm, fmt.Errorf("单个数值 '%d' 无效, 期望 %d-%d", val, min, max)
	}
	tfm.Start = val // 对于单个值，start 和 end 都设置为它
	tfm.End = val
	return tfm, nil
}

/*
根据当前的规则字符串，生成最近的一个时间点或时间段

	参数：
	1. 当前时间
	2. 当前触发次数
	返回值：
	1. 是否符合规则
	2. 符合规则的时间段（当前）
	3. 符合规则的时间段（上一个）
	4. 符合规则的时间段（下一个）
	5. 错误信息
*/
type Period struct {
	StartTime time.Time
	EndTime   time.Time
}

/*
NextPeriod 返回下一个符合规则的时间段

参数：
1. 当前时间
返回值：
1. 是否符合规则
2. 符合规则的时间段（当前）
3. 符合规则的时间段（上一个）
4. 符合规则的时间段（下一个）
5. 错误信息
*/
func (pv *PeriodVerifier) NextPeriod(t time.Time) (bool, Period, Period, Period, error) {
	// 1. 如果当前时间符合规则，则返回当前时间 - 如果时间[点]
	if pv.IsMatch(t) {
		return true, Period{StartTime: t, EndTime: t}, Period{}, Period{}, nil
	}
	// 2. 如果当前时间不符合规则，则返回当前的所在时间段（开始时间、结束时间）
	matched := false
	cur := Period{StartTime: t, EndTime: t}
	prev := Period{StartTime: t, EndTime: t}
	next := Period{StartTime: t, EndTime: t}
	if pv.Second.Matches(t.Second()) {
		matched = true
		if pv.Second.IsAny {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, 0, t.Location())
		} else {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
		}
	}
	if pv.Minute.Matches(t.Minute()) {
		matched = true
		if pv.Minute.IsAny {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, cur.EndTime.Second(), 0, t.Location())
		} else {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), cur.EndTime.Second(), 0, t.Location())
		}
	}
	if pv.Hour.Matches(t.Hour()) {
		matched = true
		if pv.Hour.IsAny {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), 0, cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), 23, cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
		} else {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
		}
	}
	if pv.Day.Matches(t.Day()) {
		if pv.Weekday.Matches(int(t.Weekday())) {
			matched = true
			if pv.Day.IsAny {
				cur.StartTime = time.Date(t.Year(), t.Month(), 1, cur.StartTime.Hour(), cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
				cur.EndTime = time.Date(t.Year(), t.Month(), 31, cur.EndTime.Hour(), cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
			} else {
				cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), cur.StartTime.Hour(), cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
				cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), cur.EndTime.Hour(), cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
			}
		} else { //如果日期match但是星期不match，则需要返回下一个日期和星期都match的时间点（从t开始）
			cur.StartTime, cur.EndTime = time.Time{}, time.Time{}
			next, _ = nextMatch_Date_Weekday(pv.Day, pv.Weekday, t)
			prev, _ = prevMatch_Date_Weekday(pv.Day, pv.Weekday, t)
		}
	}
	if pv.Month.Matches(int(t.Month())) {
		matched = true
		if pv.Month.IsAny {
			cur.StartTime = time.Date(t.Year(), 1, cur.StartTime.Day(), cur.StartTime.Hour(), cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), 12, cur.EndTime.Day(), cur.EndTime.Hour(), cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
		} else {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), cur.StartTime.Hour(), cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), cur.EndTime.Hour(), cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
		}
	}
	if pv.Year.Matches(t.Year()) {
		matched = true
		if pv.Year.IsAny {
			cur.StartTime = time.Date(1900, cur.StartTime.Month(), cur.StartTime.Day(), cur.StartTime.Hour(), cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(2100, cur.EndTime.Month(), cur.EndTime.Day(), cur.EndTime.Hour(), cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
		} else {
			cur.StartTime = time.Date(t.Year(), t.Month(), t.Day(), cur.StartTime.Hour(), cur.StartTime.Minute(), cur.StartTime.Second(), 0, t.Location())
			cur.EndTime = time.Date(t.Year(), t.Month(), t.Day(), cur.EndTime.Hour(), cur.EndTime.Minute(), cur.EndTime.Second(), 0, t.Location())
		}
	}
	return matched, cur, prev, next, nil
}

func nextMatch_Date_Weekday(day TimeFieldMatcher, weekday TimeFieldMatcher, t time.Time) (Period, error) {
	nextDay := t
	i := 0
	for i < 10000 {
		if day.Matches(nextDay.Day()) && weekday.Matches(int(nextDay.Weekday())) {
			startTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())
			endTime := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 23, 59, 59, 0, nextDay.Location())
			return Period{StartTime: startTime, EndTime: endTime}, nil
		} else {
			nextDay = nextDay.AddDate(0, 0, 1)
			i++
		}
	}
	return Period{}, fmt.Errorf("找不到符合规则的时间点")
}
func prevMatch_Date_Weekday(day TimeFieldMatcher, weekday TimeFieldMatcher, t time.Time) (Period, error) {
	prevDay := t
	i := 0
	for i < 10000 {
		if day.Matches(prevDay.Day()) && weekday.Matches(int(prevDay.Weekday())) {
			startTime := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 0, 0, 0, 0, prevDay.Location())
			endTime := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 23, 59, 59, 0, prevDay.Location())
			return Period{StartTime: startTime, EndTime: endTime}, nil
		} else {
			prevDay = prevDay.AddDate(0, 0, -1)
			i++
		}
	}
	return Period{}, fmt.Errorf("找不到符合规则的时间点")
}

// IsMatch 验证给定的时间点是否符合 PeriodVerifier 定义的规则
func (pv *PeriodVerifier) IsMatch(t time.Time) bool {
	// 匹配每个时间字段
	if !pv.Year.Matches(t.Year()) {
		return false
	}
	if !pv.Month.Matches(int(t.Month())) {
		return false
	}
	if !pv.Day.Matches(t.Day()) {
		return false
	}
	if !pv.Weekday.Matches(int(t.Weekday())) { // time.Weekday() 返回 0 (周日) 到 6 (周六)
		return false
	}
	if !pv.Hour.Matches(t.Hour()) {
		return false
	}
	if !pv.Minute.Matches(t.Minute()) {
		return false
	}
	if !pv.Second.Matches(t.Second()) {
		return false
	}
	return true // 所有字段都匹配
}
