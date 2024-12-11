package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
)

func main() {

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Failed to listen on port 9000: %v", err)
	}

	exchange := exg.NewExchange()
	exchange.AddOrderbook(0, "TSLA")

	grpcServer := grpc.NewServer()

	exg.RegisterExchangeServiceServer(grpcServer, exchange)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
	}
}
