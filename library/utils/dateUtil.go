package utils

import (
	"collection-center/internal/logger"
	"github.com/go-errors/errors"
	"strconv"
	"strings"
	"time"
)

const (
	FormatDateTime   = "2006-01-02 15:04:05"
	FormatDate       = "2006-01-02"
	FormatDate1      = "20060102"
	FormatDateNoTime = "2006-01-02 00:00:00"
	FormatDateTime2  = "2006/01/02 15:04:05"
	FormatDateForUtc = "2006-01-02T15:04:05"
)

// 获取当前时间当前月第一天日期
func GetMonthFirstDay(now time.Time) time.Time {
	currentYear, currentMonth, _ := now.Date()
	location, err := time.LoadLocation("Local")
	if err != nil {
		logger.Error("LoadLocation err:", err)
		return time.Time{}
	}
	//currentLocation := err
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, location)
	return firstOfMonth
}

// 获取当前日期不含时间time
func GetDayNoTime(now time.Time) time.Time {
	if now.IsZero() {
		return time.Time{}
	}
	format := now.Format("2006-01-02 00:00:00")
	parse, _ := time.Parse("2006-01-02 00:00:00", format)
	return parse
}

func GetYearMonthStr(now time.Time) string {
	if now.IsZero() {
		return ""
	}
	format := now.Format("200601")
	return format
}

// 获取当前年
func GetCurrentYear() int64 {
	return int64(time.Now().Year())
}

// str转time
func ParseTimeByTimeStr(str, Prefix string) (time.Time, error) {
	p := strings.TrimSpace(str)
	if p == "" {
		return time.Time{}, errors.Errorf("%s不能为空", str)
	}

	t, err := time.ParseInLocation(Prefix, str, time.Local)
	if err != nil {
		return time.Time{}, errors.Errorf("%s格式错误", Prefix)
	}
	return t, nil
}

func AddOneDay(date time.Time) time.Time {
	duration, _ := time.ParseDuration("24h")
	return date.Add(duration)
}
func AddAnyDay(date time.Time, dayNum int) time.Time {
	i := 24 * dayNum
	hourStr := strconv.Itoa(i)
	duration, _ := time.ParseDuration(hourStr + "h")
	return date.Add(duration)
}
func AddMonth(date time.Time, monthNum int) time.Time {
	return date.AddDate(0, monthNum, 0)
}
func AddYear(date time.Time, yearNum int) time.Time {
	return date.AddDate(yearNum, 0, 0)
}

func TimeToString(time time.Time, format string) string {
	return time.Format(format)
}

func AddHours(date time.Time, hour int64) time.Time {
	duration, _ := time.ParseDuration(strconv.FormatInt(hour, 10) + "h")
	return date.Add(duration)
}
