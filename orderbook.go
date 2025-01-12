package main

import (
	"fmt"
	"sort"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

func (o *Order) String() string {
	return fmt.Sprintf("size: %.2f", o.Size)

}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Limit:     &Limit{},
		Timestamp: 0,
	}
}

type Orders []*Order

func (o Orders) Len() int {
	return len(o)
}

func (o Orders) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o Orders) Less(i, j int) bool {
	return o[i].Timestamp < o[j].Timestamp
}

type (
	Limit struct {
		Price       float64
		Orders      Orders
		TotalVolume float64
	}
	limits []*Limit
)

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

//func (l *Limit) String() string {
//	return fmt.Sprintf("[price: %.2f | volume: %.2f]", l.Price, l.TotalVolume)
//}

func (l *Limit) AddOrder(order *Order) {
	order.Limit = l
	l.Orders = append(l.Orders, order)
	l.TotalVolume += order.Size
}

// DeleteOrder
// Blog: https://www.cloaks.cn/blog/2025/01/12/
func (l *Limit) DeleteOrder(order *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == order {
			// 方法 1: 使用 append 删除元素
			// l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)

			// 方法 2: 替换删除法
			l.Orders[i] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
			break
		}
	}

	order.Limit = nil
	l.TotalVolume -= order.Size

	// resort the whole resting orders
	sort.Sort(l.Orders)
}

// ByBestAsk 按最佳买价排序
type ByBestAsk struct {
	limits
}

func (a ByBestAsk) Len() int {
	return len(a.limits)
}

func (a ByBestAsk) Swap(i, j int) {
	a.limits[i], a.limits[j] = a.limits[j], a.limits[i]
}

func (a ByBestAsk) Less(i, j int) bool {
	return a.limits[i].Price < a.limits[j].Price
}

// ByBestBid 按最佳卖价排序
type ByBestBid struct {
	limits
}

func (b ByBestBid) Len() int {
	return len(b.limits)
}

func (b ByBestBid) Swap(i, j int) {
	b.limits[i], b.limits[j] = b.limits[j], b.limits[i]
}

func (b ByBestBid) Less(i, j int) bool {
	return b.limits[i].Price > b.limits[j].Price
}

type Orderbook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func (ob *Orderbook) PlaceOrder(price float64, o *Order) []*Order {
	// 1. try to match the orders
	// matching logic
	// sort the limit by price asc
	// sort the order by timestamp desc, peace

	// 2. add the rest of the order to the orderbook
	if o.Size > 0.0 {
		ob.add(price, o)
	}

	return nil
}

func (ob *Orderbook) add(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		limit.AddOrder(o)
		if o.Bid {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	// TODO: do something
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		Asks:      []*Limit{},
		Bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}
