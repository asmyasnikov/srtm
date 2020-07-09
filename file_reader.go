package srtm

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type FileReader struct {
	name string
	f    *os.File
	m    *sync.Mutex
	c    int64
}

func newFileReader(tPath string) *FileReader {
	return &FileReader{
		name: tPath,
		f:    nil,
		m:    &sync.Mutex{},
		c:    0,
	}
}

func (f *FileReader) open() error {
	f.m.Lock()
	defer f.m.Unlock()
	if atomic.AddInt64(&f.c, 1) > 1 {
		return nil
	}
	file, err := os.Open(f.name)
	if err != nil {
		return err
	}
	f.f = file
	return nil
}

func (f *FileReader) close() error {
	f.m.Lock()
	defer f.m.Unlock()
	if atomic.AddInt64(&f.c, -1) == 0 {
		return f.f.Close()
	}
	return nil
}

func (f *FileReader) elevation(idx int) (int16, error) {
	b := make([]byte, 2)
	n, err := f.f.ReadAt(b, int64(idx)*2)
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, fmt.Errorf("error on read file %s at index %d", f.name, idx)
	}
	return int16(binary.BigEndian.Uint16(b)), nil
}
