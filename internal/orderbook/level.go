package orderbook

import (
	"container/list"
	"fmt"
	"strconv"
)

type Level struct {
	levelSide Side
	price     uint64
	symbolId  uint64
	volume    uint64
	orders    *list.List
}

func (level Level) ValidateLevel() bool {
	var actualVolume uint64 = 0
	for orderElem := level.orders.Front(); orderElem != nil; orderElem = orderElem.Next() {
		order := orderElem.Value.(*Order)
		if order.IsStop() || order.IsStopLimit() || order.IsTrailingStop() || order.IsTrailingStopLimit() {
			if level.price != order.stopPrice {
				fmt.Println("Order stop price doesn't match level price")
				return false
			}
		} else {
			if level.price != order.price {
				fmt.Println("Order price doesn't match level price")
				return false
			}
		}
		if order.IsMarket() {
			fmt.Println("Level should never contain market orders")
			return false
		}
		actualVolume += order.openQuantity
	}
	if actualVolume != level.volume {
		fmt.Println("Incorrect level volume")
		return false
	}
	return true
}

func NewLevel(levelSide Side, price uint64, symbolId uint64) Level {
	level := Level{
		levelSide: levelSide,
		price:     price,
		symbolId:  symbolId,
		volume:    0,
		orders:    list.New(),
	}
	if !level.ValidateLevel() {
		panic("Error, invalid level")
	}
	return level
}

func (level *Level) String() string {
	levelString := strconv.FormatUint(level.volume, 10) + " @ " + strconv.FormatUint(level.price, 10) + "\n"
	return levelString
}

/*
func (level Level) GetOrdersConst() *list.List {
	return level.orders
}
*/

func (level *Level) GetPrice() uint64 {
	return level.price
}

func (level *Level) GetVolume() uint64 {
	return level.volume
}

func (level *Level) AddOrder(order *Order) {
	if level.levelSide != order.orderSide {
		print("Order is not on the same side as level")
		return
	}
	if level.symbolId != order.symbolId {
		print("Order and level have different symbols")
		return
	}
	level.orders.PushBack(order)
	level.volume += order.openQuantity
	if !level.ValidateLevel() {
		panic("Invalid add order operation")
	}
}

func (level *Level) PopFront() {
	if level.orders.Len() == 0 {
		panic("Cannot pop from empty level!")
	}
	orderElem := level.orders.Front()
	orderToRemove := orderElem.Value.(*Order)
	level.volume -= orderToRemove.openQuantity
	level.orders.Remove(orderElem)
	if !level.ValidateLevel() {
		panic("Invalid add order operation")
	}
}

func (level *Level) PopBack() {
	if level.orders.Len() == 0 {
		panic("Cannot pop back from empty level!")
	}
	orderElem := level.orders.Back()
	orderToRemove := orderElem.Value.(*Order)
	level.volume -= orderToRemove.openQuantity
	level.orders.Remove(orderElem)
	if !level.ValidateLevel() {
		panic("Invalid add order operation")
	}
}

func (level *Level) DeleteOrder(orderToRemove *Order) {
	if level.orders.Len() == 0 {
		panic("No orders to delete!")
	}

	for ord := level.orders.Front(); ord != nil; ord = ord.Next() {
		order := ord.Value.(*Order)
		if order.Equals(orderToRemove) {
			level.volume -= order.openQuantity
			level.orders.Remove(ord)

			if !level.ValidateLevel() {
				panic("Invalid level state after deleting order")
			}
			return
		}
	}

	panic("Order not found in the list!")
}

func (level *Level) ReduceVolume(amountToReduce uint64) {
	if level.volume < amountToReduce {
		panic("Can't reduce volume by an amount greater than current volume")
	}
	level.volume -= amountToReduce
	if !level.ValidateLevel() {
		panic("Invalid level state after reducing volume")
	}
}

func (level *Level) Front() *Order {
	if level.orders.Len() == 0 {
		panic("Level is empty")
	}
	return level.orders.Front().Value.(*Order)
}

func (level *Level) Back() *Order {
	if level.orders.Len() == 0 {
		panic("Level is empty")
	}
	return level.orders.Back().Value.(*Order)
}

func (level *Level) Empty() bool {
	return level.orders.Len() == 0
}
