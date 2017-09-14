package weekly

import (
	"fmt"
	"testing"
	"time"
)

func TestNextTick(t *testing.T) {
	t.Parallel()

	// 2017-08-20 is a Sunday.
	for _, test := range []struct {
		desc       string
		t          time.Time
		start, end Time
		freq       time.Duration
		want       time.Time
	}{
		{
			desc:  "before_interval",
			t:     time.Date(2017, 8, 21, 15, 23, 11, 423691, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 23, 17, 30, 0, 0, time.UTC),
		},
		{
			desc:  "at_first_tick",
			t:     time.Date(2017, 8, 23, 17, 30, 0, 0, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 23, 17, 31, 0, 0, time.UTC),
		},
		{
			desc:  "after_first_tick",
			t:     time.Date(2017, 8, 23, 17, 30, 22, 0, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 23, 17, 31, 0, 0, time.UTC),
		},
		{
			desc:  "at_inner_tick",
			t:     time.Date(2017, 8, 24, 4, 22, 0, 0, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 24, 4, 23, 0, 0, time.UTC),
		},
		{
			desc:  "after_inner_tick",
			t:     time.Date(2017, 8, 24, 4, 22, 36, 0, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 24, 4, 23, 0, 0, time.UTC),
		},
		{
			desc:  "at_last_tick",
			t:     time.Date(2017, 8, 24, 5, 29, 0, 0, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 30, 17, 30, 0, 0, time.UTC),
		},
		{
			desc:  "after_last_tick",
			t:     time.Date(2017, 8, 24, 5, 29, 47, 0, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 30, 17, 30, 0, 0, time.UTC),
		},
		{
			desc:  "after_interval",
			t:     time.Date(2017, 8, 25, 13, 42, 26, 0, time.UTC),
			start: MustParse("Wed 5:30PM"),
			end:   MustParse("Thu 5:30AM"),
			freq:  time.Minute,
			want:  time.Date(2017, 8, 30, 17, 30, 0, 0, time.UTC),
		},
	} {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			if got := nextTick(test.t, test.start, test.end, test.freq); got != test.want {
				t.Errorf("nextTick(%v, %v, %v, %v) = %v, want %v", test.t, test.start, test.end, test.freq, got, test.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	for i, test := range []struct {
		val  string
		want Time
	}{
		{"Sun 5:13AM", Time{day: time.Sunday, hour: 5, min: 13}},
		{"Mon 11:43AM", Time{day: time.Monday, hour: 11, min: 43}},
		{"Tue 12:00AM", Time{day: time.Tuesday}},
		{"Wed 12:00PM", Time{day: time.Wednesday, hour: 12}},
		{"Thu 7:30PM", Time{day: time.Thursday, hour: 19, min: 30}},
		{"Fri 11:59PM", Time{day: time.Friday, hour: 23, min: 59}},
		{"Sat 5:17PM", Time{day: time.Saturday, hour: 17, min: 17}},
	} {
		test := test
		t.Run(fmt.Sprintf("TestParse-%d", i), func(t *testing.T) {
			t.Parallel()
			got, err := Parse(test.val)
			if err != nil {
				t.Fatalf("Parse(%q) got unexpected error: %v", test.val, err)
			}
			if got != test.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", test.val, got, test.want)
			}
		})
	}
}

func TestInWeek(t *testing.T) {
	t.Parallel()
	for i, val := range []time.Time{
		time.Unix(0, 0),
		time.Unix(1502857357, 0),
		time.Unix(0, 423000),
		time.Unix(1502857357, 423000),
	} {
		for j, wantWeekTime := range []Time{
			MustParse("Sun 5:13AM"),
			MustParse("Mon 11:43AM"),
			MustParse("Tue 12:00AM"),
			MustParse("Wed 12:00PM"),
			MustParse("Thu 7:30PM"),
			MustParse("Fri 11:59PM"),
			MustParse("Sat 5:17PM"),
		} {
			val, wantWeekTime := val, wantWeekTime
			t.Run(fmt.Sprintf("TestInWeek-%d-%d", i, j), func(t *testing.T) {
				t.Parallel()
				gotTime := wantWeekTime.InWeek(val)
				if gotTime.Weekday() != wantWeekTime.day {
					t.Errorf("Want day %s, got %s", wantWeekTime.day, gotTime.Weekday())
				}
				if gotTime.Hour() != wantWeekTime.hour {
					t.Errorf("Want hour %d, got %d", wantWeekTime.hour, gotTime.Hour())
				}
				if gotTime.Minute() != wantWeekTime.min {
					t.Errorf("Want minute %d, got %d", wantWeekTime.min, gotTime.Hour())
				}
				if !gotTime.Equal(wantWeekTime.InWeek(gotTime)) {
					t.Errorf("InWeek not idempotent for time-in-week %+v, starting time %v", wantWeekTime, val)
				}
				if !inSameWeek(val, gotTime) {
					t.Errorf("Got %v, want time in same week as %v", t, val)
				}
			})

		}
	}
}

func TestBefore(t *testing.T) {
	t.Parallel()

	// Sorted.
	times := []Time{
		MustParse("Sun 5:13AM"),
		MustParse("Mon 11:43AM"),
		MustParse("Tue 12:00AM"),
		MustParse("Wed 12:00PM"),
		MustParse("Thu 7:30PM"),
		MustParse("Fri 11:59PM"),
		MustParse("Sat 5:17PM"),
	}

	for i, ti := range times {
		for j, tj := range times {
			i, ti, j, tj := i, ti, j, tj
			t.Run(fmt.Sprintf("TestBefore-%d-%d", i, j), func(t *testing.T) {
				t.Parallel()
				wantIBeforeJ := i < j
				if got := ti.Before(tj); got != wantIBeforeJ {
					t.Errorf("%q.Before(%q) = %v, want %v", ti, tj, got, wantIBeforeJ)
				}
				wantJBeforeI := j < i
				if got := tj.Before(ti); got != wantJBeforeI {
					t.Errorf("%q.Before(%q) = %v, want %v", tj, ti, got, wantJBeforeI)
				}
			})
		}
	}
}

func TestString(t *testing.T) {
	t.Parallel()
	for i, want := range []string{
		"Sun 5:13AM",
		"Mon 11:43AM",
		"Tue 12:00AM",
		"Wed 12:00PM",
		"Thu 7:30PM",
		"Fri 11:59PM",
		"Sat 5:17PM",
	} {
		want := want
		t.Run(fmt.Sprintf("TestString-%d", i), func(t *testing.T) {
			t.Parallel()
			wt, err := Parse(want)
			if err != nil {
				t.Fatalf("Parse(%s) returned error: %v", want, err)
			}
			if got := wt.String(); got != want {
				t.Errorf("wt.String() = %q, want %q", got, want)
			}
		})
	}
}

func inSameWeek(t1, t2 time.Time) bool {
	t1 = t1.AddDate(0, 0, -int(t1.Weekday()))
	t2 = t2.AddDate(0, 0, -int(t2.Weekday()))
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()
}
