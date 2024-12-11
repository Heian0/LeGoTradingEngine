package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	exg "github.com/Heian0/LeGoTradingEngine/internal/exchange"
)

func main() {
	conn, err := grpc.NewClient(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Could not connect: %s", err)
	}

	defer conn.Close()

	exchangeServer := exg.NewExchangeServiceClient(conn)

	orderMsg := exg.OrderMessage{
		Command:              exg.Command_ADD,
		OrderType:            exg.OrderType_LIMIT,
		OrderSide:            exg.Side_ASK,
		OrderTimeInForce:     exg.OrderTimeInForce_GTC,
		Id:                   0,
		SymbolId:             0,
		Price:                30,
		StopPrice:            0,
		TrailingAmount:       0,
		LastExecutedPrice:    0,
		Quantity:             100,
		OpenQuantity:         0,
		LastExecutedQuantity: 0,
	}

	orderMsg1 := exg.OrderMessage{
		Command:              exg.Command_ADD,
		OrderType:            exg.OrderType_LIMIT,
		OrderSide:            exg.Side_ASK,
		OrderTimeInForce:     exg.OrderTimeInForce_GTC,
		Id:                   1,
		SymbolId:             0,
		Price:                40,
		StopPrice:            0,
		TrailingAmount:       0,
		LastExecutedPrice:    0,
		Quantity:             100,
		OpenQuantity:         0,
		LastExecutedQuantity: 0,
	}

	for {
		response, err := exchangeServer.HandleOrder(context.Background(), &orderMsg)
		if err != nil {
			fmt.Printf("Error handling order: %s", err)
		}
		fmt.Printf("Exchange State:\n %s\n", response.ExchangeStatus)

		response, err = exchangeServer.HandleOrder(context.Background(), &orderMsg1)
		if err != nil {
			fmt.Printf("Error handling order: %s", err)
		}
		fmt.Printf("Exchange State:\n %s\n", response.ExchangeStatus)

		time.Sleep(1 * time.Microsecond)
	}

	/*
		response, err = exchangeServer.HandleOrder(context.Background(), &orderMsg1)
		if err != nil {
			fmt.Printf("Error handling order: %s", err)
		}
		fmt.Printf("Exchange State:\n %s\n", response.ExchangeStatus)
	*/
}
