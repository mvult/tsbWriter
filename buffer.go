package tsbWriter

import (
	"bytes"
	"sync"
)

type buffer struct {
	b      bytes.Buffer
	m      sync.Mutex
	closed bool
}

func (b *buffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

func (b *buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *buffer) Len() (n int) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Len()
}
