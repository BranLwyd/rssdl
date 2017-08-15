// Package week provides functionality for handling events that are periodic
// over the course of a week.
package week

import (
	"errors"
	"fmt"
	"time"
)

var (
	// Used by Parse.
	dayMap = map[string]time.Weekday{
		"Sun ": time.Sunday,
		"Mon ": time.Monday,
		"Tue ": time.Tuesday,
		"Wed ": time.Wednesday,
		"Thu ": time.Thursday,
		"Fri ": time.Friday,
		"Sat ": time.Saturday,
	}
)

// Time represents a specific time during a week; weeks start on Sunday and
// go through the following Saturday. A week.Time value represents an instant
// in time in every week, and may be converted to a specific instant in a
// specific week.
type Time struct {
	day                  time.Weekday
	hour, min, sec, nsec int
}

// FromTime gives a week.Time corresponding to the offset of the given
// time.Time from the beginning of the week that it falls into, in the location
// the time.Time is in.
func FromTime(t time.Time) Time {
	return Time{day: t.Weekday(), hour: t.Hour(), min: t.Minute(), sec: t.Second(), nsec: t.Nanosecond()}
}

// Parse parses a string value into a time during the week. The expected format
// is like: "Thu 7:30PM". The local location is used.
func Parse(val string) (Time, error) {
	if len(val) < 4 {
		return Time{}, errors.New("bad weekday")
	}
	day, ok := dayMap[val[:4]]
	if !ok {
		return Time{}, errors.New("bad weekday")
	}
	tt, err := time.Parse(time.Kitchen, val[4:])
	if err != nil {
		return Time{}, fmt.Errorf("bad time: %v", err)
	}
	wt := FromTime(tt)
	wt.day = day
	return wt, nil
}

// InWeek converts a given week.Time to a time.Time in the same week as the given time.Time.
func (wt Time) InWeek(tt time.Time) time.Time {
	return time.Date(tt.Year(), tt.Month(), tt.Day()+int(wt.day)-int(tt.Weekday()), wt.hour, wt.min, wt.sec, wt.nsec, tt.Location())
}

// InSameWeek determines if two time.Time values fall into the same week.
// Two time.Time values can be in the same week only if they are in the same location.
func InSameWeek(t1, t2 time.Time) bool {
	t1 = t1.AddDate(0, 0, -int(t1.Weekday()))
	t2 = t2.AddDate(0, 0, -int(t2.Weekday()))
	// TODO(bran): Location() returns a pointer; any possibility of false negatives?
	return t1.Location() == t2.Location() &&
		t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()
}
