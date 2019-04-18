package tsbWriter

import (
	"io"
	"log"
	"os"
	"time"
)

var logger *log.Logger
var max int

func init() {
	logger = log.New(os.Stdout, "", log.Lshortfile)
}

type TSBWriter struct {
	writer   io.WriteCloser
	internal *buffer
	name     string
	written  int64
}

func NewWriter(w io.WriteCloser, writeSize int, name string) TSBWriter {
	ret := TSBWriter{}
	ret.writer = w
	ret.internal = &buffer{}
	if name != "" {
		ret.name = name
		go ret.reportSize()
	}
	go ret.writeToDestination(writeSize)

	return ret
}

func (b *TSBWriter) Write(p []byte) (int, error) {
	return b.internal.Write(p)
}

func (b *TSBWriter) Close() error {
	defer func() {
		b.internal.closed = true
	}()
	return b.writer.Close()
}

func (b *TSBWriter) reportSize() {
	for {
		if b.internal.closed {
			if b.internal.b.Len() == 0 {
				return
			}
		}
		time.Sleep(time.Second * 2)
		submitReport(b.name, b.internal.Len(), b.written)
	}
}

func (b *TSBWriter) writeToDestination(writeSize int) {
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
					return
				}
				time.Sleep(time.Millisecond * 500)
				continue
			} else {
				panic(err)
			}
		}
		n, err = b.writer.Write(buf[:n])
		if err != nil {
			logger.Panicln(err)
		}
		b.written += int64(n)
	}
}
