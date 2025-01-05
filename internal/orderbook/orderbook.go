package orderbook

import (
	"container/list"
	"fmt"
	"math"
	"strconv"
	"strings"

	rbt "github.com/Heian0/LeGoTradingEngine/internal/utils/redblacktree"
	simplemath "github.com/Heian0/LeGoTradingEngine/internal/utils/simplemath"
)

type OrderBook struct {
	symbolId              uint64
	lastExecutedPrice     uint64
	trailingBidPrice      uint64
	trailingAskPrice      uint64
	orders                map[uint64]*Order
	askLevels             *LevelMap
	bidLevels             *LevelMap
	stopAskLevels         *LevelMap
	stopBidLevels         *LevelMap
	trailingStopAskLevels *LevelMap
	trailingStopBidLevels *LevelMap
}

func NewOrderbook(_symbolId uint64) *OrderBook {
	return &OrderBook{
		symbolId:              _symbolId,
		trailingAskPrice:      math.MaxUint64,
		orders:                make(map[uint64]*Order),
		askLevels:             NewLevelMap(),
		bidLevels:             NewLevelMap(),
		stopAskLevels:         NewLevelMap(),
		stopBidLevels:         NewLevelMap(),
		trailingStopAskLevels: NewLevelMap(),
		trailingStopBidLevels: NewLevelMap(),
	}
}

func (orderBook *OrderBook) LastExecutedPriceBid() uint64 {
	return orderBook.lastExecutedPrice
}

func (orderBook *OrderBook) LastExecutedPriceAsk() uint64 {
	if orderBook.lastExecutedPrice == 0 {
		return math.MaxUint64
	} else {
		return orderBook.lastExecutedPrice
	}
}

func (orderBook *OrderBook) AddOrder(order *Order) {
	switch order.orderType {
	case Market:
		orderBook.AddMarketOrder(order)
	case Limit:
		orderBook.AddLimitOrder(order)
	case Stop, StopLimit, TrailingStop, TrailingStopLimit:
		orderBook.AddStopOrder(order)
	}
	orderBook.ActivateStopOrders()
	orderBook.ValidateOrderbook()
}

func (orderBook *OrderBook) AddMarketOrder(order *Order) {
	if order.IsAsk() {
		order.price = 0
	} else {
		order.price = math.MaxUint64
	}
	orderBook.Match(order)
}

func (orderBook *OrderBook) AddLimitOrder(order *Order) {
	orderBook.Match(order)
	if !order.IsFilled() && !order.IsFillOrKill() {
		orderBook.InsertLimitOrder(order)
	} else {
		// eventHandler.handleorderdeleted...
		return
	}
}

func (orderBook *OrderBook) AddStopOrder(order *Order) {
	if order.IsTrailingStop() || order.IsTrailingStopLimit() {
		orderBook.CalculateStopPrice(order)
	}
	var marketPrice uint64 = 0
	if order.IsAsk() {
		marketPrice = orderBook.LastExecutedPriceBid()
	} else {
		marketPrice = orderBook.LastExecutedPriceAsk()
	}
	match1 := (order.IsAsk() && marketPrice <= order.stopPrice)
	match2 := (order.IsBid() && marketPrice >= order.stopPrice)
	if match1 || match2 {
		if order.IsStop() || order.IsTrailingStop() {
			order.orderType = Market
		} else {
			order.orderType = Limit
		}
		order.stopPrice = 0
		order.trailingAmount = 0
		// Handle order updated
		if order.IsMarket() {
			orderBook.AddMarketOrder(order)
		} else {
			orderBook.AddLimitOrder(order)
		}
		return
	}
	if order.IsTrailingStop() || order.IsTrailingStopLimit() {
		orderBook.InsertTrailingStopOrder(order)
	} else {
		orderBook.InsertStopOrder(order)
	}
}

