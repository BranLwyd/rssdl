// rssdld is a daemon which watches RSS feeds for new files and downloads them
// to a specified directory.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/BranLwyd/rssdl/weekly"
	"github.com/golang/protobuf/proto"
	"github.com/mmcdole/gofeed"

	pb "github.com/BranLwyd/rssdl/rssdl_proto"
)

var (
	configPath = flag.String("config", "", "Path to service configuration.")
	statePath  = flag.String("state", "", "Path to state.")
)

type Feed struct {
	URL         string
	OrderRE     *regexp.Regexp
	CheckTicker *weekly.Ticker
}

type Config struct {
	Feeds       map[string]*Feed
	DownloadDir string
}

func ParseConfig(cfg string) (*Config, error) {
	cpb := &pb.Config{}
	if err := proto.UnmarshalText(cfg, cpb); err != nil {
		return nil, fmt.Errorf("could not parse config: %v", err)
	}
	if cpb.DownloadDir == "" {
		return nil, errors.New("no download_dir specified")
	}

	c := &Config{
		Feeds:       map[string]*Feed{},
		DownloadDir: cpb.DownloadDir,
	}
	for _, fpb := range cpb.Feed {
		if _, ok := c.Feeds[fpb.Name]; ok {
			return nil, fmt.Errorf("duplicate feed name %q", fpb.Name)
		}

		re, err := regexp.Compile(fpb.OrderRegex)
		if err != nil {
			return nil, fmt.Errorf("couldn't compile order regex for feed %q: %v", fpb.Name, err)
		}
		if re.NumSubexp() != 1 {
			return nil, fmt.Errorf("order regex for feed %q had %d subexpressions, expected 1", fpb.Name, re.NumSubexp())
		}

		s, err := weekly.Parse(fpb.CheckStart)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse check start time for feed %q: %v", fpb.Name, err)
		}
		e, err := weekly.Parse(fpb.CheckEnd)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse check end time for feed %q: %v", fpb.Name, err)
		}
		freq := time.Duration(fpb.CheckFreqS) * time.Second

		c.Feeds[fpb.Name] = &Feed{
			URL:         fpb.Url,
			OrderRE:     re,
			CheckTicker: weekly.NewTicker(s, e, freq),
		}
	}
	return c, nil
}

type State struct {
	filename string

	mu sync.RWMutex // protects s
	s  *pb.State
}

func NewState(filename string) (*State, error) {
	sBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// No state file. Return an empty state.
			// Write immediately so we'll fail out now if the state is in an unwritable location.
			log.Printf("State file %q does not exist. Starting fresh", filename)
			s := &State{
				filename: filename,
				s:        &pb.State{},
			}
			if err := s.write(); err != nil {
				return nil, err
			}
			return s, nil
		}
		return nil, fmt.Errorf("could not read state file: %v", err)
	}

	s := &pb.State{}
	if err := proto.Unmarshal(sBytes, s); err != nil {
		return nil, fmt.Errorf("could not parse state: %v", err)
	}

	return &State{
		filename: filename,
		s:        s,
	}, nil
}

func (s *State) GetOrder(name string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fs := s.s.FeedState[name]
	if fs == nil {
		return ""
	}
	return fs.Order
}

func (s *State) SetOrder(name, order string) (retErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If s.write encounters an error, we may end up with in-memory state not matching written state.
	// But that's fine -- we'll retry writes, and in the meantime we don't want to re-download already-downloaded links.

	fs := s.s.FeedState[name]
	if fs == nil {
		if s.s.FeedState == nil {
			s.s.FeedState = map[string]*pb.State_FeedState{}
		}
		fs = &pb.State_FeedState{}
		s.s.FeedState[name] = fs
	}
	fs.Order = order
	return s.write()
}

// Assumes that s.mu is already locked. (needs at least a read-lock)
func (s *State) write() error {
	sBytes, err := proto.Marshal(s.s)
	if err != nil {
		return fmt.Errorf("could not marshal state proto: %v", err)
	}

	// Use a temporary file so that updates are atomic.
	f, err := ioutil.TempFile(filepath.Dir(s.filename), ".rssdl_state_")
	if err != nil {
		return fmt.Errorf("could not create state file: %v", err)
	}
	defer func() {
		f.Close()
		if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Could not remove %q: %v", f.Name(), err)
		}
	}()
	if _, err := f.Write(sBytes); err != nil {
		return fmt.Errorf("could not write state file: %v", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close state file: %v", err)
	}
	if err := os.Rename(f.Name(), s.filename); err != nil {
		return fmt.Errorf("could not rename state file: %v", err)
	}
	return nil
}

