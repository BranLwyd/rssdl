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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
				}
			`,
			want: []*Feed{
				{
					Name:        "feed name",
					URL:         "feed url",
					DownloadDir: "/download/dir",
					OrderRegexp: regexp.MustCompile("(order_regex)"),
					CheckSpecs: []weekly.TickSpecification{
						{
							Start:     weekly.MustParse("Tue 12:00PM"),
							End:       weekly.MustParse("Thu 12:00PM"),
							Frequency: 60 * time.Second,
						},
					},
				},
			},
		},
		{
			desc: "multi_check_spec",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
					check_spec {
						start: "Thu 12:00PM"
						end: "Fri 12:00PM"
						freq_s: 30
					}
				}
			`,
			want: []*Feed{
				{
					Name:        "feed name",
					URL:         "feed url",
					DownloadDir: "/download/dir",
					OrderRegexp: regexp.MustCompile("(order_regex)"),
					CheckSpecs: []weekly.TickSpecification{
						{
							Start:     weekly.MustParse("Tue 12:00PM"),
							End:       weekly.MustParse("Thu 12:00PM"),
							Frequency: 60 * time.Second,
						},
						{
							Start:     weekly.MustParse("Thu 12:00PM"),
							End:       weekly.MustParse("Fri 12:00PM"),
							Frequency: 30 * time.Second,
						},
					},
				},
			},
		},
		{
			desc: "fallback_to_defaults",
			cfg: `
				download_dir: "/download/dir"
				order_regex: "(order_regex)"
				check_spec {
					start: "Tue 12:00PM"
					end: "Thu 12:00PM"
					freq_s: 60
				}
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
					CheckSpecs: []weekly.TickSpecification{
						{
							Start:     weekly.MustParse("Tue 12:00PM"),
							End:       weekly.MustParse("Thu 12:00PM"),
							Frequency: 60 * time.Second,
						},
					},
				},
			},
		},
		{
			desc: "override_defaults",
			cfg: `
				download_dir: "/bad/download/dir"
				order_regex: "(bad_order_regex)"
				check_spec {
					start: "bad_check_start"
					end: "bad_check_end"
					freq_s: 120
				}
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
				}
			`,
			want: []*Feed{
				{
					Name:        "feed name",
					URL:         "feed url",
					DownloadDir: "/download/dir",
					OrderRegexp: regexp.MustCompile("(order_regex)"),
					CheckSpecs: []weekly.TickSpecification{
						{
							Start:     weekly.MustParse("Tue 12:00PM"),
							End:       weekly.MustParse("Thu 12:00PM"),
							Frequency: 60 * time.Second,
						},
					},
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
				}
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
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
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
						freq_s: 60
					}
				}
			`,
			wantErr: regexp.MustCompile(`has \d+ capture groups, expected 1`),
		},
		{
			desc: "no_check_spec",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
				}
			`,
			wantErr: regexp.MustCompile("has no check_spec"),
		},
		{
			desc: "no_check_start",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						end: "Thu 12:00PM"
						freq_s: 60
					}
				}
			`,
			wantErr: regexp.MustCompile("has no start"),
		},
		{
			desc: "unparseable_check_start",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "f49m1@30"
						end: "Thu 12:00PM"
						freq_s: 60
					}
				}
			`,
			wantErr: regexp.MustCompile("error parsing start"),
		},
		{
			desc: "no_check_end",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "Tue 12:00PM"
						freq_s: 60
					}
				}
			`,
			wantErr: regexp.MustCompile("has no end"),
		},
		{
			desc: "unparseable_check_end",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "Tue 12:00PM"
						end: "f49m1@30"
						freq_s: 60
					}
				}
			`,
			wantErr: regexp.MustCompile("error parsing end"),
		},
		{
			desc: "check_end_before_check_start",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "Thu 12:00PM"
						end: "Tue 12:00PM"
						freq_s: 60
					}
				}
			`,
			wantErr: regexp.MustCompile("has end before start"),
		},
		{
			desc: "no_check_freq",
			cfg: `
				feed {
					name: "feed name"
					url: "feed url"
					download_dir: "/download/dir"
					order_regex: "(order_regex)"
					check_spec {
						start: "Tue 12:00PM"
						end: "Thu 12:00PM"
					}
				}
			`,
			wantErr: regexp.MustCompile("missing or zero freq_s"),
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
					t.Errorf("Parse got error %q, wanted error matching pattern %q", err, test.wantErr)
				}
			}
		})
	}
}