// Emplace with hint breaking order logic, fix later
func (orderBook *OrderBook) InsertLimitOrder(order *Order) {
	var lvlPtr *Level
	if order.IsAsk() {
		lvlPtr = orderBook.askLevels.Emplace(order.price, Ask, order.symbolId)
	} else {
		lvlPtr = orderBook.bidLevels.Emplace(order.price, Bid, order.symbolId)
	}
	order.levelPtr = lvlPtr
	orderBook.orders[order.id] = order
	lvlPtr.AddOrder(order)
}

func (orderBook *OrderBook) InsertStopOrder(order *Order) {
	var lvlPtr *Level
	if order.IsAsk() {
		lvlPtr = orderBook.stopAskLevels.Emplace(order.stopPrice, Ask, order.symbolId)
	} else {
		lvlPtr = orderBook.stopBidLevels.Emplace(order.stopPrice, Bid, order.symbolId)
	}
	order.levelPtr = lvlPtr
	orderBook.orders[order.id] = order
	lvlPtr.AddOrder(order)
}

func (orderBook *OrderBook) InsertTrailingStopOrder(order *Order) {
	var lvlPtr *Level
	if order.IsAsk() {
		lvlPtr = orderBook.trailingStopAskLevels.Emplace(order.price, Ask, order.symbolId)
	} else {
		lvlPtr = orderBook.trailingStopAskLevels.Emplace(order.price, Bid, order.symbolId)
	}
	order.levelPtr = lvlPtr
	orderBook.orders[order.id] = order
	lvlPtr.AddOrder(order)
}

func (orderBook *OrderBook) CalculateStopPrice(order *Order) uint64 {
	if order.IsAsk() {
		marketPrice := orderBook.LastExecutedPriceBid()
		trailAmount := order.trailingAmount
		// Set new Stop Price for Ask order
		if marketPrice > trailAmount {
			order.stopPrice = marketPrice - trailAmount
		} else {
			order.stopPrice = 0
		}
	} else {
		marketPrice := orderBook.LastExecutedPriceAsk()
		trailAmount := order.trailingAmount
		// Set new Stop Price for Bid order
		if marketPrice < (math.MaxUint64 - trailAmount) {
			order.stopPrice = marketPrice + trailAmount
		} else {
			order.stopPrice = math.MaxUint64
		}
	}
	return order.stopPrice
}

func (orderBook *OrderBook) ActivateStopOrders() {
	activated := true
	for activated {
		activated = orderBook.ActivateBidStopOrders()
		orderBook.UpdateAskStopOrders()
		activated = orderBook.ActivateAskStopOrders() || activated
		orderBook.UpdateBidStopOrders()
	}
}

func (orderBook *OrderBook) ActivateBidStopOrders() bool {
	var activated_orders bool = false

	// orderBook.stopBidLevels.Iterator is now at the beginning
	orderBook.stopBidLevels.SetMapBegin()
	stopLevelsIt := orderBook.stopBidLevels.levelMapIterator

	for stopLevelsIt.Next() && stopLevelsIt.Key().(uint64) <= orderBook.LastExecutedPriceAsk() {
		activated_orders = true
		var currStopOrder *Order = stopLevelsIt.Value().(*Order)
		orderBook.ActivateStopOrder(*currStopOrder)
		orderBook.stopBidLevels.SetMapBegin()
		stopLevelsIt = orderBook.stopBidLevels.levelMapIterator
	}

	// orderBook.stopBidLevels.Iterator is now at the beginning
	orderBook.trailingStopBidLevels.SetMapBegin()
	trailingStopLevelsIt := orderBook.trailingStopBidLevels.levelMapIterator

	for trailingStopLevelsIt.Next() && trailingStopLevelsIt.Key().(uint64) <= orderBook.LastExecutedPriceAsk() {
		activated_orders = true
		var currTrailingStopOrder *Order = trailingStopLevelsIt.Value().(*Order)
		orderBook.ActivateStopOrder(*currTrailingStopOrder)
		orderBook.trailingStopBidLevels.SetMapBegin()
		trailingStopLevelsIt = orderBook.trailingStopBidLevels.levelMapIterator
	}
	return activated_orders
}

