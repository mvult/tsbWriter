package tsbWriter

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"
)

var logger *log.Logger
var max int

func init() {
	logger = log.New(os.Stdout, "", log.Lshortfile)
}

type TSBWriter struct {
	writer      io.WriteCloser
	internal    *buffer
	name        string
	written     int64
	closeChan   chan struct{}
	errorChan   chan error
	writerError bool
}

func NewWriter(w io.WriteCloser, writeSize int, name string) TSBWriter {
	ret := TSBWriter{}
	ret.writer = w
	ret.internal = &buffer{}
	ret.closeChan = make(chan struct{})
	ret.errorChan = make(chan error)
	if name != "" {
		ret.name = fmt.Sprintf("%v %v", name, rand.Int31())
		go ret.reportSize()
	}
	go ret.writeToDestination(writeSize)

	return ret
}

func (b TSBWriter) Write(p []byte) (int, error) {
	select {
	case err := <-b.errorChan:
		return 0, err
	default:
		return b.internal.Write(p)
	}
}

func (b TSBWriter) Close() error {
	if b.writerError {
		b.writer = nil
		return nil
	}

	go func() {
		b.internal.closed = true
		<-b.closeChan
		err := b.writer.Close()
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func (b *TSBWriter) reportSize() {
	for {
		if b.internal.closed {
			if b.internal.b.Len() == 0 {
				submitReport(b.name, b.internal.Len(), b.written, true)
				return
			}
		}
		time.Sleep(time.Second * 3)
		submitReport(b.name, b.internal.Len(), b.written, false)
	}
}

func (b *TSBWriter) writeToDestination(writeSize int) {

	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Panic on %v buffer: %v\nPassing panic upstream.\n", b.name, r)
			panic(r)
		}
	}()

	var n int
	var err error
	buf := make([]byte, writeSize)
	for {
		n, err = b.internal.Read(buf)
		if err != nil {
			if err == io.EOF {
				if b.internal.closed {
					if n != 0 {
						n, err = b.writer.Write(buf[:n])
						if err != nil {
							logger.Panicln(err)
						}
						b.written += int64(n)
					}
					b.closeChan <- struct{}{}
					return
				}
				time.Sleep(time.Millisecond * 500)
				continue
			} else {
				logger.Panicln(err)
			}
		}
		n, err = b.writer.Write(buf[:n])
		if err != nil {
			// logger.Panicln(err)
			err = fmt.Errorf("TSBWriter error writing to underlying writer: %v", err)
			logger.Println(err)
			b.errorChan <- err
			return
		}
		b.written += int64(n)
	}
}
