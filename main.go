package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloakscn/crypto-exchange/orderbook"
	"github.com/labstack/echo"
)

func main() {
	fmt.Println("Hello Crypto Exchange!")
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	ex := NewExchange()
	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.cancelOrder)

	e.Start(":3000")
}

type OrderType string

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}

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

type MatchedOrder struct {
	Id    int64   `json:"id"`
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderReq PlaceOrderReq

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderReq); err != nil {
		return c.JSON(http.StatusExpectationFailed, nil)
	}

	market := Market(placeOrderReq.Market)
	ob := ex.orderbooks[market]
	order := orderbook.NewOrder(placeOrderReq.Bid, placeOrderReq.Size)

	switch placeOrderReq.Type {
	case LimitOrder:
		ob.PlaceLimitOrder(placeOrderReq.Price, order)
		return c.JSON(200, map[string]any{
			"msg": "limit order placed",
		})
	case MarketOrder:
		matches := ob.PlaceMarketOrder(order)
		matchOrders := make([]*MatchedOrder, len(matches))

		isBid := false
		if order.Bid {
			isBid = true
		}

		for i := 0; i < len(matchOrders); i++ {
			id := matches[i].Bid.Id
			if isBid {
				id = matches[i].Ask.Id
			}

			matchOrders[i] = &MatchedOrder{
				Id:    id,
				Price: matches[i].Price,
				Size:  matches[i].SizeFilled,
			}
		}

		return c.JSON(200, map[string]any{
			"matches": matchOrders,
		})
	default:
		return c.JSON(http.StatusExpectationFailed, map[string]any{
			"msg": "invalid order type",
		})
	}
}

type Order struct {
	Id        int64   `json:"id"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Bid       bool    `json:"bid"`
	Timestamp int64   `json:"timestamp"`
}

type OrderbookRep struct {
	TotalBidVolume float64  `json:"totalBidVolume"`
	TotalAskVolume float64  `json:"totalAskVolume"`
	Asks           []*Order `json:"asks"`
	Bids           []*Order `json:"bids"`
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))

	ob, ok := ex.orderbooks[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "market not found",
		})
	}

	orderbookRep := OrderbookRep{
		TotalBidVolume: ob.BitTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			orderbookRep.Asks = append(orderbookRep.Asks, &Order{
				Id:        order.Id,
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
				Id:        order.Id,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			})
		}
	}

	return c.JSON(http.StatusOK, orderbookRep)
}

func (ex *Exchange) cancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]

	order, ok := ob.Orders[int64(id)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"msg": "order not found",
		})
	}

	ob.CancelOrder(order)
	return c.JSON(http.StatusOK, map[string]any{
		"msg": "order cancelled",
	})
}
