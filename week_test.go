package week

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