func main() {
	flag.Parse()
	if *configPath == "" {
		log.Fatalf("--config is required")
	}
	if *statePath == "" {
		log.Fatalf("--state is required")
	}

	// Parse config.
	cfgBytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Could not read config file: %v", err)
	}
	cfg, err := ParseConfig(string(cfgBytes))
	if err != nil {
		log.Fatalf("Could not parse config: %v", err)
	}
	if len(cfg.Feeds) == 0 {
		log.Printf("Config does not specify any feeds to watch")
	}

	// Parse state.
	s, err := NewState(*statePath)
	if err != nil {
		log.Fatalf("Could not ready state: %v", err)
	}

	// Start feed-checker goroutines.
	for name := range cfg.Feeds {
		go checkFeed(name, cfg, s)
	}
	select {}
}

func checkFeed(name string, cfg *Config, s *State) {
	f := cfg.Feeds[name]
	parser := gofeed.NewParser()
	order := s.GetOrder(name)
	orderModified := false

	log.Printf("Watching %q", name)
CHECK_LOOP:
	for range f.CheckTicker.C {
		log.Printf("[%s] Checking", name)
		feed, err := parser.ParseURL(f.URL)
		if err != nil {
			fmt.Printf("[%s] Could not parse feed: %v", name, err)
			continue
		}
		itms := feed.Items

		// Order the feed's items, oldest first.
		for _, itm := range itms {
			if itm.PublishedParsed == nil {
				fmt.Printf("[%s] %q has no published time, or time could not be parsed", name, itm.Title)
				continue CHECK_LOOP
			}
		}
		sort.SliceStable(itms, func(i, j int) bool { return itms[i].PublishedParsed.Before(*itms[j].PublishedParsed) })

		for _, itm := range itms {
			// Check order.
			m := f.OrderRE.FindStringSubmatch(itm.Title)
			if m == nil {
				continue
			}
			o := m[1]
			if o <= order {
				continue
			}

			// Download.
			log.Printf("[%s] Found %s", name, itm.Title)
			if err := download(itm.Link, cfg.DownloadDir); err != nil {
				fmt.Printf("[%s] Could not download %q: %v", name, itm.Title, err)
				break
			}
			order, orderModified = o, true
		}
		if orderModified {
			if err := s.SetOrder(name, order); err != nil {
				// TODO: if writing fails, retry writes independently of checks
				// (otherwise, pending writes may stay in memory for a week!)
				fmt.Printf("[%s] Could not update order: %v", name, err)
			} else {
				orderModified = false
			}
		}
	}
}

func download(dlURL, dir string) error {
	// Figure out eventual filename (and sanity check the URL).
	u, err := url.Parse(dlURL)
	if err != nil {
		return fmt.Errorf("could not parse URL %q: %v", dlURL, err)
	}
	bp := path.Base(u.Path)
	if strings.HasSuffix(bp, ".") || strings.HasSuffix(bp, "/") {
		return fmt.Errorf("URL %q has no filename", dlURL)
	}
	fn := filepath.Join(dir, bp)

	// Download to a temporary file first so publishing is atomic.
	f, err := ioutil.TempFile(dir, ".rssdl_download_")
	if err != nil {
		return fmt.Errorf("could not create file: %v", err)
	}
	defer func() {
		f.Close()
		if err := os.Remove(f.Name()); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Could not remove %q: %v", f.Name(), err)
		}
	}()
	resp, err := http.Get(dlURL)
	if err != nil {
		return fmt.Errorf("could not begin getting %q: %v", dlURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("got unexpected status code when getting %q: %s", dlURL, resp.StatusCode)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("could not read %q: %v", dlURL, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close file: %v", err)
	}
	if err := os.Chmod(f.Name(), 0640); err != nil {
		return fmt.Errorf("could not chmod file: %v", err)
	}
	if err := os.Rename(f.Name(), fn); err != nil {
		return fmt.Errorf("could not rename file: %v", err)
	}
	return nil
}
