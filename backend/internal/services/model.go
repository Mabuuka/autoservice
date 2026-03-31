package services

type Service struct {
	ServiceID              int64   `json:"service_id"`
	Name                   string  `json:"name"`
	Description            string  `json:"description"`
	PriceRub               float64 `json:"price_rub"`
	RegularDiscountPercent float64 `json:"regular_discount_percent"`
	DiscountedPriceRub     float64 `json:"discounted_price_rub"`
}

type CreateServiceInput struct {
	Name                   string  `json:"name"`
	Description            string  `json:"description"`
	PriceRub               float64 `json:"price_rub"`
	RegularDiscountPercent float64 `json:"regular_discount_percent"`
}
