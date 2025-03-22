package orderbook

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// of course i'm opinionated although i write
// a lot of rust myself

func assert(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)
	fmt.Println(l)
	l.DeleteOrder(buyOrderA)
	fmt.Println(l)
	l.DeleteOrder(buyOrderB)
	fmt.Println(l)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(false, 10)
	sellOrderB := NewOrder(false, 5)
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(11_000, sellOrderB)

	assert(t, len(ob.orders), 2)
	assert(t, ob.orders[sellOrderA.Id], sellOrderA)
	assert(t, len(ob.asks), 2)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)

	fmt.Printf("%+v", matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 10)
	buyOrderB := NewOrder(true, 20)
	buyOrderC := NewOrder(true, 30)
	buyOrderD := NewOrder(true, 1)

	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	assert(t, ob.BitTotalVolume(), 61.0)

	// what the hell is this?
	sellOrderA := NewOrder(false, 50)
	matches := ob.PlaceMarketOrder(sellOrderA)

	assert(t, ob.BitTotalVolume(), 11.0)
	assert(t, len(matches), 3)
	assert(t, len(ob.bids), 1)

	fmt.Printf("%+v", matches)
}

func TestOrderMethods(t *testing.T) {
	order := NewOrder(true, 10)
	assert(t, order.IsFilled(), false)
	assert(t, order.String(), "size: 10.00")

	order.Size = 0
	assert(t, order.IsFilled(), true)
}

func TestLimitFill(t *testing.T) {
	l := NewLimit(10_000)
	buyOrder := NewOrder(true, 10)
	sellOrder := NewOrder(false, 5)

	l.AddOrder(buyOrder)
	matches := l.Fill(sellOrder)

	assert(t, len(matches), 1)
	assert(t, matches[0].SizeFilled, 5.0)
	assert(t, buyOrder.Size, 5.0)
	assert(t, sellOrder.IsFilled(), true)
}

func TestOrderbookTotals(t *testing.T) {
	ob := NewOrderbook()

	// Test empty orderbook
	assert(t, ob.BitTotalVolume(), 0.0)
	assert(t, ob.AskTotalVolume(), 0.0)

	// Add orders
	buyOrder := NewOrder(true, 10)
	sellOrder := NewOrder(false, 5)
	ob.PlaceLimitOrder(10_000, buyOrder)
	ob.PlaceLimitOrder(11_000, sellOrder)

	assert(t, ob.BitTotalVolume(), 10.0)
	assert(t, ob.AskTotalVolume(), 5.0)
}

func TestOrderSorting(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(true, 10)
	buyOrderB := NewOrder(true, 20)
	sellOrderA := NewOrder(false, 5)
	sellOrderB := NewOrder(false, 15)

	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(11_000, sellOrderA)
	ob.PlaceLimitOrder(12_000, sellOrderB)

	// Test bid sorting (highest first)
	bids := ob.Bids()
	assert(t, bids[0].Price, 10_000.0)
	assert(t, bids[1].Price, 9_000.0)

	// Test ask sorting (lowest first)
	asks := ob.Asks()
	assert(t, asks[0].Price, 11_000.0)
	assert(t, asks[1].Price, 12_000.0)
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(true, 10)
	ob.PlaceLimitOrder(10_000, buyOrder)

	assert(t, ob.BitTotalVolume(), 10.0)
	fmt.Println("cancel befor: ob.bids=", ob.bids[0])

	ob.CancelOrder(buyOrder)

	assert(t, ob.BitTotalVolume(), 0.0)
	fmt.Println("cancel after: ob.bids=", ob.bids[0])

	_, ok := ob.orders[buyOrder.Id]
	assert(t, ok, false)
}

// func TestMarketOrderErrors(t *testing.T) {
// 	ob := NewOrderbook()
// 	assert(t, len(ob.bids), 1)
// }

func TestProgress(t *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Printf("\r[%-100s]\t percent %d\t%d|%d",
			strings.Repeat("#", i), i, i, 100)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println()
}
