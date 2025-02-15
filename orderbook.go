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
	Size      float64 // 数量
	Bid       bool    // 是否是买单
	Limit     *Limit  // 所属限价单
	Timestamp int64   // 时间戳
}

func (o *Order) String() string {
	return fmt.Sprintf("size: %.2f", o.Size)

}

// IsFilled check if the order is filled
func (o *Order) IsFilled() bool {
	return o.Size == 0
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

func (l *Limit) Fill(o *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, order := range l.Orders {
		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.IsFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}

		if o.IsFilled() {
			break
		}
	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.Size > b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0.0
	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}

	return Match{
		Ask:        ask,
		Bid:        bid,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
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
	asks []*Limit // 卖单
	bids []*Limit // 买单

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	var matches []Match

	if o.Bid {
		if o.Size > ob.AskTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]",
				ob.AskTotalVolume(), o.Size))
		}

		for _, limit := range ob.Asks() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(false, limit)
			}
		}
	} else {
		if o.Size > ob.BitTotalVolume() {
			panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]",
				ob.BitTotalVolume(), o.Size))
		}

		for _, limit := range ob.Bids() {
			limitMatches := limit.Fill(o)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				ob.clearLimit(true, limit)
			}
		}
	}
	return matches
}

func (ob *Orderbook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		if o.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	limit.AddOrder(o)
	// TODO: do something
}

func (ob *Orderbook) clearLimit(bid bool, l *Limit) {
	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}
		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}
		}
	}
}

func (ob *Orderbook) BitTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}
	return totalVolume
}

func (ob *Orderbook) AskTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}
	return totalVolume
}

func (ob *Orderbook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *Orderbook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}
