package fetcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type BinanceResponseItem struct {
	Symbol string  `json:"symbol" required:"true"`
	Price  float64 `json:"price,string" required:"true"`
}

type Client struct {
	Latest BinanceResponse
}

func New() *Client {
	return new(Client)
}

type BinanceResponse []BinanceResponseItem

func (c *Client) Fetch() {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	for {
		start := time.Now()
		func() {
			req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/ticker/price", nil)
			if err != nil {
				fmt.Println(fmt.Errorf("unable to create request: %w", err))
				return
			}

			resRaw, err := httpClient.Do(req)
			if err != nil {
				fmt.Println(fmt.Errorf("unable to access Binance: %w", err))
				return
			}

			var res BinanceResponse
			if err := json.NewDecoder(resRaw.Body).Decode(&res); err != nil {
				fmt.Println(fmt.Errorf("unable to parse Binance response: %w", err))
				return
			}

			n := 0
			for _, item := range res {
				if strings.Contains(item.Symbol, "USDT") {
					res[n] = item
					n++
				}
			}

			c.Latest = res[:n]
		}()

		time.Sleep(1*time.Second - time.Since(start))
	}
}
