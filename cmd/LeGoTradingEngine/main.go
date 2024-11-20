package main

import (
	"fmt"

	"github.com/Heian0/LeGoTradingEngine/internal/orderbook"
)

func main() {
	ob1 := orderbook.NewOrderbook(999)
	ob2 := orderbook.NewOrderbook(111)
	newOrder1 := orderbook.LimitAskOrder(1, 999, 100, 10, orderbook.GoodTillCancel)
	newOrder2 := orderbook.LimitAskOrder(1, 111, 100, 0, orderbook.GoodTillCancel)
	ob1.AddMarketOrder(&newOrder1)
	ob2.AddMarketOrder(&newOrder2)

	newOrder11 := orderbook.LimitBidOrder(1, 999, 100, 1, orderbook.GoodTillCancel)
	ob1.AddMarketOrder(&newOrder11)

	fmt.Println(ob1.String())
	fmt.Println("---------------------")
	fmt.Println(ob2.String())
}
