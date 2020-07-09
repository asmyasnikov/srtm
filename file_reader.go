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
	p    *sync.Pool
}

func newFileReader(tPath string) *FileReader {
	return &FileReader{
		name: tPath,
		f:    nil,
		m:    &sync.Mutex{},
		c:    0,
		p:    &sync.Pool{
			New: func() interface{} {
				return make([]byte, 2)
			},
		},
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
	b, ok := f.p.Get().([]byte)
	if !ok {
		return 0, fmt.Errorf("error on get byte slice from pool in file reader %s", f.name)
	}
	defer f.p.Put(b)
	n, err := f.f.ReadAt(b, int64(idx)*2)
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, fmt.Errorf("error on read file %s at index %d", f.name, idx)
	}
	return int16(binary.BigEndian.Uint16(b)), nil
}