func (orderBook *OrderBook) ActivateAskStopOrders() bool {
	var activated_orders bool = false

	// orderBook.stopAskLevels.Iterator is now at the end
	orderBook.stopAskLevels.SetMapEnd()
	stopLevelsIt := orderBook.stopAskLevels.levelMapIterator

	for stopLevelsIt.Prev() && stopLevelsIt.Key().(uint64) >= orderBook.LastExecutedPriceBid() {
		activated_orders = true
		var currStopOrder *Order = stopLevelsIt.Value().(*Order)
		orderBook.ActivateStopOrder(*currStopOrder)
		orderBook.stopAskLevels.SetMapEnd()
		stopLevelsIt = orderBook.stopAskLevels.levelMapIterator
	}

	// orderBook.stopAskLevels.Iterator is now at the end
	orderBook.trailingStopBidLevels.SetMapEnd()
	trailingStopLevelsIt := orderBook.trailingStopAskLevels.levelMapIterator

	for trailingStopLevelsIt.Prev() && trailingStopLevelsIt.Key().(uint64) >= orderBook.LastExecutedPriceBid() {
		activated_orders = true
		var currTrailingStopOrder *Order = trailingStopLevelsIt.Value().(*Order)
		orderBook.ActivateStopOrder(*currTrailingStopOrder)
		orderBook.trailingStopAskLevels.SetMapEnd()
		trailingStopLevelsIt = orderBook.trailingStopAskLevels.levelMapIterator
	}
	return activated_orders
}

// Passing by value because we want to delete by ID and potentially add a new market order
func (orderBook *OrderBook) ActivateStopOrder(order Order) {
	orderBook.DeleteOrder(order.id, false)
	order.stopPrice = 0
	order.trailingAmount = 0
	// Convert to limit/market
	if order.IsStop() || order.IsTrailingStop() {
		order.orderType = Market
		//Handle order updated
		orderBook.AddMarketOrder(&order)
	} else {
		order.orderType = Limit
		//Handle order updated
		orderBook.AddLimitOrder(&order)
	}
}

func (orderBook *OrderBook) UpdateBidStopOrders() {
	if orderBook.trailingAskPrice <= orderBook.LastExecutedPriceAsk() || orderBook.trailingStopBidLevels.IsEmpty() {
		orderBook.trailingAskPrice = orderBook.lastExecutedPrice
		return
	}
	newTrailingStopBidLevels := NewLevelMap()
	orderBook.trailingStopBidLevels.SetMapBegin()
	trailingStopLevelsIt := orderBook.trailingStopBidLevels.levelMapIterator

	for trailingStopLevelsIt.Next() {
		for !trailingStopLevelsIt.Value().(*Level).Empty() {
			var order *Order = trailingStopLevelsIt.Value().(*Level).Front()
			newStopPrice := orderBook.CalculateStopPrice(order)
			newPtr := newTrailingStopBidLevels.EmplaceWithHint(newStopPrice, Bid, order.symbolId, orderBook.trailingStopBidLevels.GetMapBegin())
			orderBook.orders[order.id].levelPtr = newPtr
			//Handle Order updated
		}
	}
	orderBook.trailingStopBidLevels = newTrailingStopBidLevels
	orderBook.trailingAskPrice = orderBook.lastExecutedPrice
}

func (orderBook *OrderBook) UpdateAskStopOrders() {
	if orderBook.trailingBidPrice <= orderBook.LastExecutedPriceBid() || orderBook.trailingStopBidLevels.IsEmpty() {
		orderBook.trailingBidPrice = orderBook.lastExecutedPrice
		return
	}
	newTrailingStopAskLevels := NewLevelMap()
	orderBook.trailingStopAskLevels.SetMapBegin()
	trailingStopLevelsIt := orderBook.trailingStopAskLevels.levelMapIterator

	for trailingStopLevelsIt.Next() {
		for !trailingStopLevelsIt.Value().(*Level).Empty() {
			var order *Order = trailingStopLevelsIt.Value().(*Level).Front()
			newStopPrice := orderBook.CalculateStopPrice(order)
			newPtr := newTrailingStopAskLevels.EmplaceWithHint(newStopPrice, Bid, order.symbolId, orderBook.trailingStopBidLevels.GetMapBegin())
			orderBook.orders[order.id].levelPtr = newPtr
			//Handle Order updated
		}
	}
	orderBook.trailingStopBidLevels = newTrailingStopAskLevels
	orderBook.trailingAskPrice = orderBook.lastExecutedPrice
}

