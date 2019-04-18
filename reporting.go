package tsbWriter

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"time"
)

var reportDepot map[string]report

func init() {
	reportDepot = make(map[string]report)
	go func() {
		for {
			time.Sleep(time.Second * 1)
			printUpdate()
		}
	}()
}

type report struct {
	name    string
	size    int
	written int64
}

func printUpdate() {
	if len(reportDepot) == 0 {
		return
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Buffer Name", "Current Size", "Total Written"})
		for _, r := range reportDepot {
			table.Append([]string{r.name, toMB(int64(r.size)), toMB(r.written)})
		}

		table.SetCaption(true, fmt.Sprint(time.Now().Format("15:04:05")))
		table.Render()
	}
}

// No checks whatsoever for duplicate names
func submitReport(name string, size int, written int64) {
	reportDepot[name] = report{name: name, size: size, written: written}
}

func toMB(n int64) string {
	if n > int64(1048576) {
		return fmt.Sprintf("%.2f", float64(n)/float64(1048576)) + " MB"
	} else {
		return fmt.Sprint(n)
	}
}
