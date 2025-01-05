package main

import (
	"log"
	"net"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
	"google.golang.org/grpc"
)

func main() {

	exchange := exg.NewExchange()
	exchange.AddOrderbook(0, "TSLA")

	// First server for order processing (original)
	go func() {
		lis, err := net.Listen("tcp", ":9000")
		if err != nil {
			log.Fatalf("Failed to listen on port 9000: %v", err)
		}

		grpcServer := grpc.NewServer()
		exg.RegisterExchangeServiceServer(grpcServer, exchange)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
		}
	}()

	// Second server for subscribers
	lis2, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("Failed to listen on port 9001: %v", err)
	}

	grpcServer2 := grpc.NewServer()
	exg.RegisterExchangeServiceServer(grpcServer2, exchange) // Use same exchange instance

	if err := grpcServer2.Serve(lis2); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9001: %v", err)
	}

}
