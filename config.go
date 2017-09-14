package config

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/BranLwyd/rssdl/weekly"
	"github.com/golang/protobuf/proto"

	pb "github.com/BranLwyd/rssdl/rssdl_proto"
)

type Feed struct {
	Name        string
	URL         string
	DownloadDir string
	OrderRegexp *regexp.Regexp
	CheckSpecs  []weekly.TickSpecification
}

func Parse(cfg string) ([]*Feed, error) {
	c := &pb.Config{}
	if err := proto.UnmarshalText(cfg, c); err != nil {
		return nil, fmt.Errorf("could not parse config: %v", err)
	}
	if len(c.Feed) == 0 {
		return nil, errors.New("config does not specify any feeds to watch")
	}
	feeds := make([]*Feed, 0, len(c.Feed))
	names := make(map[string]struct{}, len(c.Feed))

	for i, f := range c.Feed {
		if f.Name == "" {
			return nil, fmt.Errorf("feed at index %d has no name", i)
		}
		if _, ok := names[f.Name]; ok {
			return nil, fmt.Errorf("duplicate feed name %q", f.Name)
		}
		names[f.Name] = struct{}{}

		if f.Url == "" {
			return nil, fmt.Errorf("feed %q has no URL", f.Name)
		}

		dd := defaultString(f.DownloadDir, c.DownloadDir)
		if dd == "" {
			return nil, fmt.Errorf("feed %q has no download_dir and no default specified", f.Name)
		}

		reStr := defaultString(f.OrderRegex, c.OrderRegex)
		if reStr == "" {
			return nil, fmt.Errorf("feed %q has no order_regex and no default specified", f.Name)
		}
		re, err := regexp.Compile(reStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing order_regex for feed %q: %v", f.Name, err)
		}
		if re.NumSubexp() != 1 {
			return nil, fmt.Errorf("order regex for feed %q has %d capture groups, expected 1", f.Name, re.NumSubexp())
		}

		cs := f.CheckSpec
		if len(cs) == 0 {
			cs = c.CheckSpec
		}
		if len(cs) == 0 {
			return nil, fmt.Errorf("feed %q has no check_spec and no default specified", f.Name)
		}
		ts := make([]weekly.TickSpecification, 0, len(cs))
		for i, cs := range cs {
			if cs.Start == "" {
				return nil, fmt.Errorf("feed %q check_spec[%d] has no start", f.Name, i)
			}
			start, err := weekly.Parse(cs.Start)
			if err != nil {
				return nil, fmt.Errorf("error parsing start for feed %q check_spec[%d]: %v", f.Name, i, err)
			}

			if cs.End == "" {
				return nil, fmt.Errorf("feed %q check_spec[%d] has no end", f.Name, i)
			}
			end, err := weekly.Parse(cs.End)
			if err != nil {
				return nil, fmt.Errorf("error parsing end for feed %q check_spec[%d]: %v", f.Name, i, err)
			}

			if end.Before(start) {
				return nil, fmt.Errorf("feed %q check_spec[%d] has end before start", f.Name, i)
			}

			if cs.FreqS == 0 {
				return nil, fmt.Errorf("feed %q check_spec[%d] has missing or zero freq_s", f.Name, i)
			}
			freq := time.Duration(cs.FreqS) * time.Second
			ts = append(ts, weekly.TickSpecification{
				Start:     start,
				End:       end,
				Frequency: freq,
			})
		}

		feeds = append(feeds, &Feed{
			Name:        f.Name,
			URL:         f.Url,
			DownloadDir: dd,
			OrderRegexp: re,
			CheckSpecs:  ts,
		})
	}
	return feeds, nil
}

func defaultString(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func defaultUint32(val, defaultVal uint32) uint32 {
	if val == 0 {
		return defaultVal
	}
	return val
}
