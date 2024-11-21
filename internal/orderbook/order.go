package orderbook

import (
	"fmt"
)

type OrderType int

const (
	Limit             = 0
	Market            = 1
	Stop              = 2
	StopLimit         = 3
	TrailingStop      = 4
	TrailingStopLimit = 5
)

func (orderType OrderType) String() string {
	switch orderType {
	case Limit:
		return "Limit"
	case Market:
		return "Market"
	case Stop:
		return "Stop"
	case StopLimit:
		return "Stop Limit"
	case TrailingStop:
		return "Trailing Stop"
	case TrailingStopLimit:
		return "Trailing Stop Limit"
	default:
		return "Unknown Order Type"
	}
}

type OrderTimeInForce int

const (
	GoodTillCancel    = 0
	ImmediateOrCancel = 1
	FillOrKill        = 2
)

func (orderTimeInForce OrderTimeInForce) String() string {
	switch orderTimeInForce {
	case GoodTillCancel:
		return "Good Till Cancel"
	case ImmediateOrCancel:
		return "Immediate Or Cancel"
	case FillOrKill:
		return "Fill Or Kill"
	default:
		return "Unknown Order In Force"
	}
}

type Side int

const (
	Bid = 0
	Ask = 1
)

func (orderSide Side) String() string {
	switch orderSide {
	case Bid:
		return "Bid"
	case Ask:
		return "Ask"
	default:
		return "Unknown Order Side"
	}
}

type Order struct {
	orderType            OrderType
	orderSide            Side
	orderTimeInForce     OrderTimeInForce
	id                   uint64
	symbolId             uint64
	price                uint64
	stopPrice            uint64
	trailingAmount       uint64
	lastExecutedPrice    uint64
	quantity             uint64
	executedQuantity     uint64
	openQuantity         uint64
	lastExecutedQuantity uint64
	levelPtr             *Level
}

// OrderToString returns a formatted string with Order details
func (order Order) String() string {
	return fmt.Sprintf("Order ID: %d\nType: %v\nSide: %v\nTime in Force: %v\nSymbol ID: %d\nPrice: %d\nStop Price: %d\nTrailing Amount: %d\nQuantity: %d\nExecuted Quantity: %d\nOpen Quantity: %d\nLast Executed Price: %d\nLast Executed Quantity: %d",
		order.id,
		order.orderType,
		order.orderSide,
		order.orderTimeInForce,
		order.symbolId,
		order.price,
		order.stopPrice,
		order.trailingAmount,
		order.quantity,
		order.executedQuantity,
		order.openQuantity,
		order.lastExecutedPrice,
		order.lastExecutedQuantity,
	)
}

// ---------------- Order Input Verification

func (order Order) ValidateOrder() bool {
	switch order.orderType {
	case Market:
		if order.orderTimeInForce == GoodTillCancel {
			fmt.Printf("Invalid Market order, please review the following order (id: %v, symbolId: %v, quantity: %v, orderTimeInForce: %v)\n", order.id, order.symbolId, order.quantity, order.orderTimeInForce)
			return false
		}
	case Stop:
		if order.orderTimeInForce == FillOrKill {
			fmt.Printf("Invalid Stop order, please review the following order (id: %v, symbolId: %v, quantity: %v, stopPrice: %v, orderTimeInForce: %v)\n", order.id, order.symbolId, order.quantity, order.stopPrice, order.orderTimeInForce)
			return false
		}
	case TrailingStop:
		if order.orderTimeInForce == FillOrKill {
			fmt.Printf("Invalid Trailing Stop order, please review the following order (id: %v, symbolId: %v, quantity: %v, trailingAmount: %v, orderTimeInForce: %v)\n", order.id, order.symbolId, order.quantity, order.trailingAmount, order.orderTimeInForce)
			return false
		}
	}

	// Valid Order
	return true
}

// -----------------------------------------------

func MarketBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: Market, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func MarketAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: Market, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func LimitBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: Limit, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func LimitAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: Limit, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func StopBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: Stop, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, stopPrice: _stopPrice}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func StopAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: Stop, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, stopPrice: _stopPrice}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func StopLimitBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: StopLimit, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, stopPrice: _stopPrice}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func StopLimitAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _stopPrice uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: StopLimit, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, stopPrice: _stopPrice}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func TrailingStopBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: TrailingStop, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, trailingAmount: _trailingAmount}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func TrailingStopAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: TrailingStop, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, trailingAmount: _trailingAmount}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func TrailingStopLimitBidOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: TrailingStop, orderSide: Bid, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, trailingAmount: _trailingAmount}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func TrailingStopLimitAskOrder(_id uint64, _symbolId uint64, _quantity uint64, _price uint64, _trailingAmount uint64, _orderTimeInForce OrderTimeInForce) Order {
	order := Order{orderType: TrailingStop, orderSide: Ask, orderTimeInForce: _orderTimeInForce, id: _id, symbolId: _symbolId, quantity: _quantity, openQuantity: _quantity, price: _price, trailingAmount: _trailingAmount}
	if !order.ValidateOrder() {
		panic("Error, invalid order")
	}
	return order
}

func (order *Order) ExecuteOrder(_quantity uint64, _price uint64) {
	if order.openQuantity < _quantity {
		panic("Error, invalid order execution")
	}
	order.openQuantity -= _quantity
	order.executedQuantity += _quantity
	order.lastExecutedPrice = _price
	order.lastExecutedQuantity = _quantity
}

func (order *Order) IsAsk() bool { return order.orderSide == Ask }

func (order *Order) IsBid() bool { return order.orderSide == Bid }

func (order *Order) IsMarket() bool { return order.orderType == Market }

func (order *Order) IsLimit() bool { return order.orderType == Limit }

func (order *Order) IsStop() bool { return order.orderType == Stop }

func (order *Order) IsStopLimit() bool { return order.orderType == StopLimit }

func (order *Order) IsTrailingStop() bool { return order.orderType == TrailingStop }

func (order *Order) IsTrailingStopLimit() bool { return order.orderType == TrailingStopLimit }

func (order *Order) IsGoodTillCancel() bool { return order.orderTimeInForce == GoodTillCancel }

func (order *Order) IsImmediateOrCancel() bool { return order.orderTimeInForce == ImmediateOrCancel }

func (order *Order) IsFillOrKill() bool { return order.orderTimeInForce == FillOrKill }

func (order *Order) IsFilled() bool { return order.openQuantity == 0 }

func (order *Order) Equals(otherOrder *Order) bool { return order.id == otherOrder.id }
