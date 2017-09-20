// Package weekly provides functionality for handling events that are periodic
// over the course of a week.
package weekly

import (
	"container/heap"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"sort"
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

type ticker struct {
	spec TickSpecification
	nxt  time.Time
}

// tickerHeap implements heap.Interface.
type tickerHeap []*ticker

func (h tickerHeap) Len() int            { return len(h) }
func (h tickerHeap) Less(i, j int) bool  { return h[i].nxt.Before(h[j].nxt) }
func (h tickerHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *tickerHeap) Push(x interface{}) { *h = append(*h, x.(*ticker)) }
func (h *tickerHeap) Pop() interface{} {
	var val *ticker
	*h, val = (*h)[:len(*h)-1], (*h)[len(*h)-1]
	return val
}

// NewTicker returns a ticker that starts and stops ticking at the same time each week.
func NewTicker(tickSpecs []TickSpecification) (*Ticker, error) {
	// Create heap of tickers based on tick specifications.
	var tickers tickerHeap
	now := time.Now()
	if len(tickSpecs) == 0 {
		return nil, errors.New("no tick specifications")
	}
	for _, ts := range tickSpecs {
		if ts.End.Before(ts.Start) {
			return nil, errors.New("end is before start")
		}
		if ts.Frequency <= 0 {
			return nil, errors.New("freq is nonpositive")
		}
		tickers = append(tickers, &ticker{
			spec: ts,
			nxt:  nextTick(now, ts),
		})
	}
	sort.Sort(tickers)
	for i := 1; i < len(tickers); i++ {
		if tickers[i].spec.Start.Before(tickers[i-1].spec.End) {
			return nil, errors.New("tick specifications overlap")
		}
	}
	heap.Init(&tickers)

	// Set up RNG.
	var buf [8]byte
	if _, err := crand.Read(buf[:]); err != nil {
		return nil, fmt.Errorf("could not seed RNG: %v", err)
	}
	seed := int64(binary.LittleEndian.Uint64(buf[:]))
	rnd := rand.New(rand.NewSource(seed))

	// Create the last few variables, start ticking, and return channel to user.
	ch := make(chan time.Time)
	done := make(chan struct{})
	go tick(ch, done, rnd, tickers)
	return &Ticker{
		C:    ch,
		done: done,
	}, nil
}

func tick(ch chan<- time.Time, done chan struct{}, rnd *rand.Rand, tickers tickerHeap) {
	for {
		// Go to sleep until we reach the next tick.
		ticker := tickers[0]
		nxt := ticker.nxt.Add(time.Duration(rnd.Float64() * float64(ticker.spec.Frequency)))
		tmr := time.NewTimer(time.Until(nxt))
		select {
		case <-tmr.C:
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

		ticker.nxt = nextTick(ticker.nxt, ticker.spec)
		heap.Fix(&tickers, 0)
	}
}

func nextTick(tck time.Time, spec TickSpecification) time.Time {
	s, e := spec.Start.InWeek(tck), spec.End.InWeek(tck)
	switch {
	case tck.Before(s):
		// We haven't started ticking yet this week.
		return s

	case tck.Before(e):
		// We are currently ticking. Figure out the next tick from when we are.
		nxt := s.Add(spec.Frequency * (1 + (tck.Sub(s) / spec.Frequency)))
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
