// Package weekly provides functionality for handling events that are periodic
// over the course of a week.
package weekly

import (
	"errors"
	"fmt"
	"time"
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

// TickSpecification is used with NewTicker. It specifies a period each week
// when ticks occur, and how frequently ticks occur during that period.
type TickSpecification struct {
	Start, End Time          // when to start and stop ticking each week
	Frequency  time.Duration // how often to tick while ticking
}

// NewTicker returns a ticker that starts and stops ticking at the same time each week.
func NewTicker(tickSpecs []TickSpecification) (*Ticker, error) {
	for _, ts := range tickSpecs {
		if ts.End.Before(ts.Start) {
			return nil, errors.New("end is before start")
		}
		if ts.Frequency <= 0 {
			return nil, errors.New("freq is nonpositive")
		}
	}

	ch := make(chan time.Time)
	done := make(chan struct{})
	for _, ts := range tickSpecs {
		go tick(ch, done, ts.Start, ts.End, ts.Frequency)
	}
	return &Ticker{
		C:    ch,
		done: done,
	}, nil
}

func tick(ch chan<- time.Time, done chan struct{}, start, end Time, freq time.Duration) {
	tck := time.Now()
	for {
		// Go to sleep until we reach the next tick.
		nxt := nextTick(tck, start, end, freq)
		tmr := time.NewTimer(time.Until(nxt))
		select {
		case <-tmr.C:
			// Send the tick from the  in case it was delayed.
			// Drop the tick if it is not ready to be received.
			select {
			case ch <- nxt:
			default:
			}

		case <-done:
			if !tmr.Stop() {
				<-tmr.C
			}
			return
		}
		tck = nxt
	}
}

func nextTick(tck time.Time, start, end Time, freq time.Duration) (nxt time.Time) {
	s, e := start.InWeek(tck), end.InWeek(tck)
	switch {
	case tck.Before(s):
		// We haven't started ticking yet this week.
		return s

	case tck.Before(e):
		// We are currently ticking. Figure out the next tick from when we are.
		nxt := s.Add(freq * (1 + (tck.Sub(s) / freq)))
		if nxt.Before(e) {
			return nxt
		}
		// The next tick is after the end of the ticking interval. We're done ticking this week.
		fallthrough

	default:
		// We are done ticking this week. Wait until we start ticking next week.
		return s.AddDate(0, 0, 7)
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
	day, ok := strToDay[val[:4]]
	if !ok {
		return Time{}, errors.New("bad weekday")
	}
	t, err := time.Parse(time.Kitchen, val[4:])
	if err != nil {
		return Time{}, fmt.Errorf("bad time: %v", err)
	}
	return Time{day: day, hour: t.Hour(), min: t.Minute()}, nil
}

func MustParse(val string) Time {
	t, err := Parse(val)
	if err != nil {
		panic(fmt.Sprintf("Parse(%q): %v", val, err))
	}
	return t
}

// InWeek converts a given weekly.Time to a time.Time in the same week as the
// given time.Time.
func (wt Time) InWeek(tt time.Time) time.Time {
	return time.Date(tt.Year(), tt.Month(), tt.Day()+int(wt.day)-int(tt.Weekday()), wt.hour, wt.min, 0, 0, tt.Location())
}

func (wt Time) Before(owt Time) bool {
	return wt.day < owt.day ||
		(wt.day == owt.day && wt.hour < owt.hour) ||
		(wt.day == owt.day && wt.hour == owt.hour && wt.min < owt.min)
}

func (wt Time) String() string {
	ampm := "AM"
	if wt.hour >= 12 {
		ampm = "PM"
	}
	mhr := wt.hour % 12
	if mhr == 0 {
		mhr = 12
	}
	return fmt.Sprintf("%s %d:%02d%s", dayToStr[wt.day], mhr, wt.min, ampm)
}

var (
	// Used by Parse.
	strToDay = map[string]time.Weekday{
		"Sun ": time.Sunday,
		"Mon ": time.Monday,
		"Tue ": time.Tuesday,
		"Wed ": time.Wednesday,
		"Thu ": time.Thursday,
		"Fri ": time.Friday,
		"Sat ": time.Saturday,
	}

	// Used by String.
	dayToStr = map[time.Weekday]string{
		time.Sunday:    "Sun",
		time.Monday:    "Mon",
		time.Tuesday:   "Tue",
		time.Wednesday: "Wed",
		time.Thursday:  "Thu",
		time.Friday:    "Fri",
		time.Saturday:  "Sat",
	}
)
