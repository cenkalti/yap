package datetime

import (
	"errors"
	"time"
)

const (
	dateLayout      = "2006-01-02"
	timeLayout      = "15:04"
	dateTimeLayout  = "2006-01-02 15:04"
	dateTimeLayoutT = "2006-01-02T15:04"
)

var dayReset = time.Date(0, 0, 0, 4, 0, 0, 0, time.Local)

type DateTime struct {
	time.Time
	HasTime bool
}

func New(t time.Time) DateTime {
	return DateTime{
		Time:    t,
		HasTime: true,
	}
}

func Parse(value string) (dt DateTime, err error) {
	switch len(value) {
	case len(dateLayout):
		dt.Time, err = time.Parse(dateLayout, value)
		dt.Time = replaceTime(dt.Time, dayReset)
	case len(timeLayout):
		dt.HasTime = true
		dt.Time, err = time.Parse(timeLayout, value)
		dt.Time = replaceDate(dt.Time, time.Now())
	case len(dateTimeLayoutT):
		dt.HasTime = true
		dt.Time, err = time.Parse(dateTimeLayoutT, value)
	default:
		err = errors.New("datetime: invalid value")
	}
	return
}

func replaceDate(t, td time.Time) time.Time {
	return time.Date(td.Year(), td.Month(), td.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
}

func replaceTime(t, tt time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), tt.Hour(), tt.Minute(), tt.Second(), tt.Nanosecond(), time.Local)
}

func (dt DateTime) String() string {
	if dt.HasTime {
		return dt.Format(dateTimeLayout)
	}
	return dt.Format(dateLayout)
}