func (orderBook *OrderBook) DelOrder(orderId uint64) {
	orderBook.DeleteOrder(orderId, true)
	orderBook.ActivateStopOrders()
	orderBook.ValidateOrderbook()
}

// Use this if we dont want to immediately try activating stop orders
func (orderBook *OrderBook) DeleteOrder(orderId uint64, noti bool) {
	order := orderBook.orders[orderId]
	if order.orderType == Market {
		fmt.Println("Should not be trying to delete a market order")
		return
	}
	level := order.levelPtr

	if noti {
		// Handle order deleted
	}

	level.DeleteOrder(order)
	if level.Empty() {
		switch order.orderType {
		case Limit:
			if order.IsAsk() {
				orderBook.askLevels.Delete(level.price)
			} else {
				orderBook.bidLevels.Delete(level.price)
			}
		case Stop:
		case StopLimit:
			if order.IsAsk() {
				orderBook.stopAskLevels.Delete(level.price)
			} else {
				orderBook.stopBidLevels.Delete(level.price)
			}
		case TrailingStop:
		case TrailingStopLimit:
			if order.IsAsk() {
				orderBook.trailingStopAskLevels.Delete(level.price)
			} else {
				orderBook.trailingStopBidLevels.Delete(level.price)
			}
		default:
			panic("Code should never reach this point, you are trying to delete a market order")
		}
		delete(orderBook.orders, orderId)
	}

}

func (orderBook *OrderBook) ReplaceOrder(orderId uint64, newOrderId uint64, newPrice uint64) {
	order := orderBook.orders[orderId]
	newOrder := *order
	newOrder.id = newOrderId
	if order.IsStop() || order.IsStopLimit() || order.IsTrailingStop() {
		newOrder.stopPrice = newPrice

	} else {
		newOrder.price = newPrice
	}
	orderBook.DeleteOrder(orderId, true)
	orderBook.AddOrder(&newOrder)
	orderBook.ActivateStopOrders()
	orderBook.ValidateOrderbook()
}

func (orderBook *OrderBook) Match(order *Order) {
	if order.IsFillOrKill() && !orderBook.CanMatch(order) {
		fmt.Println("Order is a Fill or Kill that could not be matched")
		return
	}
	print("order being matched\n")
	if order.IsAsk() {
		orderBook.bidLevels.SetMapEnd()
		bidLevelsIt := orderBook.bidLevels.levelMapIterator
		askOrder := order
		for bidLevelsIt.Prev() && bidLevelsIt.Key().(uint64) >= askOrder.price && !askOrder.IsFilled() {
			bidLevel := bidLevelsIt.Value().(*Level)
			bidOrder := bidLevel.Front()
			executingPrice := bidOrder.price
			orderBook.ExecuteOrders(askOrder, bidOrder, executingPrice)
			bidLevel.ReduceVolume(bidOrder.lastExecutedQuantity)
			if bidOrder.IsFilled() {
				orderBook.DeleteOrder(bidOrder.id, true)
			}
			orderBook.bidLevels.SetMapEnd()
		}
	}
	if order.IsBid() {
		orderBook.askLevels.SetMapBegin()
		askLevelsIt := orderBook.askLevels.levelMapIterator
		bidOrder := order
		for askLevelsIt.Next() && askLevelsIt.Key().(uint64) <= bidOrder.price && !bidOrder.IsFilled() {
			askLevel := askLevelsIt.Value().(*Level)
			askOrder := askLevel.Front()
			executingPrice := askOrder.price
			orderBook.ExecuteOrders(askOrder, bidOrder, executingPrice)
			askLevel.ReduceVolume(askOrder.lastExecutedQuantity)
			if askOrder.IsFilled() {
				orderBook.DeleteOrder(askOrder.id, true)
			}
		}
		orderBook.askLevels.SetMapBegin()
	}
}

