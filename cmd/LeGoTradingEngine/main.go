package main

import (
	"fmt"

	"github.com/Heian0/LeGoTradingEngine/internal/orderbook"
)

func main() {
	ob := orderbook.NewOrderbook(999)

	newAskLimitOrder1 := orderbook.LimitAskOrder(1, 999, 100, 30, orderbook.GoodTillCancel)
	ob.AddOrder(&newAskLimitOrder1)
	fmt.Println(ob.String())

	newBidLimitOrder1 := orderbook.LimitBidOrder(2, 999, 50, 10, orderbook.GoodTillCancel)
	ob.AddOrder(&newBidLimitOrder1)
	fmt.Println(ob.String())

	newBidStopOrder1 := orderbook.StopBidOrder(3, 999, 8, 20, orderbook.GoodTillCancel)
	ob.AddOrder(&newBidStopOrder1)
	fmt.Println(ob.String())

	newAskLimitOrder2 := orderbook.LimitAskOrder(4, 999, 20, 5, orderbook.GoodTillCancel)
	ob.AddOrder(&newAskLimitOrder2)
	fmt.Println(ob.String())
}
