package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	// "strconv"
	"log/slog"
	"sync"

	"net/http"

	"github.com/gorilla/websocket"
)

type TickerData struct {
	EventTime       int64   `json:"event_time"`
	Symbol          string  `json:"symbol"`
	Price           float64 `json:"price"`
	PriceChange     float64 `json:"price_change"`
	PriceChangePercent float64 `json:"price_change_percent"`
	WeightedAvgPrice float64 `json:"weighted_avg_price"`
	PrevClosePrice  float64 `json:"prev_close_price"`
	LastQty         float64 `json:"last_qty"`
	BidPrice        float64 `json:"bid_price"`
	AskPrice        float64 `json:"ask_price"`
	OpenPrice       float64 `json:"open_price"`
	HighPrice       float64 `json:"high_price"`
	LowPrice        float64 `json:"low_price"`
	Volume          float64 `json:"volume"`
	QuoteVolume     float64 `json:"quote_volume"`
	OpenTime        int64   `json:"open_time"`
	CloseTime       int64   `json:"close_time"`
	FirstId         int64   `json:"first_id"`
	LastId          int64   `json:"last_id"`
	Count           int64   `json:"count"`
}

/* 

PriceChange: The absolute price change (float64).
PriceChangePercent: The percentage of price change (float64).
WeightedAvgPrice: The weighted average price (float64).
PrevClosePrice: The closing price of the previous period (float64).
LastQty: The quantity of the last trade (float64).
BidPrice: The current highest bid price (float64).
AskPrice: The current lowest ask price (float64).
OpenPrice: The opening price of the current period (float64).
HighPrice: The highest price of the current period (float64).
LowPrice: The lowest price of the current period (float64).
Volume: The trading volume in the base asset (float64).
QuoteVolume: The trading volume in the quote asset (float64).
OpenTime: The opening time of the current period (int64).
CloseTime: The closing time of the current period (int64).
FirstId: The ID of the first trade in the period (int64).
LastId: The ID of the last trade in the period (int64).
Count: The number of trades in the period (int64).

*/

// todo: run goroutine tests (like gorace) to check for data races

type LastPrice struct {
	Symbol    string `json:"symbol"`
	Price     float64 `json:"price"`
	EventTime int64  `json:"event_time"`
}

var (
	latestPrices   map[string]LastPrice
	latestPricesMu sync.RWMutex
	lastPriceChan  chan LastPrice
	done           chan struct{}
)

func main() {
	// todo: create context and pass in context.Context
	latestPrices = make(map[string]LastPrice)
	lastPriceChan = make(chan LastPrice, 100) // Buffered channel to prevent blocking
	done = make(chan struct{})

	u := url.URL{Scheme: "ws", Host: "localhost:8081", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go processMessage(c)
	go processLatestPrice()

	mux := http.NewServeMux()
	mux.HandleFunc("/latest-price", GetLatestPrice)
	log.Println("Starting HTTP server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func processMessage(c *websocket.Conn) {
	defer close(lastPriceChan)
	for {
		select {
		case <-done:
			slog.Info("Stopping message processing")
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			

			var tickerData TickerData
			err = json.Unmarshal(message, &tickerData)
			if err != nil {
				log.Println("unmarshal:", err)
				continue
			}


			lastPriceMsg := LastPrice{
				Symbol:    tickerData.Symbol,
				Price:     tickerData.Price,
				EventTime: tickerData.EventTime,
			}

			// /* make a call to insert into db */

			// todo put this in a goroutine and also data base call in a goroutine
			select {
			case lastPriceChan <- lastPriceMsg:
			case <-done:
				return
			}
		}
	}
}

func processLatestPrice() {
	for {
		select {
		case price, ok := <-lastPriceChan:
			if !ok {
				log.Println("Price channel closed")
				return
			}
			latestPricesMu.Lock()
			lastPrice, exists := latestPrices[price.Symbol]
			if !exists || price.EventTime > lastPrice.EventTime {
				// update the latest price
				latestPrices[price.Symbol] = price
			}
			latestPricesMu.Unlock()
			// slog.Info("Wrote to latestPrices")
		case <-done:
			fmt.Println("Stopping price updates")
			return
		}
	}
}

func GetLatestPrice(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Symbol parameter is required", http.StatusBadRequest)
		return
	}

	latestPricesMu.RLock()
	price, exists := latestPrices[symbol]
	latestPricesMu.RUnlock()

	if !exists {
		http.Error(w, "Price not found for the given symbol", http.StatusNotFound)
		return
	}

	slog.Info("Latest price update - Symbol: %s, Price: %s, EventTime: %d\n",)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}