func (orderBook *OrderBook) ExecuteOrderWithSpecifiedPrice(orderId uint64, quantity uint64, price uint64) {
	order := orderBook.orders[orderId]
	executingLevel := order.levelPtr
	executingQuantity := simplemath.Min(quantity, order.GetOpenQuantity())
	order.ExecuteOrder(executingQuantity, price)
	orderBook.lastExecutedPrice = price
	// Handle order executed
	executingLevel.ReduceVolume(order.GetLastExecutedQuantity())
	if order.IsFilled() {
		orderBook.DeleteOrder(orderId, true)
	}
	orderBook.ActivateStopOrders()
	orderBook.ValidateOrderbook()
}

// For market orders
func (orderBook *OrderBook) ExecuteOrderWithoutSpecifiedPrice(orderId uint64, quantity uint64) {
	order := orderBook.orders[orderId]
	executingLevel := order.levelPtr
	executingQuantity := simplemath.Min(quantity, order.GetOpenQuantity())
	price := order.GetPrice()
	order.ExecuteOrder(executingQuantity, price)
	orderBook.lastExecutedPrice = price
	// Handle order executed
	executingLevel.ReduceVolume(order.GetLastExecutedQuantity())
	if order.IsFilled() {
		orderBook.DeleteOrder(orderId, true)
	}
	orderBook.ActivateStopOrders()
	orderBook.ValidateOrderbook()
}

func (orderBook *OrderBook) CancelOrder(orderId uint64, cancellingQuantity uint64) {
	order := orderBook.orders[orderId]
	cancellingLevel := order.levelPtr
	preCancelQuantity := order.GetOpenQuantity()
	order.ReduceQuantity(cancellingQuantity)
	// Handle order cancelled
	cancellingLevel.ReduceVolume(preCancelQuantity - order.GetOpenQuantity())
	if order.IsFilled() {
		orderBook.DeleteOrder(orderId, true)
	}
	orderBook.ActivateStopOrders()
	orderBook.ValidateOrderbook()
}

func (orderBook *OrderBook) ExecuteOrders(askOrder *Order, bidOrder *Order, executingPrice uint64) {
	matchedQuantity := simplemath.Min(askOrder.openQuantity, bidOrder.openQuantity)
	askOrder.ExecuteOrder(matchedQuantity, executingPrice)
	bidOrder.ExecuteOrder(matchedQuantity, executingPrice)
	// Handle Ask order execution
	// Handle Bid order execution
	print("Order being executed at price:\n")
	print(executingPrice)
	print("\n")
	orderBook.lastExecutedPrice = executingPrice
}

func (orderBook *OrderBook) CanMatch(order *Order) bool {
	var availableQuantity uint64 = 0
	if order.IsAsk() {
		orderBook.bidLevels.SetMapEnd()
		bidLevelsIt := orderBook.bidLevels.levelMapIterator
		for bidLevelsIt.Prev() && bidLevelsIt.Key().(uint64) >= order.price {
			var quantityNeeded uint64 = order.openQuantity - availableQuantity
			availableQuantity += simplemath.Min(quantityNeeded, bidLevelsIt.Value().(*Level).volume)
			if availableQuantity >= order.openQuantity {
				return true
			}
		}
	} else {
		orderBook.askLevels.SetMapBegin()
		askLevelsIt := orderBook.askLevels.levelMapIterator
		for askLevelsIt.Next() && askLevelsIt.Key().(uint64) <= order.price {
			var quantityNeeded uint64 = order.openQuantity - availableQuantity
			availableQuantity += simplemath.Min(quantityNeeded, askLevelsIt.Value().(*Level).volume)
			if availableQuantity >= order.openQuantity {
				return true
			}
		}
	}
	return false
}

