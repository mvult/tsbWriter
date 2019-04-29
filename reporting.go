package tsbWriter

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
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
}

type report struct {
	name     string
	size     int
	written  int64
	complete bool
}

func printUpdate() {
	if len(reportDepot) == 0 {
		return
	} else {
		var buf bytes.Buffer
		_, err := buf.Write([]byte(fmt.Sprintf("\r")))
		if err != nil {
			panic(err)
		}
		// table := tablewriter.NewWriter(os.Stdout)
		table := tablewriter.NewWriter(&buf)
		table.SetHeader([]string{"Buffer Name", "Current Size", "Total Written"})
		for _, r := range reportDepot {
			table.Append([]string{r.name, toMB(int64(r.size)), toMB(r.written)})
		}

		table.SetCaption(true, fmt.Sprint(time.Since(initialTime)))
		table.Render()
		_, err = os.Stdout.Write(buf.Bytes())
		if err != nil {
			panic(err)
		}
		// table := tablewriter.NewWriter(os.Stdout)
	}
}

// No checks whatsoever for duplicate names
func submitReport(name string, size int, written int64, complete bool) {
	reportMutex.Lock()
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
