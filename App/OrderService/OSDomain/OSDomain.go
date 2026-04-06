package OSDomain

import "time"

type OrderDomain struct {
	UserID      string
	OrderID     string
	MarketName  string
	Price       float64
	Amount      float64
	OrderStatus string
	CreatedAt   time.Time
}