func (orderBook *OrderBook) ValidateOrderbook() {
	orderBook.ValidateLimitOrders()
	orderBook.ValidateStopOrders()
	orderBook.ValidateTrailingStopOrders()
}

/*
else {
	currBestAsk = orderBook.askLevels.GetMapBegin().Value.(*Level).price
}
if orderBook.bidLevels.IsEmpty() {
	currBestBid = 0
} else {
	currBestBid = orderBook.bidLevels.GetMapEnd().Value.(*Level).price
}
*/

func (orderBook *OrderBook) GetBestBid() *Level {
	bestBid := orderBook.bidLevels.GetMapEnd()
	// Default
	if bestBid == nil {
		print("No Bid orders exist - a default value was used\n")
		return &Level{
			levelSide: Bid,
			price:     0,
			symbolId:  0,
			volume:    0,
			orders:    list.New(),
		}
	}
	return bestBid.Value.(*Level)
}

func (orderBook *OrderBook) GetBestAsk() *Level {

	bestAsk := orderBook.askLevels.GetMapBegin()

	// Default - can cosider logging defaults
	if bestAsk == nil {
		print("No Ask orders exist - a default value was used\n")
		return &Level{
			levelSide: Ask,
			price:     math.MaxUint64,
			symbolId:  0,
			volume:    0,
			orders:    list.New(),
		}
	}
	return bestAsk.Value.(*Level)
}

func (orderBook *OrderBook) ValidateLimitOrders() {
	var currBestBid uint64
	var currBestAsk uint64
	if orderBook.askLevels.IsEmpty() {
		currBestAsk = math.MaxUint64
	} else {
		currBestAsk = orderBook.askLevels.GetMapBegin().Value.(*Level).price
	}
	if orderBook.bidLevels.IsEmpty() {
		currBestBid = 0
	} else {
		currBestBid = orderBook.bidLevels.GetMapEnd().Value.(*Level).price
	}
	if !(currBestAsk >= currBestBid) {
		print(currBestAsk)
		print(currBestBid)
		panic("Best bid price should never be lower than best ask price!")
	}

	orderBook.askLevels.SetMapBegin()
	itr := orderBook.askLevels.levelMapIterator
	for itr.Next() {
		level := itr.Value().(*Level)
		if level.Empty() {
			panic("There should be no empty limit levels in the orderbook")
		}
		if level.price != itr.Key().(uint64) {
			panic("The price of the limit level does not match with its key")
		}
		if level.levelSide != Ask {
			panic("Limit level with Bid side is in the ask side of the book")
		}
		for ord := level.orders.Front(); ord != nil; ord = ord.Next() {
			order := ord.Value.(*Order)
			if order.IsFilled() {
				panic("There should be no filled orders in a limit level")
			}
			if order.orderType != Limit {
				panic("There is a non Limit type order in a Limit level")
			}
		}

	}

	orderBook.bidLevels.SetMapEnd()
	itr = orderBook.bidLevels.levelMapIterator
	for itr.Prev() {
		level := itr.Value().(*Level)
		if level.Empty() {
			panic("There should be no limit empty levels in the orderbook")
		}
		if level.price != itr.Key().(uint64) {
			panic("The price of the limit level does not match with its key")
		}
		if level.levelSide != Bid {
			panic("Limit level with Ask side is in the bid side of the book")
		}
		for ord := level.orders.Front(); ord != nil; ord = ord.Next() {
			order := ord.Value.(*Order)
			if order.IsFilled() {
				panic("There should be no filled orders in a limit level")
			}
			if order.orderType != Limit {
				panic("There is a non Limit type order in a Limit level")
			}
		}
	}
}

