package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	concurrentRequests := 100
	totalRequests := 1000

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			symbol := symbols[i%len(symbols)]
			url := fmt.Sprintf("http://localhost:8080/latest-price?symbol=%s", symbol)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Request %d failed: %v\n", i, err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("Request %d completed with status: %s\n", i, resp.Status)
		}(i)

		if (i+1)%concurrentRequests == 0 {
			wg.Wait()
		}
	}

	wg.Wait()
	duration := time.Since(startTime)
	fmt.Printf("Total time: %v\n", duration)
	fmt.Printf("Requests per second: %.2f\n", float64(totalRequests)/duration.Seconds())
}