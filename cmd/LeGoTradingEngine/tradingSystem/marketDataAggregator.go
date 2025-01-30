package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

	"google.golang.org/protobuf/proto"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
)

type MultiMarketDataAggregator struct {
	udpConn1 *net.UDPConn
	udpConn2 *net.UDPConn
	shm      *SharedMemory
	queue    *SharedSPMCQueue
}

func NewMultiMarketDataAggregator(port1 int, port2 int) (*MultiMarketDataAggregator, error) {

	// Setup UDP multicast listeners
	addr1 := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: port1,
	}

	udpconn1, err := net.ListenUDP("udp", addr1)
	if err != nil {
		return nil, fmt.Errorf("failed to setup UDP listener: %v", err)
	}

	err = udpconn1.SetReadBuffer(65536) // Set buffer
	if err != nil {
		log.Printf("Failed to set read buffer: %v", err)
	}

	group1 := net.IPv4(239, 0, 0, 1)
	mreq1 := make([]byte, 8)
	copy(mreq1, group1.To4())
	copy(mreq1[4:], net.IPv4(0, 0, 0, 0).To4())

	fd1, err := udpconn1.File()
	if err != nil {
		return nil, fmt.Errorf("failed to get socket fd: %v", err)
	}
	defer fd1.Close()

	err = syscall.SetsockoptString(int(fd1.Fd()), syscall.IPPROTO_IP, syscall.IP_ADD_MEMBERSHIP, string(mreq1))
	if err != nil {
		log.Printf("Failed to join multicast group: %v", err)
	}

	addr2 := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: port2,
	}

	udpconn2, err := net.ListenUDP("udp", addr2)
	if err != nil {
		return nil, fmt.Errorf("failed to setup UDP listener: %v", err)
	}

	err = udpconn2.SetReadBuffer(65536) // Set buffer
	if err != nil {
		log.Printf("Failed to set read buffer: %v", err)
	}

	group2 := net.IPv4(239, 0, 0, 1)
	mreq2 := make([]byte, 8)
	copy(mreq2, group2.To4())
	copy(mreq2[4:], net.IPv4(0, 0, 0, 0).To4())

	fd2, err := udpconn2.File()
	if err != nil {
		return nil, fmt.Errorf("failed to get socket fd: %v", err)
	}
	defer fd2.Close()

	err = syscall.SetsockoptString(int(fd2.Fd()), syscall.IPPROTO_IP, syscall.IP_ADD_MEMBERSHIP, string(mreq2))
	if err != nil {
		log.Printf("Failed to join multicast group: %v", err)
	}

	shm, err := NewSharedMemory("/tmp/MarketDataAggregatorMem", 65536, 4)
	if err != nil {
		log.Fatalf("Failed to create shared memory: %v", err)
	}

	sharedSPMCqueue, err := NewSharedSPMCQueue("/tmp/marketdata_queue", 1024)
	if err != nil {
		return nil, err
	}

	return &MultiMarketDataAggregator{
		udpConn1: udpconn1,
		udpConn2: udpconn2,
		shm:      shm,
		queue:    sharedSPMCqueue,
	}, nil
}

type BasicMarketDataAggregator struct {
	udpConn *net.UDPConn
	shm     *SharedMemory
	queue   *SharedSPMCQueue
}

func NewBasicMarketDataAggregator(port int) (*BasicMarketDataAggregator, error) {

	// Setup UDP multicast listener
	addr := &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: port,
	}

	udpconn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to setup UDP listener: %v", err)
	}

	err = udpconn.SetReadBuffer(65536) // Set buffer
	if err != nil {
		log.Printf("Failed to set read buffer: %v", err)
	}

	group := net.IPv4(239, 0, 0, 1)
	mreq := make([]byte, 8)
	copy(mreq, group.To4())
	copy(mreq[4:], net.IPv4(0, 0, 0, 0).To4())

	fd, err := udpconn.File()
	if err != nil {
		return nil, fmt.Errorf("failed to get socket fd: %v", err)
	}
	defer fd.Close()

	err = syscall.SetsockoptString(int(fd.Fd()), syscall.IPPROTO_IP, syscall.IP_ADD_MEMBERSHIP, string(mreq))
	if err != nil {
		log.Printf("Failed to join multicast group: %v", err)
	}

	shm, err := NewSharedMemory("/tmp/MarketDataAggregatorMem", 65536, 4)
	if err != nil {
		log.Fatalf("Failed to create shared memory: %v", err)
	}

	sharedSPMCqueue, err := NewSharedSPMCQueue("/tmp/marketdata_queue", 1024)
	if err != nil {
		return nil, err
	}

	return &BasicMarketDataAggregator{
		udpConn: udpconn,
		shm:     shm,
		queue:   sharedSPMCqueue,
	}, nil
}

func (mda *BasicMarketDataAggregator) Close() {
	if mda.udpConn != nil {
		mda.udpConn.Close()
	}
	if mda.shm != nil {
		mda.shm.Close()
	}
}

func (mda *BasicMarketDataAggregator) ListenForUpdates() error {
	buffer := make([]byte, 65536)
	state := &exg.OrderBookState{}

	log.Println("Starting to listen for orderbook updates...")
	for {
		n, remoteAddr, err := mda.udpConn.ReadFromUDP(buffer)
		if err != nil {
			return fmt.Errorf("error reading UDP: %v", err)
		}

		log.Printf("Received %d bytes from %v", n, remoteAddr)
		//log.Printf("Raw data: %v", buffer[:n])

		receiveTime := time.Now().UnixNano()

		state.Reset()
		err = proto.Unmarshal(buffer[:n], state)
		if err != nil {
			log.Printf("Error unmarshaling state: %v", err)
			log.Printf("This might not be a protobuf message, raw content: %s", string(buffer[:n]))
			continue
		}

		// Write to shared memory
		data, err := proto.Marshal(state)
		if err != nil {
			log.Printf("Error marshaling data: %v", err)
			continue
		}

		// Write size first, then data
		binary.LittleEndian.PutUint32(mda.shm.mmap[0:4], uint32(len(data)))
		copy(mda.shm.mmap[4:], data)

		mda.queue.Write(data)

		latencyNs := receiveTime - state.Timestamp
		latencyMs := float64(latencyNs) / 1_000_000 // Convert to milliseconds

		log.Printf("Latency: %.3f ms", latencyMs)
		displayOrderBook(state)
	}
}

func (mda *BasicMarketDataAggregator) Subscribe() {
	mda.queue.RegisterConsumer()
}