func (orderBook *OrderBook) ValidateStopOrders() {

	orderBook.stopAskLevels.SetMapBegin()
	itr := orderBook.stopAskLevels.levelMapIterator
	for itr.Next() {
		level := itr.Value().(*Level)
		if level.Empty() {
			panic("There should be no empty stop levels in the orderbook")
		}
		if level.price > orderBook.lastExecutedPrice {
			panic("Stop ask order has a stop price that is greater than the last executed price, but was not stopped")
		}
		if level.price != itr.Key().(uint64) {
			panic("The price of the stop level does not match with its key")
		}
		if level.levelSide != Ask {
			panic("Stop level with Bid side is in the ask side of the book")
		}
		for ord := level.orders.Front(); ord != nil; ord = ord.Next() {
			order := ord.Value.(*Order)
			if order.IsFilled() {
				panic("There should be no filled orders in a stop level")
			}
			if order.orderType != Stop && order.orderType != StopLimit {
				panic("There is a non Stop type order in a Stop level")
			}
		}
	}

	orderBook.stopBidLevels.SetMapBegin()
	itr = orderBook.stopBidLevels.levelMapIterator
	for itr.Next() {
		level := itr.Value().(*Level)
		if level.Empty() {
			panic("There should be no empty stop levels in the orderbook")
		}
		if level.price < orderBook.lastExecutedPrice {
			panic("Stop bid order has a stop price that is less than the last executed price, but was not stopped")
		}
		if level.price != itr.Key().(uint64) {
			panic("The price of the stop level does not match with its key")
		}
		if level.levelSide != Bid {
			panic("Stop level with Ask side is in the bid side of the book")
		}
		for ord := level.orders.Front(); ord != nil; ord = ord.Next() {
			order := ord.Value.(*Order)
			if order.IsFilled() {
				panic("There should be no filled orders in a stop level")
			}
			if order.orderType != Stop && order.orderType != StopLimit {
				panic("There is a non Stop type order in a Stop level")
			}
		}
	}
}

func (orderBook *OrderBook) ValidateTrailingStopOrders() {

	orderBook.trailingStopAskLevels.SetMapBegin()
	itr := orderBook.trailingStopAskLevels.levelMapIterator
	for itr.Next() {
		level := itr.Value().(*Level)
		if level.Empty() {
			panic("There should be no empty trailing stop levels in the orderbook")
		}
		if level.price > orderBook.lastExecutedPrice {
			panic("trailing stop ask order has a trailing stop price that is greater than the last executed price, but was not stopped")
		}
		if level.price != itr.Key().(uint64) {
			panic("The price of the trailing stop level does not match with its key")
		}
		if level.levelSide != Ask {
			panic("trailing stop level with Bid side is in the ask side of the book")
		}
		for ord := level.orders.Front(); ord != nil; ord = ord.Next() {
			order := ord.Value.(*Order)
			if order.IsFilled() {
				panic("There should be no filled orders in a trailing stop level")
			}
			if order.orderType != TrailingStop && order.orderType != TrailingStopLimit {
				panic("There is a non trailing stop type order in a trailing stop level")
			}
		}
	}

	orderBook.trailingStopBidLevels.SetMapBegin()
	itr = orderBook.trailingStopBidLevels.levelMapIterator
	for itr.Next() {
		level := itr.Value().(*Level)
		if level.Empty() {
			panic("There should be no empty trailing stop levels in the orderbook")
		}
		if level.price < orderBook.lastExecutedPrice {
			panic("Trailing stop bid order has a stop price that is less than the last executed price, but was not stopped")
		}
		if level.price != itr.Key().(uint64) {
			panic("The price of the trailing stop level does not match with its key")
		}
		if level.levelSide != Bid {
			panic("trailing stop level with Ask side is in the bid side of the book")
		}
		for ord := level.orders.Front(); ord != nil; ord = ord.Next() {
			order := ord.Value.(*Order)
			if order.IsFilled() {
				panic("There should be no filled orders in a trailing stop level")
			}
			if order.orderType != TrailingStop && order.orderType != TrailingStopLimit {
				panic("There is a non trailing stop type order in a trailing stop level")
			}
		}
	}
}

