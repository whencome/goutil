package timeutil

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/whencome/goutil"
)

// TimeRange 时间范围（unix时间，秒）
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

// Dates 获取时间范围内的日期列表
func (tr TimeRange) Dates() []string {
	dates := make([]string, 0)
	start := tr.StartTime
	for start.Before(tr.EndTime) {
		date := start.Format("2006-01-02")
		dates = append(dates, date)
		start = start.AddDate(0, 0, 1)
	}
	return dates
}

// GetLocation 获取系统使用的时区信息
func GetLocation() *time.Location {
	location, _ := time.LoadLocation("Asia/Shanghai")
	return location
}

// Now 获取当前时间（指定时区）
func Now() time.Time {
	loc := GetLocation()
	return time.Now().In(loc)
}

// FromUnixTime 将unix时间根据给定格式进行格式化
func FromUnixTime(timestamp int64, format string) string {
	if timestamp <= 0 {
		return ""
	}
	ut := time.Unix(timestamp, 0)
	return ut.Format(format)
}

// DateFromUnixTime 将unix时间戳转换为日期格式
func DateFromUnixTime(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	ut := time.Unix(timestamp, 0)
	return ut.Format("2006-01-02")
}

// DateTimeFromUnixTime 将unix时间戳转换为日期时间格式
func DateTimeFromUnixTime(timestamp int64) string {
	if timestamp <= 0 {
		return ""
	}
	ut := time.Unix(timestamp, 0)
	return ut.Format("2006-01-02 15:04:05")
}

// MillisecondToTime 将毫秒转换成时间格式
func MillisecondToTime(ms int64) string {
	tm := time.Unix(0, ms*int64(time.Millisecond))
	return tm.Format("2006-01-02 15:04:05.000")
}

// UnixTimeFromDateTime 将日期时间格式转换为unix时间戳
func UnixTimeFromDateTime(dt string) int64 {
	if dt == "" {
		return 0
	}
	// t, err := time.Parse("2006-01-02 15:04:05", dt)
	t, err := time.ParseInLocation("2006-01-02 15:04:05", dt, GetLocation())
	if err != nil {
		return 0
	}
	return t.Unix()
}

// UnixTime 将日期时间格式转换为unix时间戳
func UnixTime(dt, format string) int64 {
	if dt == "" {
		return 0
	}
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	t, err := time.ParseInLocation(format, dt, GetLocation())
	if err != nil {
		return 0
	}
	return t.Unix()
}

// MillisecondsFromDateTime 将日期时间转换为毫秒
func MillisecondsFromDateTime(dt, format string) int64 {
	if dt == "" {
		return 0
	}
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	t, err := time.ParseInLocation(format, dt, GetLocation())
	if err != nil {
		return 0
	}
	return t.UnixNano() / 1e6
}

// Unix 将日期时间格式转换为unix时间戳
func Unix(t string) int64 {
	format := GetFormat(t)
	if t == "" || format == "" {
		return 0
	}
	return UnixTime(t, format)
}

// GetFormat 获取日期时间戳格式
func GetFormat(t string) string {
	// 默认24小时制
	if m, e := regexp.MatchString(`\d{4}\-\d{1,2}\-\d{1,2}\s\d{1,2}:\d{1,2}:\d{1,2}`, t); e == nil && m {
		return "2006-01-02 15:04:05"
	}
	if m, e := regexp.MatchString(`\d{4}\-\d{1,2}\-\d{1,2}`, t); e == nil && m {
		return "2006-01-02"
	}
	if m, e := regexp.MatchString(`\d{1,2}\/\d{1,2}\/\d{4}`, t); e == nil && m {
		return "01/02/2006"
	}
	return ""
}

