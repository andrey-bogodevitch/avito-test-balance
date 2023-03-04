package dal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CurrencyConverter struct {
	addr   string
	apiKey string
}

func NewCurrencyConverter(addr string, apiKey string) *CurrencyConverter {
	return &CurrencyConverter{
		addr:   addr,
		apiKey: apiKey,
	}
}

func (c *CurrencyConverter) Convert(amount int, baseCurrency string, resultCurrency string) (float64, error) {
	resp, err := http.Get(
		fmt.Sprintf(
			"%s/convert?apikey=%s&from=%s&to=%s&amount=%d",
			c.addr, c.apiKey, baseCurrency, resultCurrency, amount,
		),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	type Response struct {
		Result float64 `json:"result"`
	}

	var result Response
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.Result, nil
}
