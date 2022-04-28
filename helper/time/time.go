package time

import (
	"time"
)

// Time PHP获取当前时间戳
func Time() int64 {
	return time.Now().Unix()
}

// Date PHP格式化时间 timestamp=0 默认为当前时间
func Date(timestamp int64, timeLayout string) string {
	if timestamp == 0 {
		timestamp = Time()
	}
	return time.Unix(timestamp, 0).Format(timeLayout)
}

// StrToTime PHP时间字符串转时间戳
func StrToTime(datetime, timeLayout string) int64 {
	times := NewDateTime(datetime, timeLayout)
	return times.Unix()
}

// Diff 比较两个格式一样的时间的时间差
func Diff(time1, time2 time.Time) (year, month, day, hour, minute, second int) {
	var local *time.Location
	local = time.Local
	if time1.After(time2) {
		time1, time2 = time2, time1
	}
	y1, m1, d1 := time1.Date()
	h1, i1, s1 := time1.Clock()

	y2, m2, d2 := time2.Date()
	h2, i2, s2 := time2.Clock()

	year = y2 - y1
	month = int(m2 - m1)
	day = d2 - d1
	hour = h2 - h1
	minute = i2 - i1
	second = s2 - s1

	if second < 0 {
		second += 60
		minute--
	}
	if minute < 0 {
		minute += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		t := time.Date(y2, m2, 0, 0, 0, 0, 0, local)
		day += t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}
	return
}

// NewDateTime 创建一个时间Time
func NewDateTime(dateTime, timeLayout string) time.Time {
	times, _ := time.ParseInLocation(timeLayout, dateTime, time.Local)
	return times
}
