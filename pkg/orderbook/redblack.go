package orderbook

import (
	"container/list"
	"fmt"
	"time"

	"github.com/emirpasic/gods/v2/trees/redblacktree"
)

type OrderSide int

const (
	Bid OrderSide = iota
	Ask
)

func (s OrderSide) String() string {
	if s == Bid {
		return "Bid"
	}
	return "Ask"
}

type Order struct {
	ID        string
	Side      OrderSide
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

type OrderBook struct {
	Bids *redblacktree.Tree[float64, *list.List] // Buy orders, descending order
	Asks *redblacktree.Tree[float64, *list.List] // Sell orders, ascending order
}

func NewOrderBook() *OrderBook {
	bidComparator := func(a, b float64) int {
		if a > b {
			return -1
		}
		if a < b {
			return 1
		}
		return 0
	}

	askComparator := func(a, b float64) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
	return &OrderBook{
		Bids: redblacktree.NewWith[float64, *list.List](bidComparator),
		Asks: redblacktree.NewWith[float64, *list.List](askComparator),
	}
}

func (ob *OrderBook) Insert(order Order) {
	tree := ob.Bids
	if order.Side == Ask {
		tree = ob.Asks
	}

	queue, found := tree.Get(order.Price)
	if !found {
		queue = list.New()
		tree.Put(order.Price, queue)
	}

	queue.PushBack(order)
}

func (ob *OrderBook) Remove(side OrderSide, price float64, orderID string) bool {
	tree := ob.Bids
	if side == Ask {
		tree = ob.Asks
	}

	queue, found := tree.Get(price)
	if !found {
		return false
	}

	for e := queue.Front(); e != nil; e = e.Next() {
		if order, ok := e.Value.(Order); ok && order.ID == orderID {
			queue.Remove(e)
			if queue.Len() == 0 {
				tree.Remove(price)
			}
			return true
		}
	}
	return false
}

func (ob *OrderBook) Match() []string {
	var tradeLogs []string

	for !ob.Bids.Empty() && !ob.Asks.Empty() {
		bidIter := ob.Bids.Iterator()
		askIter := ob.Asks.Iterator()
		if !bidIter.Next() || !askIter.Next() {
			break
		}

		bidPrice := bidIter.Key()
		askPrice := askIter.Key()

		if bidPrice < askPrice {
			break
		}

		bidQueue := bidIter.Value()
		askQueue := askIter.Value()

		bidElement := bidQueue.Front()
		askElement := askQueue.Front()
		if bidElement == nil || askElement == nil {
			continue
		}

		bidOrder := bidElement.Value.(Order)
		askOrder := askElement.Value.(Order)

		tradeQty := min(bidOrder.Quantity, askOrder.Quantity)
		tradeLog := fmt.Sprintf("Trade: %s buys from %s, qty:%.2f, price:%.2f", bidOrder.ID, askOrder.ID, tradeQty, askPrice)
		tradeLogs = append(tradeLogs, tradeLog)
	}
	return tradeLogs
}

func (ob *OrderBook) GetBestBid() (float64, float64, bool) {
	if ob.Bids.Empty() {
		return 0, 0, false
	}
	iter := ob.Bids.Iterator()
	if iter.Next() {
		price := iter.Key()
		queue := iter.Value()
		if queue.Front() != nil {
			order := queue.Front().Value.(Order)
			return price, order.Quantity, true
		}
	}
	return 0, 0, false
}

func (ob *OrderBook) GetBestAsk() (float64, float64, bool) {
	if ob.Asks.Empty() {
		return 0, 0, false
	}
	iter := ob.Asks.Iterator()
	if iter.Next() {
		price := iter.Key()
		queue := iter.Value()
		if queue.Front() != nil {
			order := queue.Front().Value.(Order)
			return price, order.Quantity, true
		}
	}
	return 0, 0, false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
