package orderAdapter

import (
	"context"
	"gRPCbigapp/app/domain"
	"testing"

	orderpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/order"
)

type MockDB struct{}

func (mdb *MockDB) Create(ctx context.Context, od *domain.OrderDomain) error {
	return nil
}

func (mdb *MockDB) GetStatus(ctx context.Context, id string) (string, error) {
	return "Создан", nil
}

func TestCreateOrder(t *testing.T) {
	testdb := &MockDB{}

	service := NewOrderService(testdb)

	req := &orderpb.CreateOrderRequest{
		UserId:    "uuid1",
		MarketId:  "Некий-Ecomerce1",
		OrderType: "limited",
		Price:     1,
		Quantity:  1,
		UserRole:  "user",
	}

	response, err := service.CreateOrder(context.Background(), req)

	if err != nil {
		t.Fatalf("неизвестная ошибка: %v", err)
	}

	if response.OrderId == "" {
		t.Fatalf("невозможно вернуть id пользователя: %v", response)
	}
}
