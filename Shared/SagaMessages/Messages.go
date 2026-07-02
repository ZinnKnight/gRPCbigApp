package SagaMessages

// выделил контракты отдельно что б сохранить логику быстрой замены, если понадобится

const (
	TopicOrderEvents  = "order.events"
	TopicSagaCommands = "order.saga.commands"
	TopicSagaReplies  = "order.saga.replies"
	TopicOrderStatus  = "order.status"
)

const (
	EventOrderCreated   = "OrderCreated"
	CommandReserveStock = "ReserveStock"
	EventStockReserved  = "StockReserved"
	EventStockRejected  = "StockRejected"

	EventOrderStatusChanged = "OrderStatusChanged"
)

// управляющие пути и для внутреннего пользования и для выкидки наружу

type OrderCreatedPayload struct {
	OrderID  string `json:"order_id"`
	UserID   string `json:"user_id"`
	MarketID string `json:"market_id"`
	Price    string `json:"price"`
	Amount   string `json:"amount"`
	Status   string `json:"status"`
}

type ReserveStockPayload struct {
	OrderID  string `json:"order_id"`
	UserID   string `json:"user_id"`
	MarketID string `json:"market_id"`
	Amount   string `json:"amount"`
}

type StockPayloadReplay struct {
	OrderID  string `json:"order_id"`
	MarketID string `json:"market_id"`
	Reason   string `json:"reason"`
}

type OrderStatusChangedPayload struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}
