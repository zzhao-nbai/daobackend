package models

type PriceResult struct {
	Filecoin struct {
		Usd float64 `json:"usd"`
	} `json:"filecoin"`
}

