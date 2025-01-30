package main

import (
	"fmt"
	"log"
	"net"
	"os"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
	"google.golang.org/grpc"
)

func main() {

	args := os.Args[1:]

	switch len(args) {
	case 0:
		fmt.Println("Please specify the exchange/server simulation you wish to run. 1 = Basic (SPY Orderbook only), 2 = Arbitrage Simulation.")
	case 1:
		if args[0] == "1" {

			fmt.Println("Running Basic Exchange. Please run the client side code.")

			exchange := exg.NewExchange()
			exchange.AddOrderbook(0, "SPY")

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

			// Setup UDP broadcast connection
			addr := &net.UDPAddr{
				IP:   net.IPv4(239, 0, 0, 1), // Multicast address
				Port: 8011,
			}

			udpConn, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				log.Fatalf("Failed to setup UDP broadcast: %v", err)
			}
			defer udpConn.Close()

			// Set up UDP broadcaster in the exchange
			exchange.SetupBroadcaster(udpConn)

			select {}

		} else if args[0] == "2" {
			fmt.Println("Running Arbitrage Simulation. Please run the client side code.")

			exchange1 := exg.NewExchange()
			exchange1.AddOrderbook(0, "LEBRON")

			exchange2 := exg.NewExchange()
			exchange2.AddOrderbook(0, "LEBRON")

			go func() {
				lis, err := net.Listen("tcp", ":9000")
				if err != nil {
					log.Fatalf("Failed to listen on port 9000: %v", err)
				}

				grpcServer := grpc.NewServer()
				exg.RegisterExchangeServiceServer(grpcServer, exchange1)

				if err := grpcServer.Serve(lis); err != nil {
					log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
				}
			}()

			go func() {
				lis, err := net.Listen("tcp", ":9001")
				if err != nil {
					log.Fatalf("Failed to listen on port 9001: %v", err)
				}

				grpcServer := grpc.NewServer()
				exg.RegisterExchangeServiceServer(grpcServer, exchange2)

				if err := grpcServer.Serve(lis); err != nil {
					log.Fatalf("Failed to serve gRPC server over port 9001: %v", err)
				}
			}()

			// Setup UDP broadcast connection
			addr1 := &net.UDPAddr{
				IP:   net.IPv4(239, 0, 0, 1), // Multicast address
				Port: 8011,
			}

			udpConn1, err := net.DialUDP("udp", nil, addr1)
			if err != nil {
				log.Fatalf("Failed to setup UDP broadcast: %v", err)
			}
			defer udpConn1.Close()

			// Setup UDP broadcast connection
			addr2 := &net.UDPAddr{
				IP:   net.IPv4(239, 0, 0, 1), // Multicast address
				Port: 8012,
			}

			udpConn2, err := net.DialUDP("udp", nil, addr2)
			if err != nil {
				log.Fatalf("Failed to setup UDP broadcast: %v", err)
			}
			defer udpConn2.Close()

			exchange1.SetupBroadcaster(udpConn1)
			exchange2.SetupBroadcaster(udpConn2)

			select {}

		} else {
			fmt.Println("Please specify a valid exchange/server simulation you wish to run. 1 = Basic (SPY Orderbook only), 2 = Arbitrage Simulation.")
		}
	}
}
