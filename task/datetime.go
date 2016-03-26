package task

import "time"

const (
	dateFormat = "2006-01-02"
	timeFormat = "15:04"
)

func ParseDate(s string) (time.Time, error) {
	return time.Parse(dateFormat, s)
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}
