package exchange

import (
	ob "github.com/Heian0/LeGoTradingEngine/internal/orderbook"
)

type Exchange struct {
	orderBooks map[uint64]*ob.OrderBook
	symbolMap  map[uint64]*ob.Symbol
}

func (exchange *Exchange) AddSymbol(symbolId uint64, ticker string) {
	_, symbolExists := exchange.symbolMap[symbolId]
	if symbolExists {
		panic("Trying to insert a symbol which already exists")
	}
	newSymbol := ob.NewSymbol(symbolId, ticker)
	exchange.symbolMap[symbolId] = &newSymbol
}

func (exchange *Exchange) AddOrderbook(symbolId uint64) {
	symbol, symbolExists := exchange.symbolIdMap[symbolId]
	if symbolExists {
		panic("Trying to insert an orderbook which already exists")
	}
	exchange.symbolIdMap[symbolId] = ob.NewSymbol()
}
