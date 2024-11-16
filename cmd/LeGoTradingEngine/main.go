package main

import (
	"fmt"

	"github.com/Heian0/LeGoTradingEngine/internal/orderbook"
)

func main() {
	testOrder := orderbook.MarketAskOrder(0, 0, 10, orderbook.FillOrKill)
	testOrder.ExecuteOrder(10, 0)
	fmt.Println(testOrder.IsFilled())
}
