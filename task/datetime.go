package task

import "time"

const (
	dateFormat     = "2006-01-02"
	timeFormat     = "15:04"
	dateTimeFormat = dateFormat + " " + timeFormat
)

func ParseDate(s string) (time.Time, error) {
	return time.Parse(dateFormat, s)
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse(timeFormat, s)
}

func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateFormat)
}

func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(timeFormat)
}

func FormatDateTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateTimeFormat)
}
