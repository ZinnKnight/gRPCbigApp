package domain

type OrderDomain struct {
	UserID      string
	OrderID     string
	MarketName  string
	Price       float64
	Amount      float64
	OrderStatus string
}
