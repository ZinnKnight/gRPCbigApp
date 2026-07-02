package Streaming

import "sync"

type Update struct {
	OrderID string
	Status  string
}

type Hub struct {
	mu     sync.Mutex
	subs   map[string]map[int]chan Update
	nextID int
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[string]map[int]chan Update),
	}
}

func (h *Hub) Subscribe(orderID string) (int, chan Update) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.nextID
	h.nextID++

	ch := make(chan Update, 1)

	if h.subs[orderID] == nil {
		h.subs[orderID] = make(map[int]chan Update)
	}
	h.subs[orderID][id] = ch
	return id, ch
}

func (h *Hub) Unsubscribe(orderID string, id int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if m, ok := h.subs[orderID]; ok {
		delete(m, id)
		if len(m) == 0 {
			delete(h.subs, orderID)
		}
	}
}

func (h *Hub) Publish(update Update) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, ch := range h.subs[update.OrderID] {
		select {
		case ch <- update:
		default:
		}
	}
}
