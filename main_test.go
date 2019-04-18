package tsbWriter

import (
	// "bytes"
	"io"
	"os"
	"testing"
	"time"
)

type slowWriter struct {
	total int64
}

func (sw slowWriter) Write(p []byte) (int, error) {
	time.Sleep(time.Nanosecond * 1)
	return len(p), nil
}

func (sw slowWriter) Close() error {
	return nil
}

func TestMain(t *testing.T) {
	var err error
	source, err := os.Open("testToSend.mp4")
	if err != nil {
		panic(err)
	}
	target := slowWriter{}
	tsbWriter := NewWriter(target, 8192, "test")

	buf := make([]byte, 1024)
	var n int

	for {
		n, err = source.Read(buf)
		if err != nil {
			if err == io.EOF {
				time.Sleep(time.Second * 35)
				return
			}
			panic(err)
		}
		tsbWriter.Write(buf[:n])
	}

}
