package main

import (
	"fmt"

	"github.com/Heian0/LeGoTradingEngine/internal/orderbook"
)

func main() {
	ob := orderbook.NewOrderbook(999)

	newAskLimitOrder1 := orderbook.LimitAskOrder(1, 999, 100, 10, orderbook.GoodTillCancel)
	newAskLimitOrder2 := orderbook.LimitAskOrder(2, 999, 100, 8, orderbook.GoodTillCancel)
	newAskLimitOrder3 := orderbook.LimitAskOrder(3, 999, 100, 9, orderbook.GoodTillCancel)

	newBidLimitOrder1 := orderbook.LimitBidOrder(4, 999, 50, 10, orderbook.GoodTillCancel)
	newBidMarketOrder1 := orderbook.MarketBidOrder(5, 999, 20, orderbook.ImmediateOrCancel)

	newBidStopOrder1 := orderbook.StopBidOrder(6, 999, 10, 5, orderbook.FillOrKill)

	ob.AddLimitOrder(&newAskLimitOrder1)
	ob.AddLimitOrder(&newAskLimitOrder2)
	ob.AddLimitOrder(&newAskLimitOrder3)
	fmt.Println(ob.String())

	ob.AddLimitOrder(&newBidLimitOrder1)
	fmt.Println(ob.String())

	ob.AddMarketOrder(&newBidMarketOrder1)
	fmt.Println(ob.String())

	ob.AddStopOrder(&newBidStopOrder1)
	fmt.Println(ob.String())

}
