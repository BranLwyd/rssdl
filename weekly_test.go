package weekly

import (
	"fmt"
	"testing"
	"time"
)

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
		i, test := i, test
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
		for j, wantWeekTimeStr := range []string{
			"Sun 5:13AM",
			"Mon 11:43AM",
			"Tue 12:00AM",
			"Wed 12:00PM",
			"Thu 7:30PM",
			"Fri 11:59PM",
			"Sat 5:17PM",
		} {
			i, j, val, wantWeekTimeStr := i, j, val, wantWeekTimeStr
			t.Run(fmt.Sprintf("TestInWeek-%d-%d", i, j), func(t *testing.T) {
				t.Parallel()
				wantWeekTime, err := Parse(wantWeekTimeStr)
				if err != nil {
					t.Fatalf("Unexpected error parsing Time string %q: %v", wantWeekTimeStr, err)
				}
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
					t.Errorf("InWeek not idempotent for time-in-week %s, starting time %v", wantWeekTimeStr, val)
				}
				if !inSameWeek(val, gotTime) {
					t.Errorf("Got %v, want time in same week as %v", t, val)
				}
			})

		}
	}
}

func inSameWeek(t1, t2 time.Time) bool {
	t1 = t1.AddDate(0, 0, -int(t1.Weekday()))
	t2 = t2.AddDate(0, 0, -int(t2.Weekday()))
	return t1.Year() == t2.Year() &&
		t1.Month() == t2.Month() &&
		t1.Day() == t2.Day()
}