// Format2UnixTimestamp 将给定的数字转换成时间戳（保留到秒）,用于B端数据格式不统一的场景
func Format2UnixTimestamp(i int64) int64 {
	if i == 0 {
		return i
	}
	// 判断是时间戳还是日期时间格式
	location := GetLocation()
	ts := goutil.String(i)
	t, err := time.ParseInLocation("20060102150405", ts, location)
	// 如果出现错误，表示给定的数据是时间戳
	if err != nil {
		// 如果是毫秒，则转换到秒
		if len(ts) > 10 {
			ts = ts[0:10]
			return goutil.Int64(ts)
		}
		return i
	}
	return t.Unix()
}

// TimeFromUnixTimestamp 根据时间戳获取时间信息
func TimeFromUnixTimestamp(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// TodayBeginTime 获取当天开始时间(秒)
func TodayBeginTime() int64 {
	timeStr := time.Now().Format("2006-01-02")
	location := GetLocation()
	t, _ := time.ParseInLocation("2006-01-02", timeStr, location)
	return t.Unix()
}

// DateFromDateTime 从datetime中提取日期信息，仅支持"YYYY-MM-dd HH:mm:ss"格式
func DateFromDateTime(dt string) string {
	dt = strings.TrimSpace(dt)
	if dt == "" {
		return "0000-0-00"
	}
	spacePos := strings.Index(dt, " ")
	if spacePos > 0 {
		return dt[0:spacePos]
	}
	return dt
}

// GetDateBeginUnixTime 获取给定时间当日零点的Unix时间
func GetDateBeginUnixTime(t time.Time) int64 {
	zeroTime := GetDateBeginTime(t)
	return zeroTime.Unix()
}

// GetDateBeginTime 获取给定时间当日零点的时间
func GetDateBeginTime(t time.Time) time.Time {
	timeStr := t.Format("2006-01-02")
	zeroTime, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr+" 00:00:00", GetLocation())
	return zeroTime
}

// UnixTimeBeforeDays 获取指定天数前的unix时间，单位：秒
func UnixTimeBeforeDays(n int) int64 {
	return time.Now().Unix() - int64(n)*86400
}

// UnixTimeAfterDays 获取指定天数后的unix时间，单位：秒
func UnixTimeAfterDays(n int) int64 {
	return time.Now().Unix() + int64(n)*86400
}

// WeekdayFromUnixTime 从unix时间中获取weekday信息
func WeekdayFromUnixTime(timestamp int64) time.Weekday {
	t := time.Unix(timestamp, 0)
	return t.Weekday()
}

// FirstDayTime 计算给定时间作为第一天所在当天持续的时间（秒）
func FirstDayTime(timestamp int64) int64 {
	t := time.Unix(timestamp, 0)
	nextDayBegin := UnixTime(t.Format("2006-01-02"), "2006-01-02") + 86400
	return nextDayBegin - timestamp
}

// LastDayTime 计算给定时间作为最后一天所在当天持续的时间（秒）
func LastDayTime(timestamp int64) int64 {
	t := time.Unix(timestamp, 0)
	theDayBegin := UnixTime(t.Format("2006-01-02"), "2006-01-02")
	return timestamp - theDayBegin
}

// MonthTimeRange 计算一个月的时间范围，返回的时间范围是一个左闭右开区间（即不应包含右边的值）
func MonthTimeRange(yearMonth string) (TimeRange, error) {
	if m, e := regexp.MatchString(`^\d{4}\-\d{1,2}$`, yearMonth); e != nil || !m {
		return TimeRange{}, errors.New("invalid month format")
	}
	parts := strings.Split(yearMonth, "-")
	year := goutil.Int(parts[0])
	month := goutil.Int(parts[1])
	loc := GetLocation()
	beginTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	endTime := beginTime.AddDate(0, +1, 0)
	return TimeRange{
		StartTime: beginTime,
		EndTime:   endTime,
	}, nil
}

// SplitTimeRangeByDay 将给定的时间按自然天进行拆分
func SplitTimeRangeByDay(beginTime, endTime int64) []TimeRange {
	return SplitTimeRangeByDays(beginTime, endTime, 1)
}

