// rssdld is a daemon which watches RSS feeds for new files and downloads them
// to a specified directory.
package main

import (
	"flag"
	"fmt"
)

var (
	configPath = flag.String("config", "", "Path to service configuration.")
	statePath  = flag.String("state", "", "Path to state.")
)

func main() {
	fmt.Println("rssdld")
}
