package config

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/BranLwyd/rssdl/weekly"
)

func TestParse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		desc    string
		cfg     string
		want    []*Feed
		wantErr *regexp.Regexp
	}{
		{
			desc: "basic",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			want: []*Feed{
				{
					Name:        "feed name",
					URL:         "feed url",
					DownloadDir: "/download/dir",
					OrderRegexp: regexp.MustCompile("(order_regex)"),
					CheckStart:  weekly.MustParse("Tue 12:00PM"),
					CheckEnd:    weekly.MustParse("Thu 12:00PM"),
					CheckFreq:   60 * time.Second,
				},
			},
		},
		{
			desc: "fallback_to_defaults",
			cfg: `
				download_dir: "/download/dir"
				order_regex: "(order_regex)"
				check_start: "Tue 12:00PM"
				check_end: "Thu 12:00PM"
				check_freq_s: 60
				feed {
					name: "feed name"
					url: "feed url"
				}
			`,
			want: []*Feed{
				{
					Name:        "feed name",
					URL:         "feed url",
					DownloadDir: "/download/dir",
					OrderRegexp: regexp.MustCompile("(order_regex)"),
					CheckStart:  weekly.MustParse("Tue 12:00PM"),
					CheckEnd:    weekly.MustParse("Thu 12:00PM"),
					CheckFreq:   60 * time.Second,
				},
			},
		},
		{
			desc: "override_defaults",
			cfg: `
				download_dir: "/bad/download/dir"
				order_regex: "(bad_order_regex)"
				check_start: "bad_check_start"
				check_end: "bad_check_end"
				check_freq_s: 120
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			want: []*Feed{
				{
					Name:        "feed name",
					URL:         "feed url",
					DownloadDir: "/download/dir",
					OrderRegexp: regexp.MustCompile("(order_regex)"),
					CheckStart:  weekly.MustParse("Tue 12:00PM"),
					CheckEnd:    weekly.MustParse("Thu 12:00PM"),
					CheckFreq:   60 * time.Second,
				},
			},
		},
		{
			desc:    "unparseable",
			cfg:     `^#$mf90@#`,
			wantErr: regexp.MustCompile("could not parse config"),
		},
		{
			desc:    "no_feed",
			cfg:     ``,
			wantErr: regexp.MustCompile("config does not specify any feeds to watch"),
		},
		{
			desc: "no_name",
			cfg: `
				feed {
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile(`feed at index \d+ has no name`),
		},
		{
			desc: "duplicate_name",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("duplicate feed name"),
		},
		{
			desc: "no_url",
			cfg: `
				feed {
					name: "feed name"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("feed .* has no URL"),
		},
		{
			desc: "no_download_dir",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("has no download_dir"),
		},
		{
			desc: "no_order_regex",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("has no order_regex"),
		},
		{
			desc: "unparseable_order_regex",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: ")"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("error parsing order_regex"),
		},
		{
			desc: "order_regex_no_capture",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "order_regex"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile(`has \d+ capture groups, expected 1`),
		},
		{
			desc: "order_regex_multi_capture",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order)_(regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile(`has \d+ capture groups, expected 1`),
		},
		{
			desc: "no_check_start",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("has no check_start"),
		},
		{
			desc: "unparseable_check_start",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "f49m1@30"
					check_end: "Thu 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("error parsing check_start"),
		},
		{
			desc: "no_check_end",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("has no check_end"),
		},
		{
			desc: "unparseable_check_end",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "f49m1@30"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("error parsing check_end"),
		},
		{
			desc: "check_end_before_check_start",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Thu 12:00PM"
					check_end: "Tue 12:00PM"
					check_freq_s: 60
				}
			`,
			wantErr: regexp.MustCompile("has check_end before check_start"),
		},
		{
			desc: "no_check_freq",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_start: "Tue 12:00PM"
					check_end: "Thu 12:00PM"
				}
			`,
			wantErr: regexp.MustCompile("has no check_freq_s"),
		},
	} {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(test.cfg)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Got %v, want %v", got, test.want)
			}
			switch {
			case test.wantErr == nil && err != nil:
				t.Errorf("Unexpected error: %v", err)
			case test.wantErr != nil:
				if err == nil || !test.wantErr.MatchString(err.Error()) {
					t.Errorf("Got error %q, wanted error matching pattern %q", err, test.wantErr)
				}
			}
		})
	}
}