// SplitTimeRangeByDays 将给定的时间按固定的天数进行拆分
func SplitTimeRangeByDays(beginTime, endTime int64, days int) []TimeRange {
	begin := TimeFromUnixTimestamp(beginTime)
	end := TimeFromUnixTimestamp(endTime)
	ranges := make([]TimeRange, 0)
	// 如果days小于或者等于0， 则将参数作为一个时间范围返回
	if days <= 0 {
		ranges = append(ranges, TimeRange{
			StartTime: begin,
			EndTime:   end,
		})
		return ranges
	}
	goOn := true
	start := begin
	for goOn {
		if start.After(end) {
			break
		}
		startTime := GetDateBeginTime(start)
		endTime := startTime.AddDate(0, 0, days)
		if startTime.Before(start) {
			startTime = start
		}
		if endTime.After(end) || endTime.Equal(end) {
			goOn = false
			endTime = end
		}
		dateRange := TimeRange{
			StartTime: startTime,
			EndTime:   endTime,
		}
		ranges = append(ranges, dateRange)
		// 重置开始时间
		start = endTime
	}
	return ranges
}

// SplitTimeRange 根据给定的函数进行拆分
func SplitTimeRange(beginTime, endTime int64, f func(startTime, endTime time.Time) (TimeRange, bool)) []TimeRange {
	begin := TimeFromUnixTimestamp(beginTime)
	end := TimeFromUnixTimestamp(endTime)
	ranges := make([]TimeRange, 0)
	goOn := true
	start := begin
	for goOn {
		if start.After(end) || start.Equal(end) {
			break
		}
		timeRange, hasNext := f(start, end)
		if !hasNext {
			goOn = false
		}
		start = timeRange.EndTime
		ranges = append(ranges, timeRange)
	}
	return ranges
}

// GetInterval 计算两个时间间隔
func GetInterval(startTime, endTime int64, includeWeekend bool) int64 {
	if endTime <= startTime {
		return 0
	}
	if includeWeekend {
		return endTime - startTime
	}
	// 计算不包含周末的时间
	interval := endTime - startTime
	aWeekTime := int64(86400 * 7)
	weeks := interval / aWeekTime
	if weeks > 0 {
		interval -= weeks * 86400 * 2
	}
	leftStartTime := startTime + weeks*aWeekTime
	// 将除去整周时间后的时间单独计算
	startWeekday := WeekdayFromUnixTime(leftStartTime)
	endWeekday := WeekdayFromUnixTime(endTime)
	if startWeekday > time.Sunday && endWeekday < time.Saturday {
		return interval
	}
	// 剩余的时间是在一周之内，且小于一周
	// 如果开始日期工作日与结束日期工作日在同一天，则只需判断是否是在周末
	if startWeekday == endWeekday {
		if startWeekday == time.Sunday || startWeekday == time.Saturday {
			// 两个时间在周末
			interval -= endTime - leftStartTime
		}
		// 表示时间在工作日
		return interval
	}

	// 只需要处理周末即可
	if startWeekday < endWeekday {
		if startWeekday == time.Sunday {
			interval -= FirstDayTime(leftStartTime)
		}
		if endWeekday == time.Saturday {
			interval -= LastDayTime(endTime)
		}
		return interval
	}

	// 有可能跨越了一个周末
	if startWeekday > endWeekday {
		// startWeekday不可能是星期天（0）
		if startWeekday < time.Saturday {
			// 跨越一个完整的周末
			if endWeekday > time.Sunday {
				interval -= 86400 * 2
			} else {
				// endWeekday == time.Sunday
				interval -= 86400 + LastDayTime(endTime)
			}
			return interval
		}
		// startWeekday = 6 (星期六)
		if endWeekday > time.Sunday {
			interval -= 86400 + FirstDayTime(leftStartTime)
		} else {
			// endWeekday == time.Sunday
			interval -= FirstDayTime(leftStartTime) + LastDayTime(endTime)
		}
		return interval
	}
	return interval
}
