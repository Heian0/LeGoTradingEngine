package exchange

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	ob "github.com/Heian0/LeGoTradingEngine/internal/orderbook"
	"github.com/google/uuid"
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

	clients sync.Map
}

type UpdateChannel struct {
	ch             chan *OrderBookState
	latestState    atomic.Value
	lastUpdateTime atomic.Value
}

func NewUpdateChannel() *UpdateChannel {
	return &UpdateChannel{
		// Smaller buffer size for better performance
		ch: make(chan *OrderBookState, 5),
	}
}

type ClientMetrics struct {
	clientId   string
	latency    time.Duration
	bufferSize int
}

type LevelStateData struct {
	price    uint64
	quantity uint64
}

type OrderBookStateData struct {
	Bids              []*LevelStateData
	Asks              []*LevelStateData
	LastExecutedPrice uint64
	BestBid           uint64
	BestAsk           uint64
	Spread            uint64
}

func (exchange *Exchange) HasOrderBook(symbolId uint64) bool {
	exchange.Mu.RLock()
	defer exchange.Mu.RUnlock()
	_, exists := exchange.orderBooks[symbolId]
	return exists
}

func (exchange *Exchange) ClearAndSendLatest(updateCh *UpdateChannel) {
	// Clear out any old messages
	for {
		select {
		case <-updateCh.ch: // Evicts
		default:
			// When empty execute this
			goto SendLatest
		}
	}

SendLatest:
	// Get the most recent state we stored in atomic.Value
	if latest := updateCh.latestState.Load(); latest != nil {
		updateCh.lastUpdateTime.Store(time.Now())
		select {
		case updateCh.ch <- latest.(*OrderBookState):
			// Successfully sent latest state
		default:
			// If channel still full (very rare), skip this update, could consider panicking/logging
		}
	}
}

func (exchange *Exchange) MonitorClient(clientId string, updateCh *UpdateChannel, metrics chan<- ClientMetrics) {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for range ticker.C {
		if lastUpdate, ok := updateCh.lastUpdateTime.Load().(time.Time); ok {
			metrics <- ClientMetrics{
				clientId:   clientId,
				latency:    time.Since(lastUpdate),
				bufferSize: len(updateCh.ch),
			}
		}
	}
}

func (exchange *Exchange) ProcessMetrics(metrics <-chan ClientMetrics) {
	for metric := range metrics {
		if metric.latency > time.Millisecond*100 {
			log.Printf("High latency for subscriber %s: %v", metric.clientId, metric.latency)
		}
		if metric.bufferSize > 10 {
			log.Printf("Buffer filling for subscriber %s: %d/20", metric.clientId, metric.bufferSize)
		}
	}
}

func (obs *OrderBookState) ObsToString() string {
	var sb strings.Builder

	// Fuck need to add book name, do that later
	sb.WriteString("Order Book:\n")
	sb.WriteString("Bids:\n")

	for i := len(obs.Bids) - 1; i >= 0; i-- {
		bid := obs.Bids[i]
		sb.WriteString(fmt.Sprintf("  Price: %d, Quantity: %d\n", bid.Price, bid.Quantity))
	}

	sb.WriteString("Asks:\n")
	for _, ask := range obs.Asks {
		sb.WriteString(fmt.Sprintf("  Price: %d, Quantity: %d\n", ask.Price, ask.Quantity))
	}

	sb.WriteString(fmt.Sprintf("Last Executed Price: %d\n", obs.LastExecutedPrice))
	sb.WriteString(fmt.Sprintf("Best Bid: %d\n", obs.BestBid))
	sb.WriteString(fmt.Sprintf("Best Ask: %d\n", obs.BestAsk))
	sb.WriteString(fmt.Sprintf("Spread: %d\n", obs.Spread))

	return sb.String()
}

func (exchange *Exchange) NotifyClients(symbolId uint64) {
	state := exchange.GetOrderBookState(symbolId)

	exchange.clients.Range(func(key, value interface{}) bool {
		updateCh := value.(*UpdateChannel)
		updateCh.latestState.Store(state)
		updateCh.lastUpdateTime.Store(time.Now())

		select {
		case updateCh.ch <- state:
		default:
			exchange.ClearAndSendLatest(updateCh)
		}
		return true
	})
}

// SubscribeToOrderBook implements ExchangeServiceServer.
func (exchange *Exchange) SubscribeToOrderBook(req *SubscribeRequest, stream ExchangeService_SubscribeToOrderBookServer) error {
	symbolId := req.GetSymbolId()

	if !exchange.HasOrderBook(symbolId) {
		fmt.Printf("This security is not supported by the exchange.")
		return nil
	}

	updateCh := NewUpdateChannel()
	clientId := uuid.New().String()
	exchange.clients.Store(clientId, updateCh)

	metrics := make(chan ClientMetrics, 100)
	go exchange.MonitorClient(clientId, updateCh, metrics)
	go exchange.ProcessMetrics(metrics)

	defer func() {
		exchange.clients.Delete(clientId)
		close(updateCh.ch)
		close(metrics)
	}()

	initialState := exchange.GetOrderBookState(symbolId)
	if err := stream.Send(initialState); err != nil {
		return err
	}

	updateCh.lastUpdateTime.Store(time.Now())

	for {
		select {
		case state := <-updateCh.ch:
			if err := stream.Send(state); err != nil {
				return err
			}
			updateCh.lastUpdateTime.Store(time.Now())
		case <-stream.Context().Done():
			return nil
		}
	}
}

// mustEmbedUnimplementedExchangeServiceServer implements ExchangeServiceServer.
func (exchange *Exchange) mustEmbedUnimplementedExchangeServiceServer() {
	panic("unimplemented")
}

func (exchange *Exchange) HandleOrder(ctx context.Context, orderMessage *OrderMessage) (*OrderResponseMessage, error) {

	fmt.Println("Recieved order from client!")
	order := createOrderFromMessage(orderMessage)

	switch orderMessage.Command {
	case Command_ADD:
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

	exchange.NotifyClients(order.GetSymbolId())
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

func (exchange *Exchange) GetOrderBookState(symbolId uint64) *OrderBookState {
	// Should be locking the book and not the exchange - change this later
	exchange.RLock()
	defer exchange.RUnlock()

	orderBook := exchange.orderBooks[symbolId]
	var obs OrderBookState

	obs.Bids = []*Level{}
	obs.Asks = []*Level{}

	topBids := orderBook.GetTopNBids(10)
	topAsks := orderBook.GetTopNAsks(10)

	for _, bid := range topBids {
		lvl := &Level{
			Price:    bid.GetPrice(),
			Quantity: bid.GetVolume(),
		}
		obs.Bids = append(obs.Bids, lvl)
	}

	for _, ask := range topAsks {
		lvl := &Level{
			Price:    ask.GetPrice(),
			Quantity: ask.GetVolume(),
		}
		obs.Asks = append(obs.Asks, lvl)
	}

	obs.LastExecutedPrice = orderBook.LastExecutedPriceAsk()
	obs.BestBid = orderBook.GetBestBid().GetPrice()
	obs.BestAsk = orderBook.GetBestAsk().GetPrice()
	obs.Spread = obs.BestAsk - obs.BestBid

	return &obs
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
