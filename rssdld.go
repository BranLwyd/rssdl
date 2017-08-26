// rssdld is a daemon which watches RSS feeds for new files and downloads them
// to a specified directory.
package main

import (
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
	"sort"
	"strings"

	"github.com/BranLwyd/rssdl/config"
	"github.com/BranLwyd/rssdl/state"
	"github.com/BranLwyd/rssdl/weekly"
	"github.com/mmcdole/gofeed"
)

var (
	configPath = flag.String("config", "", "Path to service configuration file.")
	statePath  = flag.String("state", "", "Path to state file.")
)

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
	feeds, err := config.Parse(string(cfgBytes))
	if err != nil {
		log.Fatalf("Could not parse config: %v", err)
	}

	// Parse state.
	s, err := state.Open(*statePath)
	if err != nil {
		log.Fatalf("Could not open state: %v", err)
	}

	// Start feed-checker goroutines.
	for _, feed := range feeds {
		go checkFeed(feed, s)
	}
	select {}
}

func checkFeed(f *config.Feed, s *state.State) {
	parser := gofeed.NewParser()
	order := s.GetOrder(f.Name)
	orderModified := false

	ticker, err := weekly.NewTicker(f.CheckStart, f.CheckEnd, f.CheckFreq)
	if err != nil {
		log.Fatalf("[%s] Could not create ticker: %v", f.Name, err)
	}

	log.Printf("Watching %q", f.Name)
CHECK_LOOP:
	for range ticker.C {
		log.Printf("[%s] Checking", f.Name)
		feed, err := parser.ParseURL(f.URL)
		if err != nil {
			fmt.Printf("[%s] Could not parse feed: %v", f.Name, err)
			continue
		}
		itms := feed.Items

		// Order the feed's items, oldest first.
		for _, itm := range itms {
			if itm.PublishedParsed == nil {
				fmt.Printf("[%s] %q has no published time, or time could not be parsed", f.Name, itm.Title)
				continue CHECK_LOOP
			}
		}
		sort.SliceStable(itms, func(i, j int) bool { return itms[i].PublishedParsed.Before(*itms[j].PublishedParsed) })

		for _, itm := range itms {
			// Check order.
			m := f.OrderRegexp.FindStringSubmatch(itm.Title)
			if m == nil {
				continue
			}
			o := m[1]
			if o <= order {
				continue
			}

			// Download.
			log.Printf("[%s] Found %s", f.Name, itm.Title)
			if err := download(itm.Link, f.DownloadDir); err != nil {
				fmt.Printf("[%s] Could not download %q: %v", f.Name, itm.Title, err)
				break
			}
			order, orderModified = o, true
		}
		if orderModified {
			if err := s.SetOrder(f.Name, order); err != nil {
				// TODO: if writing fails, retry writes independently of checks
				// (otherwise, pending writes may stay in memory for a week!)
				fmt.Printf("[%s] Could not update order: %v", f.Name, err)
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
