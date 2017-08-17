// Package weekly provides functionality for handling events that are periodic
// over the course of a week.
package weekly

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

// A Ticker holds a channel that delivers ticks of a clock at intervals.
// It starts & stops ticking at the same time each week.
type Ticker struct {
	C    <-chan time.Time
	done chan struct{}
}

// Stop closes the ticker and releases any resources it has acquired.
func (t *Ticker) Stop() {
	close(t.done)
}

// NewTicker returns a ticker that starts and stops ticking at the same time each week.
func NewTicker(start, end Time, freq time.Duration) *Ticker {
	ch := make(chan time.Time)
	done := make(chan struct{})
	go tick(ch, done, start, end, freq)
	return &Ticker{
		C:    ch,
		done: done,
	}
}

func tick(ch chan<- time.Time, done chan struct{}, start, end Time, freq time.Duration) {
	now := time.Now()
	for {
		// Figure out when the next tick is.
		var nextTick time.Time

		s, e := start.InWeek(now), end.InWeek(now)
		switch {
		case now.Before(s):
			// We haven't started ticking yet this week.
			nextTick = s

		case now.Before(e):
			// We are currently ticking. Figure out the next tick from when we are.
			nextTick = s.Add(freq * (1 + (now.Sub(s) / freq)))
			if nextTick.Before(e) {
				break
			}
			// The next tick is after the end of the ticking interval. We're done ticking this week.
			fallthrough

		default:
			// We are done ticking this week. Wait until we start ticking next week.
			nextTick = s.AddDate(0, 0, 7)
		}

		// Go to sleep until we are ready to tick.
		tmr := time.NewTimer(nextTick.Sub(now))
		select {
		case t := <-tmr.C:
			// Send the tick from the timer in case it was delayed.
			// Drop the tick if no one is ready to receive it.
			select {
			case ch <- t:
			default:
			}

		case <-done:
			if !tmr.Stop() {
				<-tmr.C
			}
			return
		}

		// Set now to be the tick time.
		// Use the computed time instead of the time from the timer in order to maintain determinism.
		now = nextTick
	}
}

// Time represents a specific time during a week; weeks start on Sunday and go
// through the following Saturday. A weekly.Time value represents an instant in
// time in every week, and may be converted to a specific instant in a specific
// week.
type Time struct {
	day       time.Weekday
	hour, min int
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
	t, err := time.Parse(time.Kitchen, val[4:])
	if err != nil {
		return Time{}, fmt.Errorf("bad time: %v", err)
	}
	return Time{day: day, hour: t.Hour(), min: t.Minute()}, nil
}

// InWeek converts a given weekly.Time to a time.Time in the same week as the
// given time.Time.
func (wt Time) InWeek(tt time.Time) time.Time {
	return time.Date(tt.Year(), tt.Month(), tt.Day()+int(wt.day)-int(tt.Weekday()), wt.hour, wt.min, 0, 0, tt.Location())
}
