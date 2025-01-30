package main

import (
	"encoding/binary"
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"unsafe"
)

func nextPowerOfTwo(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

type SharedSPMCQueue struct {
	file *os.File
	mmap []byte
	size int
	mask uint64
	// Memory layout:
	// [0-7]:   head (write position)
	// [8-15]:  readCounter (number of reads for current position)
	// [16-23]: numConsumers (total number of consumers)
	// [24+]:   buffer data
}

func NewSharedSPMCQueue(name string, size int) (*SharedSPMCQueue, error) {
	// Size must be power of 2
	size = nextPowerOfTwo(size)
	mask := uint64(size - 1)

	// Calculate total size needed:
	// 24 bytes for head + readCounter + numConsumers
	// plus size * buffer entry size
	totalSize := 24 + (size * 65536) // Assuming max message size of 64KB

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	err = file.Truncate(int64(totalSize))
	if err != nil {
		file.Close()
		return nil, err
	}

	mmap, err := syscall.Mmap(int(file.Fd()), 0, totalSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, err
	}

	// Initialize the shared memory
	// Set head to 0
	atomic.StoreUint64((*uint64)(unsafe.Pointer(&mmap[0])), 0)
	// Set readCounter to 0
	atomic.StoreUint64((*uint64)(unsafe.Pointer(&mmap[8])), 0)
	// Set numConsumers to 0
	atomic.StoreUint64((*uint64)(unsafe.Pointer(&mmap[16])), 0)

	return &SharedSPMCQueue{
		file: file,
		mmap: mmap,
		size: size,
		mask: mask,
	}, nil
}

func (q *SharedSPMCQueue) Write(data []byte) error {
	head := atomic.LoadUint64((*uint64)(unsafe.Pointer(&q.mmap[0])))
	numConsumers := atomic.LoadUint64((*uint64)(unsafe.Pointer(&q.mmap[16])))

	// Wait until all consumers have read previous entry
	for atomic.LoadUint64((*uint64)(unsafe.Pointer(&q.mmap[8]))) < numConsumers {
		runtime.Gosched() // Yield to other goroutines while waiting
	}

	// Reset read counter for new data
	atomic.StoreUint64((*uint64)(unsafe.Pointer(&q.mmap[8])), 0)

	// Write data
	pos := 24 + (int(head&q.mask) * 65536)
	binary.LittleEndian.PutUint32(q.mmap[pos:pos+4], uint32(len(data)))
	copy(q.mmap[pos+4:pos+4+len(data)], data)

	// Increment head after write is complete
	atomic.AddUint64((*uint64)(unsafe.Pointer(&q.mmap[0])), 1)
	return nil
}

func (q *SharedSPMCQueue) Read() ([]byte, bool) {
	head := atomic.LoadUint64((*uint64)(unsafe.Pointer(&q.mmap[0])))
	readCount := atomic.AddUint64((*uint64)(unsafe.Pointer(&q.mmap[8])), 1)
	numConsumers := atomic.LoadUint64((*uint64)(unsafe.Pointer(&q.mmap[16])))

	// If we're the last reader, move to next entry
	if readCount == numConsumers {
		atomic.StoreUint64((*uint64)(unsafe.Pointer(&q.mmap[8])), 0)
	}

	pos := 24 + (int((head-1)&q.mask) * 65536)
	size := binary.LittleEndian.Uint32(q.mmap[pos : pos+4])

	data := make([]byte, size)
	copy(data, q.mmap[pos+4:pos+4+int(size)])

	return data, true
}

func (q *SharedSPMCQueue) RegisterConsumer() {
	atomic.AddUint64((*uint64)(unsafe.Pointer(&q.mmap[16])), 1)
}

func (q *SharedSPMCQueue) UnregisterConsumer() {
	atomic.AddUint64((*uint64)(unsafe.Pointer(&q.mmap[16])), ^uint64(0)) // Subtract 1
}
