package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
	zmq4 "github.com/pebbe/zmq4"
)

type Client struct {
	conn      *grpc.ClientConn
	client    exg.ExchangeServiceClient
	publisher *zmq4.Socket
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	// Setup ZMQ publisher
	publisher, err := zmq4.NewSocket(1)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %v", err)
	}

	err = publisher.Bind("tcp://*:5555")
	if err != nil {
		conn.Close()
		publisher.Close()
		return nil, fmt.Errorf("failed to bind publisher: %v", err)
	}

	return &Client{
		conn:      conn,
		client:    exg.NewExchangeServiceClient(conn),
		publisher: publisher,
	}, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.publisher != nil {
		c.publisher.Close()
	}
}

func (c *Client) SubscribeToOrderBook(symbolId uint64) error {
	req := &exg.SubscribeRequest{
		SymbolId: symbolId,
	}

	stream, err := c.client.SubscribeToOrderBook(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %v", err)
	}

	log.Println("Successfully subscribed to orderbook stream")
	updateCount := 0
	startTime := time.Now()

	for {
		log.Println("Waiting for next update...")
		state, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("error receiving update: %v", err)
		}

		// Serialize the state using protobuf
		data, err := proto.Marshal(state)
		if err != nil {
			log.Printf("Error marshaling state: %v", err)
			continue
		}

		// Publish protobuf bytes through ZMQ
		_, err = c.publisher.SendBytes(data, 0)
		if err != nil {
			log.Printf("Error publishing update: %v", err)
			continue
		}

		updateCount++
		elapsed := time.Since(startTime)
		log.Printf("Update #%d received after %v", updateCount, elapsed)
		log.Printf("Orderbook state - Bids: %d, Asks: %d", len(state.Bids), len(state.Asks))

		displayOrderBook(state)
	}
}

func displayOrderBook(state *exg.OrderBookState) {
	fmt.Print(state.ObsToString())
}

func main() {
	client, err := NewClient("localhost:9001")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.SubscribeToOrderBook(0); err != nil {
		log.Fatalf("Subscription error: %v", err)
	}
}
