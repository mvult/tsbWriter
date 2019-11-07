package tsbWriter

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	_ "os"
	"sync"
	"time"
)

var rd reportDepot
var initialTime time.Time
var previousReport time.Time

func init() {
	rd = reportDepot{internal: make(map[string]report), lock: sync.RWMutex{}}
	previousReport = time.Now()
	initialTime = time.Now()
	go startServer()
}

type reportDepot struct {
	internal map[string]report
	lock     sync.RWMutex
}

func (rd *reportDepot) GetAll() map[string]report {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	ret := make(map[string]report)
	for k, v := range rd.internal {
		ret[k] = v
	}
	return ret
}

func (rd *reportDepot) Get(k string) (report, bool) {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	ret, ok := rd.internal[k]
	return ret, ok
}

func (rd *reportDepot) Set(k string, v report) {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	rd.internal[k] = v
}

func (rd *reportDepot) Del(k string) {
	rd.lock.Lock()
	defer rd.lock.Unlock()
	delete(rd.internal, k)
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
	for _, r := range rd.GetAll() {
		table.Append([]string{r.name, toMB(int64(r.size)), toMB(r.written)})
	}

	table.SetCaption(true, fmt.Sprint(time.Since(initialTime)))
	table.Render()

	b.messages <- fmt.Sprintf("%s", buf.Bytes())
}

// No checks whatsoever for duplicate names
func submitReport(name string, size int, written int64, complete bool) {

	if complete {
		rd.Del(name)
		printUpdate()
		previousReport = time.Now()
		return
	}
	rd.Set(name, report{name: name, size: size, written: written, complete: complete})

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
