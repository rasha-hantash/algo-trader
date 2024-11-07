package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var cryptoSymbols = []string{
	"BTCUSDT", "ETHUSDT", "XRPUSDT", "LTCUSDT", "ADAUSDT",
	"DOTUSDT", "LINKUSDT", "BNBUSDT", "SOLUSDT", "DOGEUSDT",
	"UNIUSDT", "MATICUSDT", "AVAXUSDT", "ATOMUSDT", "ALGOUSDT",
	"XTZUSDT", "XLMUSDT", "VETUSDT", "FILUSDT", "TRXUSDT",
}

type TickerData struct {
	EventTime       int64   `json:"event_time"`
	Symbol          string  `json:"symbol"`
	Price           float64  `json:"price"`
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

// Global random number generator
var rng *rand.Rand

func init() {
	// Initialize the random number generator with a time-based seed
	source := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(source)
}


func main() {

	http.HandleFunc("/ws", handleConnections)

	log.Println("Starting server on :8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	log.Println("Client connected")

	for {
		err := ws.WriteJSON(generateData())
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
		time.Sleep(time.Millisecond) // Send data every second
	}
}

func generateData() TickerData {
	now := time.Now()
	fourSecondsAgo := now.Add(-4 * time.Second)
	symbol := cryptoSymbols[rng.Intn(len(cryptoSymbols))]
	basePrice := 40000.0
	if symbol != "BTCUSD" {
		basePrice = 100.0 // Adjust base price for non-BTC symbols
	}
	
	price := basePrice + rng.Float64()*basePrice*0.1 // 10% variation
	prevClosePrice := basePrice + rng.Float64()*basePrice*0.1
	priceChange := price - prevClosePrice
	priceChangePercent := (priceChange / prevClosePrice) * 100

	return TickerData{
		EventTime:         generateTimestamp(fourSecondsAgo, now),
		Symbol:             symbol,
		Price:              price,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		WeightedAvgPrice:   price + rng.Float64()*10 - 5,
		PrevClosePrice:     prevClosePrice,
		LastQty:            rng.Float64() * 10,
		BidPrice:           price - rng.Float64(),
		AskPrice:           price + rng.Float64(),
		OpenPrice:          prevClosePrice,
		HighPrice:          price + rng.Float64()*10,
		LowPrice:           price - rng.Float64()*10,
		Volume:             10000 + rng.Float64()*90000,
		QuoteVolume:        (10000 + rng.Float64()*90000) * price,
		OpenTime:           now.Add(-24 * time.Hour).UnixNano() / int64(time.Millisecond),
		CloseTime:          now.UnixNano() / int64(time.Millisecond),
		FirstId:            now.UnixNano() - 1000000,
		LastId:             now.UnixNano(),
		Count:              100 + rng.Int63n(900),
	}
}

func generateTimestamp(start, end time.Time) int64 {
	if rng.Float32() < 0.95 {
		return end.UnixNano() / int64(time.Millisecond)
	}
	diff := end.Sub(start)
	return start.Add(time.Duration(rng.Int63n(int64(diff)))).UnixNano() / int64(time.Millisecond)
}