package exchange

import (
	"context"
	"fmt"
	"sync"

	ob "github.com/Heian0/LeGoTradingEngine/internal/orderbook"
)

// Strict Validation Exchange - will not accept an order for a non supported security
type Exchange struct {
	orderBooks map[uint64]*ob.OrderBook
	// Sort of unneeded as symbolId is stored in the Orderbook struct
	// However might be useful when we want to just grab symbol names
	symbolMap map[uint64]*ob.Symbol
	ui        *ExchangeUI

	Name     string
	Mu       sync.RWMutex
	updateCh chan struct{}
}

func (exchange *Exchange) HandleOrder(ctx context.Context, orderMessage *OrderMessage) (*OrderResponseMessage, error) {

	fmt.Println("Recieved order from client!")

	switch orderMessage.Command {
	case Command_ADD:
		order := createOrderFromMessage(orderMessage)
		exchange.AddOrder(order)
	case Command_DELETE:
		fmt.Println("Task status is PENDING")
	case Command_CANCEL:
		fmt.Println("Task status is ACTIVE")
	case Command_REPLACE:
		fmt.Println("Task status is COMPLETED")
	default:
		panic("Unknown command sent")
	}

	return &OrderResponseMessage{ExchangeStatus: exchange.String()}, nil
}

func NewExchange() *Exchange {
	exchange := Exchange{
		orderBooks: make(map[uint64]*ob.OrderBook),
		symbolMap:  make(map[uint64]*ob.Symbol),
		Name:       "New Exchange",
		updateCh:   make(chan struct{}, 1),
	}
	return &exchange
}

func createOrderFromMessage(orderMessage *OrderMessage) *ob.Order {
	switch orderMessage.OrderType {
	case OrderType_LIMIT:
		if orderMessage.OrderSide == Side_ASK {
			order := ob.LimitAskOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.Price, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		} else if orderMessage.OrderSide == Side_BID {
			order := ob.LimitBidOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.Price, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		}

	case OrderType_MARKET:
		if orderMessage.OrderSide == Side_ASK {
			order := ob.MarketAskOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		} else if orderMessage.OrderSide == Side_BID {
			order := ob.MarketBidOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		}

	case OrderType_STOP:
		if orderMessage.OrderSide == Side_ASK {
			order := ob.StopAskOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.StopPrice, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		} else if orderMessage.OrderSide == Side_BID {
			order := ob.StopBidOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.StopPrice, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		}

	case OrderType_STOP_LIMIT:
		if orderMessage.OrderSide == Side_ASK {
			order := ob.StopLimitAskOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.Price, orderMessage.StopPrice, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		} else if orderMessage.OrderSide == Side_BID {
			order := ob.StopLimitBidOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.Price, orderMessage.StopPrice, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		}

	case OrderType_TRAILING_STOP:
		if orderMessage.OrderSide == Side_ASK {
			order := ob.TrailingStopAskOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.TrailingAmount, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		} else if orderMessage.OrderSide == Side_BID {
			order := ob.TrailingStopBidOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.TrailingAmount, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		}

	case OrderType_TRAILING_STOP_LIMIT:
		if orderMessage.OrderSide == Side_ASK {
			order := ob.TrailingStopLimitAskOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.Price, orderMessage.TrailingAmount, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		} else if orderMessage.OrderSide == Side_BID {
			order := ob.TrailingStopLimitBidOrder(orderMessage.Id, orderMessage.SymbolId, orderMessage.Quantity, orderMessage.Price, orderMessage.TrailingAmount, protoToObEnumOTIF(orderMessage.OrderTimeInForce))
			return &order
		}

	default:
		panic("Unknown Ordertype in add order command")
	}
	return nil
}

// RLock locks the exchange for reading
func (e *Exchange) RLock() {
	e.Mu.RLock()
}

// RUnlock unlocks the exchange
func (e *Exchange) RUnlock() {
	e.Mu.RUnlock()
}

// Private function because this should only ever be called by AddOrderbook
func (exchange *Exchange) addSymbol(symbolId uint64, ticker string) {
	// Symbol should never exist as it gets caught be AddOrderbook
	exchange.checkOrderbookDoesNotExist(symbolId)
	newSymbol := ob.NewSymbol(symbolId, ticker)
	exchange.symbolMap[symbolId] = &newSymbol
}

