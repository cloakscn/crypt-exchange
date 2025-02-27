package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloakscn/crypto-exchange/orderbook"
	"github.com/labstack/echo"
)

func main() {
	fmt.Println("Hello Crypto Exchange!")
	e := echo.New()

	ex := NewExchange()
	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)

	e.Start(":3000")
}

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

type Market string

const (
	MarketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*orderbook.Orderbook)
	orderbooks[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type PlaceOrderReq struct {
	Type   OrderType `json:"type"` // limit or market
	Bid    bool      `json:"bid"`
	Size   float64   `json:"size"`
	Price  float64   `json:"price"`
	Market Market    `json:"market"`
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderReq PlaceOrderReq

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderReq); err != nil {
		return c.JSON(http.StatusExpectationFailed, nil)
	}

	market := Market(placeOrderReq.Market)
	ob := ex.orderbooks[market]
	order := orderbook.NewOrder(placeOrderReq.Bid, placeOrderReq.Size)

	ob.PlaceLimitOrder(placeOrderReq.Price, order)

	return c.JSON(200, map[string]any{
		"msg": "order placed",
	})
}

type Order struct {
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Bid       bool    `json:"bid"`
	Timestamp int64   `json:"timestamp"`
}

type OrderbookRep struct {
	Asks []*Order `json:"asks"`
	Bids []*Order `json:"bids"`
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))

	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "market not found",
		})
	}

	var orderbookRep OrderbookRep
	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			orderbookRep.Asks = append(orderbookRep.Asks, &Order{
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			})
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			orderbookRep.Bids = append(orderbookRep.Bids, &Order{
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			})
		}
	}

	return c.JSON(http.StatusOK, orderbookRep)
}
