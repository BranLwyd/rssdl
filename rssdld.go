// rssdld is a daemon which watches RSS feeds for new files and downloads them
// to a specified directory.
package main

import (
	"context"
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
	"time"

	"github.com/BranLwyd/rssdl/alert"
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

	ticker, err := weekly.NewTicker(f.CheckSpecs)
	if err != nil {
		log.Fatalf("[%s] Could not create ticker: %v", f.Name, err)
	}

	log.Printf("Watching %q", f.Name)
CHECK_LOOP:
	for range ticker.C {
		log.Printf("[%s] Checking", f.Name)
		feed, err := parser.ParseURL(f.URL)
		if err != nil {
			sendAlert(f.Alerter, alert.ERROR, fmt.Sprintf("[%s] Could not parse feed", f.Name))
			fmt.Printf("[%s] Could not parse feed: %v", f.Name, err)
			continue
		}
		itms := feed.Items

		// Order the feed's items, oldest first.
		for _, itm := range itms {
			if itm.PublishedParsed == nil {
				sendAlert(f.Alerter, alert.ERROR, fmt.Sprintf("[%s] Item with no publish time", f.Name))
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
				sendAlert(f.Alerter, alert.ERROR, fmt.Sprintf("[%s] Could not download item", f.Name))
				fmt.Printf("[%s] Could not download %q: %v", f.Name, itm.Title, err)
				break
			} else {
				sendAlert(f.Alerter, alert.NEW_ITEM, fmt.Sprintf("[%s] Got new item: %s", f.Name, o))
			}
			order, orderModified = o, true
		}
		if orderModified {
			if err := s.SetOrder(f.Name, order); err != nil {
				// TODO: if writing fails, retry writes independently of checks
				// (otherwise, pending writes may stay in memory for a week!)
				sendAlert(f.Alerter, alert.ERROR, fmt.Sprintf("[%s] Error updating order", f.Name))
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

func sendAlert(a alert.Alerter, code alert.Code, details string) {
	const alertTimeout = time.Minute

	if a != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), alertTimeout)
			defer cancel()
			if err := a.Alert(ctx, code, details); err != nil {
				log.Printf("Error while alerting ([%s] %s): %v", code, details, err)
			}
		}()
	}
}