func (exchange *Exchange) AddOrderbook(symbolId uint64, ticker string) {
	exchange.checkOrderbookDoesNotExist(symbolId)
	exchange.addSymbol(symbolId, ticker)
	exchange.orderBooks[symbolId] = ob.NewOrderbook(symbolId)
	// Handle a new orderbook/symbol added
}

func (exchange *Exchange) DeleteOrderbook(symbolId uint64) {
	// Tbh only one of these checks is needed...or tbh if you try to delete a non existent
	// book nothing should happen, just for error checking atm.
	exchange.checkOrderbookExists(symbolId)
	delete(exchange.symbolMap, symbolId)
	delete(exchange.orderBooks, symbolId)
	//Handle symbol deletion
}

func (exchange *Exchange) AddOrder(order *ob.Order) {
	symbolId := order.GetSymbolId()
	exchange.checkOrderbookExists(symbolId)
	exchange.orderBooks[symbolId].AddOrder(order)
	// Handle order added
}

func (exchange *Exchange) DeleteOrder(order *ob.Order) {
	symbolId := order.GetSymbolId()
	exchange.checkOrderbookExists(symbolId)
	exchange.orderBooks[symbolId].DelOrder(order.GetId())
	// Handle order deleted
}

func (exchange *Exchange) CancelOrder(order *ob.Order, cancellingQuantity uint64) {
	symbolId := order.GetSymbolId()
	exchange.checkOrderbookExists(symbolId)
	if cancellingQuantity <= 0 {
		panic("Cancelling quantity must be positive")
	}
	exchange.orderBooks[symbolId].CancelOrder(order.GetId(), cancellingQuantity)
	// Handle order cancelled
}

func (exchange *Exchange) ReplaceOrder(order *ob.Order, newOrderId uint64, newPrice uint64) {
	symbolId := order.GetSymbolId()
	exchange.checkOrderbookExists(symbolId)
	exchange.orderBooks[symbolId].ReplaceOrder(order.GetId(), newOrderId, newPrice)
	// Handle order replaced
}

func (exchange *Exchange) ExecuteOrderWithSpecifiedPrice(symbolId uint64, orderId uint64, quantity uint64, price uint64) {
	exchange.checkOrderbookExists(symbolId)
	orderBook := exchange.orderBooks[symbolId]
	if quantity <= 0 {
		panic("Quantity must be positive")
	}
	if price <= 0 {
		panic("Price must be positive")
	}
	orderBook.ExecuteOrderWithSpecifiedPrice(orderId, quantity, price)
}

func (exchange *Exchange) ExecuteOrderWithoutPrice(symbolId uint64, orderId uint64, quantity uint64) {
	exchange.checkOrderbookExists(symbolId)
	orderBook := exchange.orderBooks[symbolId]
	if quantity <= 0 {
		panic("Quantity must be positive")
	}
	orderBook.ExecuteOrderWithoutSpecifiedPrice(orderId, quantity)
}

func (exchange *Exchange) checkOrderbookExists(symbolId uint64) bool {
	_, symbolExists := exchange.symbolMap[symbolId]
	if !symbolExists {
		panic("Symbol doesn't exist")
		//return false
	}
	_, bookExists := exchange.orderBooks[symbolId]
	if !bookExists {
		panic("Orderbook doesn't exist")
		//return false
	}
	return true
}

func (exchange *Exchange) checkOrderbookDoesNotExist(symbolId uint64) bool {
	_, symbolExists := exchange.symbolMap[symbolId]
	if symbolExists {
		panic("Symbol already exists")
		//return false
	}
	_, bookExists := exchange.orderBooks[symbolId]
	if bookExists {
		panic("Orderbook already exists")
		//return false
	}
	return true
}

func (exchange *Exchange) String() string {
	var output string
	for key, value := range exchange.orderBooks {
		output += fmt.Sprintf("Ticker: %s", exchange.symbolMap[key])
		output += value.OrderbookString()
		output += "\n--------------\n"
	}
	return output
}

func (exchange *Exchange) Test() {
	exchange.Name = "Changed name!"
	fmt.Println("Changed exchange name")
	//exchange.ui.Redraw()
}

func protoToObEnumOTIF(protoOTIF OrderTimeInForce) ob.OrderTimeInForce {
	if protoOTIF == OrderTimeInForce_FOK {
		return ob.FillOrKill
	}
	if protoOTIF == OrderTimeInForce_GTC {
		return ob.GoodTillCancel
	}
	if protoOTIF == OrderTimeInForce_IOC {
		return ob.ImmediateOrCancel
	}
	panic("Invalid Proto OTIF")
}