func (orderBook *OrderBook) GetTopNBids(n int) []*Level {
	topNBids := []*Level{}
	orderBook.bidLevels.SetMapEnd()
	itr := orderBook.bidLevels.levelMapIterator
	var count int = 0
	for itr.Prev() && count < n {
		topNBids = append(topNBids, itr.Value().(*Level))
		count++
	}
	return topNBids
}

func (orderBook *OrderBook) GetTopNAsks(n int) []*Level {
	topNAsks := []*Level{}
	orderBook.askLevels.SetMapBegin()
	itr := orderBook.askLevels.levelMapIterator
	var count int = 0
	for itr.Next() && count < n {
		topNAsks = append(topNAsks, itr.Value().(*Level))
		count++
	}
	return topNAsks
}

func (orderBook *OrderBook) String() string {
	var bookString strings.Builder
	bookString.WriteString("SYMBOL ID : " + strconv.FormatUint(orderBook.symbolId, 10) + "\n")
	bookString.WriteString("LAST TRADED PRICE: " + fmt.Sprintf("%.2f", float64(orderBook.lastExecutedPrice)) + "\n")

	//Set iterator
	var itr rbt.Iterator

	// Bid orders
	bookString.WriteString("BID ORDERS\n")
	orderBook.bidLevels.SetMapBegin()
	itr = orderBook.bidLevels.levelMapIterator
	for itr.Next() {
		bookString.WriteString(itr.Value().(*Level).String())
	}

	// Ask orders
	bookString.WriteString("ASK ORDERS\n")
	orderBook.askLevels.SetMapBegin()
	itr = orderBook.askLevels.levelMapIterator
	for itr.Next() {
		bookString.WriteString(itr.Value().(*Level).String())
	}

	// Stop bid orders
	bookString.WriteString("STOP BID ORDERS\n")
	orderBook.stopBidLevels.SetMapBegin()
	itr = orderBook.stopBidLevels.levelMapIterator
	for itr.Next() {
		bookString.WriteString(itr.Value().(*Level).String())
	}

	// Stop ask orders
	bookString.WriteString("STOP ASK ORDERS\n")
	orderBook.stopAskLevels.SetMapBegin()
	itr = orderBook.stopAskLevels.levelMapIterator
	for itr.Next() {
		bookString.WriteString(itr.Value().(*Level).String())
	}

	// Trailing stop bid orders
	bookString.WriteString("TRAILING STOP BID ORDERS\n")
	orderBook.trailingStopBidLevels.SetMapBegin()
	itr = orderBook.trailingStopBidLevels.levelMapIterator
	for itr.Next() {
		bookString.WriteString(itr.Value().(*Level).String())
	}

	// Trailing stop ask orders
	bookString.WriteString("TRAILING STOP ASK ORDERS\n")
	orderBook.trailingStopAskLevels.SetMapBegin()
	itr = orderBook.trailingStopAskLevels.levelMapIterator
	for itr.Next() {
		bookString.WriteString(itr.Value().(*Level).String())
	}

	return bookString.String()
}

func (orderBook *OrderBook) OrderbookString() string {
	var bookString strings.Builder

	//Set iterator
	var itr rbt.Iterator

	// Bid orders
	bookString.WriteString("BID ORDERS\n")
	orderBook.bidLevels.SetMapBegin()
	itr = orderBook.bidLevels.levelMapIterator
	var count int = 0
	for itr.Next() && count < 15 {
		bookString.WriteString(itr.Value().(*Level).String())
		count++
	}

	// Ask orders
	count = 0
	bookString.WriteString("ASK ORDERS\n")
	orderBook.askLevels.SetMapEnd()
	itr = orderBook.askLevels.levelMapIterator
	for itr.Prev() && count < 15 {
		bookString.WriteString(itr.Value().(*Level).String())
		count++
	}

	return bookString.String()
}
