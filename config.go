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
	cpb := &pb.Config{}
	if err := proto.UnmarshalText(cfg, cpb); err != nil {
		return nil, fmt.Errorf("could not parse config: %v", err)
	}
	if len(cpb.Feed) == 0 {
		return nil, errors.New("config does not specify any feeds to watch")
	}
	feeds := make([]*Feed, 0, len(cpb.Feed))
	names := make(map[string]struct{}, len(cpb.Feed))

	for i, fpb := range cpb.Feed {
		if fpb.Name == "" {
			return nil, fmt.Errorf("unnamed feed at index %d", i)
		}
		if _, ok := names[fpb.Name]; ok {
			return nil, fmt.Errorf("duplicate feed name %q", fpb.Name)
		}
		names[fpb.Name] = struct{}{}

		if fpb.Url == "" {
			return nil, fmt.Errorf("feed %q has no URL", fpb.Name)
		}

		dd := defaultString(fpb.DownloadDir, cpb.DownloadDir)
		if dd == "" {
			return nil, fmt.Errorf("feed %q has no download_dir and no default specified", fpb.Name)
		}

		reStr := defaultString(fpb.OrderRegex, cpb.OrderRegex)
		if reStr == "" {
			return nil, fmt.Errorf("feed %q had no order_regex and no default specified", fpb.Name)
		}
		re, err := regexp.Compile(reStr)
		if err != nil {
			return nil, fmt.Errorf("couldn't compile order regex for feed %q: %v", fpb.Name, err)
		}
		if re.NumSubexp() != 1 {
			return nil, fmt.Errorf("order regex for feed %q had %d subexpressions, expected 1", fpb.Name, re.NumSubexp())
		}

		csStr := defaultString(fpb.CheckStart, cpb.CheckStart)
		if csStr == "" {
			return nil, fmt.Errorf("feed %q had no check_start and no default specified", fpb.Name)
		}
		cs, err := weekly.Parse(csStr)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse check start time for feed %q: %v", fpb.Name, err)
		}

		ceStr := defaultString(fpb.CheckEnd, cpb.CheckEnd)
		if ceStr == "" {
			return nil, fmt.Errorf("feed %q had no check_end and no default specified", fpb.Name)
		}
		ce, err := weekly.Parse(fpb.CheckEnd)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse check end time for feed %q: %v", fpb.Name, err)
		}

		if ce.Before(cs) {
			return nil, fmt.Errorf("feed %q has check_end before check_start", fpb.Name)
		}

		cfs := defaultUint32(fpb.CheckFreqS, cpb.CheckFreqS)
		if cfs == 0 {
			return nil, fmt.Errorf("feed %q had no check_freq_s and no default specified", fpb.Name)
		}
		cf := time.Duration(cfs) * time.Second

		feeds = append(feeds, &Feed{
			Name:        fpb.Name,
			URL:         fpb.Url,
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
