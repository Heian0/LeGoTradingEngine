package orderbook

import (
	"fmt"
	"math"
	"strconv"
	"strings"

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

func NewOrderbook(_symbolId uint64) OrderBook {
	return OrderBook{
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

func (orderBook *OrderBook) AddOrder(order Order) {
	switch order.orderType {
	case Market:
		orderBook.AddMarketOrder(&order)
		break
	case Limit:
		orderBook.AddLimitOrder(&order)
		break
	case Stop:
	case StopLimit:
	case TrailingStop:
	case TrailingStopLimit:
		orderBook.AddStopOrder(&order)
		break
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

func (orderBook *OrderBook) InsertLimitOrder(order *Order) {
	var lvlPtr *Level
	if order.IsAsk() {
		lvlPtr = orderBook.askLevels.EmplaceWithHint(order.price, Ask, order.symbolId, orderBook.askLevels.GetMapBegin())
	} else {
		lvlPtr = orderBook.bidLevels.EmplaceWithHint(order.price, Bid, order.symbolId, orderBook.askLevels.GetMapEnd())
	}
	order.levelPtr = lvlPtr
	orderBook.orders[order.id] = order
	lvlPtr.AddOrder(order)
}

func (orderBook *OrderBook) InsertStopOrder(order *Order) {
	var lvlPtr *Level
	if order.IsAsk() {
		lvlPtr = orderBook.stopAskLevels.Emplace(order.price, Ask, order.symbolId)
	} else {
		lvlPtr = orderBook.stopAskLevels.Emplace(order.price, Bid, order.symbolId)
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
	orderBook.trailingStopBidLevels, newTrailingStopBidLevels = newTrailingStopBidLevels, orderBook.trailingStopBidLevels
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
	orderBook.trailingStopBidLevels, newTrailingStopAskLevels = newTrailingStopAskLevels, orderBook.trailingStopBidLevels
	orderBook.trailingAskPrice = orderBook.lastExecutedPrice
}

func (orderBook *OrderBook) DeleteOrder(orderId uint64) {

}

func (orderBook *OrderBook) DeleteOrderNoti(orderId uint64, noti bool) {

}

func (orderBook *OrderBook) Match(order *Order) {
	if order.IsFillOrKill() && !orderBook.CanMatch(order) {
		fmt.Println("Order is a Fill or Kill that could not be matched")
		return
	}
	if order.IsAsk() {
		orderBook.bidLevels.SetMapEnd()
		bidLevelsIt := orderBook.bidLevels.levelMapIterator
		askOrder := order
		for bidLevelsIt.Prev() && bidLevelsIt.Key().(uint64) >= askOrder.price && !askOrder.IsFilled() {
			bidLevel := bidLevelsIt.Value().(*Order)
			bidOrder := bidLevel.levelPtr.Front()
			executingPrice := bidOrder.price
			orderBook.ExecuteOrders(askOrder, bidOrder, executingPrice)
			bidLevel.levelPtr.ReduceVolume(bidOrder.lastExecutedQuantity)
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
		for askLevelsIt.Next() && askLevelsIt.Key().(uint64) <= bidOrder.price && !bidOrder.Filled() {
			askLevel := askLevelsIt.Value().(*Order)
			askOrder := askLevel.levelPtr.Front()
			executingPrice := askOrder.price
			orderBook.ExecuteOrders(askOrder, bidOrder, executingPrice)
			askLevel.levelPtr.ReduceVolume(askOrder.lastExecutedQuantity)
			if askOrder.IsFilled() {
				orderBook.DeleteOrder(askOrder.id, true)
			}
		}
		orderBook.askLevels.SetMapBegin()
	}
}

func (orderBook *OrderBook) ExecuteOrders(askOrder *Order, bidOrder *Order, executingPrice uint64) {
	matchedQuantity := simplemath.Min(askOrder.openQuantity, bidOrder.openQuantity)
	askOrder.ExecuteOrder(matchedQuantity, executingPrice)
	bidOrder.ExecuteOrder(matchedQuantity, executingPrice)
	// Handle Ask order execution
	// Handle Bid order execution
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

func (orderBook *OrderBook) String() string {
	var bookString strings.Builder
	bookString.WriteString("SYMBOL ID : " + strconv.FormatUint(orderBook.symbolId, 10) + "\n")
	bookString.WriteString("LAST TRADED PRICE: " + fmt.Sprintf("%f", book.lastExecutedPrice) + "\n")

	// Bid orders
	bookString.WriteString("BID ORDERS\n")
	for price, level := range orderBook.bidLevels {
		bookString.WriteString(level.String() + "\n")
	}

	// Ask orders
	bookString.WriteString("ASK ORDERS\n")
	for price, level := range book.adkLevels {
		bookString.WriteString(level.String() + "\n")
	}

	// Ask stop order
	bookString.WriteString("ASK STOP ORDERS\n")
	for _, level := range book.stopAskLevels {
		bookString.WriteString(level.String())
	}

	// Bid trailing stop orders
	bookString.WriteString("BID TRAILING STOP ORDERS\n")
	for price, level := range book.trailingStopBidLevels {
		bookString.WriteString(level.String())
	}

	// ASK TRAILING STOP ORDERS
	sb.WriteString("ASK TRAILING STOP ORDERS\n")
	for _, level := range book.trailingStopAskLevels {
		sb.WriteString(level.toString())
	}

	return bookString.String()
}
