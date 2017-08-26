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
	CheckStart  weekly.Time
	CheckEnd    weekly.Time
	CheckFreq   time.Duration
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

		csStr := defaultString(f.CheckStart, c.CheckStart)
		if csStr == "" {
			return nil, fmt.Errorf("feed %q has no check_start and no default specified", f.Name)
		}
		cs, err := weekly.Parse(csStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing check_start for feed %q: %v", f.Name, err)
		}

		ceStr := defaultString(f.CheckEnd, c.CheckEnd)
		if ceStr == "" {
			return nil, fmt.Errorf("feed %q has no check_end and no default specified", f.Name)
		}
		ce, err := weekly.Parse(ceStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing check_end for feed %q: %v", f.Name, err)
		}

		if ce.Before(cs) {
			return nil, fmt.Errorf("feed %q has check_end before check_start", f.Name)
		}

		cfs := defaultUint32(f.CheckFreqS, c.CheckFreqS)
		if cfs == 0 {
			return nil, fmt.Errorf("feed %q has no check_freq_s and no default specified", f.Name)
		}
		cf := time.Duration(cfs) * time.Second

		feeds = append(feeds, &Feed{
			Name:        f.Name,
			URL:         f.Url,
			DownloadDir: dd,
			OrderRegexp: re,
			CheckStart:  cs,
			CheckEnd:    ce,
			CheckFreq:   cf,
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
