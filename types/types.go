package types

type PriceResponse struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	DateTime      string `json:"datetime"`
	Price         string `json:"close"`
	PercentChange string `json:"percent_change"`
}
