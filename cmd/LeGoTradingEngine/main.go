package main

import {
	"fmt",
	"github.com/Heian0/LeGoTradingEngine/internal/orderbook"
}

func main() {
	order := orderbook.MarketAskOrder(0, 0, 10, orderbook.OrderTimeInForce.FillOrKill)
}