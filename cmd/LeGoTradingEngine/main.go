package main

import (
	"fmt"

	"github.com/Heian0/LeGoTradingEngine/internal/orderbook"
)

func main() {
	lvlMap := orderbook.NewLevelMap()
	lvlMap.Put(0, orderbook.NewLevel(orderbook.Ask, 100, 420))
	lvlMap.Put(3, orderbook.NewLevel(orderbook.Ask, 100, 420))
	lvlMap.Emplace(2, orderbook.Bid, 788)
	fmt.Println(lvlMap.String())
}
