package tsbWriter

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	_ "os"
	"sync"
	"time"
)

var reportDepot map[string]report
var initialTime time.Time
var previousReport time.Time
var reportMutex sync.Mutex

func init() {
	reportDepot = make(map[string]report)
	previousReport = time.Now()
	initialTime = time.Now()
	go startServer()
}

type report struct {
	name     string
	size     int
	written  int64
	complete bool
}

func printUpdate() {
	var buf bytes.Buffer

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Buffer Name", "Current Size", "Total Written"})
	for _, r := range reportDepot {
		table.Append([]string{r.name, toMB(int64(r.size)), toMB(r.written)})
	}

	table.SetCaption(true, fmt.Sprint(time.Since(initialTime)))
	table.Render()

	b.messages <- fmt.Sprintf("%s", buf.Bytes())
}

// No checks whatsoever for duplicate names
func submitReport(name string, size int, written int64, complete bool) {
	reportMutex.Lock()

	if complete {

		delete(reportDepot, name)
		reportMutex.Unlock()
		printUpdate()
		previousReport = time.Now()
		return
	}
	reportDepot[name] = report{name: name, size: size, written: written, complete: complete}
	reportMutex.Unlock()

	if time.Since(previousReport) > time.Second*2 {
		printUpdate()
		previousReport = time.Now()
	}
}

func toMB(n int64) string {
	if n > int64(1048576) {
		return fmt.Sprintf("%.2f", float64(n)/float64(1048576)) + " MB"
	} else {
		return fmt.Sprint(n)
	}
}
