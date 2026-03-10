package tests

import (
	"context"
	"gRPCbigapp/App/domain"
	"testing"

	orderTest "gRPCbigapp/App/adapters/grpcAdapters"
	orderpb "gRPCbigapp/Protofiles/gRPCbigapp/Protofiles/order"
)

type MockDB struct{}

func (mdb *MockDB) Create(ctx context.Context, od domain.OrderDomain) error {
	return nil
}

func (mdb *MockDB) GetOrderSatus(ctx context.Context, id string) (string, error) {
	// если правильно понимаю, то тут мы тестируем именно возможность взятия по Get, поэтому прописал как заглушку "создан"
	return "Создан", nil
}

func TestCreateOrder(t *testing.T) {
	testdb := &MockDB{}

	service := orderTest.NewOrderService(testdb)

	req := &orderpb.CreateOrderRequest{
		UserId:    "uuid1",
		MarketId:  "Некий-Ecomerce1",
		OrderType: "limited",
		Price:     1,
		Quantity:  1,
		UserRole:  "user",
	}

	response, err := &service.CreateOrder(context.Background(), req)

	if err != nil {
		t.Fatalf("неизвестная ошибка: %v", err)
	}

	if response.OrderId == "" {
		t.Fatalf("невозможно вернуть id пользователя: %v", response)
	}
}
