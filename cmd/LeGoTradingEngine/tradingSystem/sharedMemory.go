package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type SharedMemory struct {
	file       *os.File
	mmap       []byte
	bufferSize int
	numBuffers int
	// First 8 bytes for write index, next 8 for read index
	// Then numBuffers * bufferSize for actual data
}

func NewSharedMemory(name string, bufferSize int, numBuffers int) (*SharedMemory, error) {
	// Total size: 16 bytes for indexes + (bufferSize * numBuffers) for data
	totalSize := 16 + (bufferSize * numBuffers)

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	err = file.Truncate(int64(totalSize))
	if err != nil {
		return nil, err
	}

	mmap, err := syscall.Mmap(int(file.Fd()), 0, totalSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &SharedMemory{
		file:       file,
		mmap:       mmap,
		bufferSize: bufferSize,
		numBuffers: numBuffers,
	}, nil
}

func (sm *SharedMemory) Write(data []byte) error {
	if len(data) > sm.bufferSize {
		return fmt.Errorf("data too large for buffer")
	}

	// Get current write index atomically
	writeIndex := atomic.LoadUint64((*uint64)(unsafe.Pointer(&sm.mmap[0])))

	// Calculate buffer position
	bufferStart := 16 + (int(writeIndex%uint64(sm.numBuffers)) * sm.bufferSize)

	// Write data
	copy(sm.mmap[bufferStart:bufferStart+len(data)], data)

	// Update write index atomically
	atomic.AddUint64((*uint64)(unsafe.Pointer(&sm.mmap[0])), 1)

	return nil
}

func (sm *SharedMemory) Close() error {
	if err := syscall.Munmap(sm.mmap); err != nil {
		return err
	}
	return sm.file.Close()
}
