import {
	"fmt"
}

type OrderType int
const (
	Limit = 0,
	Market = 1,
	Stop = 2,
	StopLimit = 3,
    TrailingStop = 4,
    TrailingStopLimit = 5
)

func (orderType OrderType) String() string {
	switch orderType {
		case Limit: return "Limit"
		case Market: return "Market"
		case Stop: return "Stop"
		case StopLimit: return "Stop Limit"
		case TrailingStop: return "Trailing Stop"
		case TrailingStopLimit: return "Trailing Stop Limit"
		default: return "Unknown Order Type"
	}
}

type OrderTimeInForce int
const (
	GoodTillCancel = 0
	ImmediateOrCancel = 1
	FillOrKill = 2
)

func (orderTimeInForce OrderTimeInForce) String() string {
	switch orderTimeInForce {
		case GoodTillCancel: return "Good Till Cancel"
		case ImmediateOrCancel: return "Immediate Or Cancel"
		case FillOrKill: return "Fill Or Kill"
		default: return "Unknown Order In Force"
	}
}

type OrderSide int
const (
	Bid = 0
	Ask = 1
)

func (orderSide OrderSide) String() string {
	switch orderSide {
		case Bid: return "Bid"
		case Ask: return "Ask"
		default: return "Unknown Order Side"
	}
}

type Order struct {

	OrderType orderType
    OrderSide orderSide
    OrderTimeInForce orderTimeInForce
	id uint64
	symbolId uint64
	price uint64
	stopPrice = uint64
	trailingAmount = uint64
	lastExecutedPrice = uint64
	quantity = uint64
	executedQuantity = uint64
	openQuantity = uint64
	lastExecutedQuantity = uint64
}

// OrderToString returns a formatted string with Order details
func (order Order) String() string {
	return fmt.Sprintf("Order ID: %d\nType: %v\nSide: %v\nTime in Force: %v\nSymbol ID: %d\nPrice: %d\nStop Price: %d\nTrailing Amount: %d\nQuantity: %d\nExecuted Quantity: %d\nOpen Quantity: %d\nLast Executed Price: %d\nLast Executed Quantity: %d",
		order.ID,
		order.OrderType,
		order.OrderSide,
		order.OrderTimeInForce,
		order.SymbolID,
		order.Price,
		order.StopPrice,
		order.TrailingAmount,
		order.Quantity,
		order.ExecutedQuantity,
		order.OpenQuantity,
		order.LastExecutedPrice,
		order.LastExecutedQuantity,
	)
}

// ---------------- Order Input Verification

func ValidateOrder(order *Order) bool {
	switch order.orderType {
	case Market:
		if _id < 0 || _symbolId < 0 || _quantity < 0 || _orderTimeInForce == GoodTillCancel { 
			fmt.Printf("Invalid Market order, please review the following order (id: %v, symbolId: %v, quantity: %v, orderTimeInForce: %v)", _id, _symbolId, _quantity, _orderTimeInForce)
			return false 
		}
	case Limit:
		if _id < 0 || _symbolId < 0 || _quantity < 0 || _price < 0 { 
			fmt.Printf("Invalid Limit order, please review the following order (id: %v, symbolId: %v, quantity: %v, price: %v)", _id, _symbolId, _quantity, _price)
			return false 
		}
	case Stop:
		if _id < 0 || _symbolId < 0 || _quantity < 0 ||	_stop_price < 0 || _orderTimeInForce == GoodTillCancel { 
			fmt.Printf("Invalid Stop order, please review the following order (id: %v, symbolId: %v, quantity: %v, stopPrice: %v, orderTimeInForce: %v)", _id, _symbolId, _quantity, _stopPrice, _orderTimeInForce)
			return false 
		}
	case StopLimit:
		if _id < 0 || _symbolId < 0 || _quantity < 0 ||	_price < 0 || _stop_price < 0 { 
			fmt.Printf("Invalid Stop Limit order, please review the following order (id: %v, symbolId: %v, quantity: %v, price: %v, stopPrice: %v, orderTimeInForce: %v)", _id, _symbolId, _quantity, _price, _stopPrice, _orderTimeInForce)
			return false 
		}
	case TrailingStop:
		if _id < 0 || _symbolId < 0 || _quantity < 0 || _trailingAmount < 0 || _orderTimeInForce == GoodTillCancel { 
			fmt.Printf("Invalid Trailing Stop order, please review the following order (id: %v, symbolId: %v, quantity: %v, trailingAmount: %v, orderTimeInForce: %v)", _id, _symbolId, _quantity, _trailingAmount, _orderTimeInForce)
			return false 
		}
	case TrailingStopLimit:
		if _id < 0 || _symbolId < 0 || _quantity < 0 || _price < 0 || _trailingAmount < 0 { 
			fmt.Printf("Invalid Trailing Stop Limit order, please review the following order (id: %v, symbolId: %v, quantity: %v, price: %v, trailingAmount: %v)", _id, _symbolId, _quantity, _price, _trailingAmount)
			return false 
		}
	}

	// Valid Order
	return true
}

// -----------------------------------------------

func MarketBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _orderTimeInForce OrderTimeInForce) *Order {
	order := &Order { orderType: Market, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity }
	if !ValidateOrder(order) { 
	 	panic("Error, invalid order")
		return nil
	}
	return order
}

func MarketAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _orderTimeInForce OrderTimeInForce) *Order {
	order :=  &Order { orderType: Market, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func LimitBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _orderTimeInForce OrderTimeInForce) *Order {
	order := Order { orderType: Limit, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, openQuantity: _quantity, price: _price }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func LimitAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _orderTimeInForce OrderTimeInForce) *Order {
	order := &Order { orderType: Limit, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, price: _price }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func StopBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: Stop, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, stopPrice: _stopPrice }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func StopAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: Stop, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, stopPrice: _stopPrice }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func StopLimitBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: StopLimit, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, stopPrice: _stopPrice }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func StopLimitAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: StopLimit, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, stopPrice: _stopPrice }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func TrailingStopBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: TrailingStop, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, trailingAmount: _trailingAmount }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func TrailingStopAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: TrailingStop, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, trailingAmount: _trailingAmount }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func TrailingStopLimitBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: TrailingStop, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, trailingAmount: _trailingAmount }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}

func TrailingStopLimitAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) {
	order := &Order { orderType: TrailingStop, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, trailingAmount: _trailingAmount }
	if !ValidateOrder(order) { 
		panic("Error, invalid order")
	   return nil
	}
	return order
}