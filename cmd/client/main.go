package main

import (
	"bufio"
	"context"
	"fmt"
	"gRPCbigapp/app/interceptors/id_interceptor"
	orderpb "gRPCbigapp/protofiles/gRPCbigapp/Protofiles/order"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	adders := os.Getenv("ORDER_SERVICE_ADDRESSES")
	if adders == "" {
		adders = "localhost:50051"
	}

	connection, err := grpc.NewClient(adders,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(id_interceptor.XRequestIdInterceptor),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка подключения: %v\n", err)
		os.Exit(1)
	}
	defer connection.Close()

	client := orderpb.NewOrderServiceClient(connection)
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Команды: create, status, exit")

	for {
		fmt.Print("\n>")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "create":
			createOrder(client, reader)
		case "status":
			getStatus(client, reader)
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Неизвестная команда, просьба ввести одну из команд ещё раз: create, status, exit")
		}
	}
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Printf("%s: ", label)
	val, _ := reader.ReadString('\n')
	return strings.TrimSpace(val)
}

func createOrder(client orderpb.OrderServiceClient, reader *bufio.Reader) {
	userId := prompt(reader, "user_id")
	marketId := prompt(reader, "market_id")
	amountStr := prompt(reader, "amount")
	priceStr := prompt(reader, "price")
	userRole := prompt(reader, "user_role")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		fmt.Printf("Неверная цена: %v\n", err)
		return
	}
	quantity, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		fmt.Printf("Неправильное количество: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.CreateOrder(ctx, &orderpb.CreateOrderRequest{
		UserId:   userId,
		MarketId: marketId,
		Price:    price,
		Quantity: quantity,
		UserRole: userRole,
	})
	if err != nil {
		fmt.Printf("Возникла ошибка при создании заказа: %v\n", err)
		return
	}

	fmt.Printf("Заказ создан. ID: %s\n, Статус: %s\n", response.OrderId, response.OrderStatus)
}

func getStatus(client orderpb.OrderServiceClient, reader *bufio.Reader) {
	orderId := prompt(reader, "order_id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.GetOrderStatus(ctx, &orderpb.GetOrderStatusRequest{
		OrderId: orderId,
	})
	if err != nil {
		fmt.Printf("Возникла ошибка при получении статуса: %v\n", err)
		return
	}
	fmt.Printf("Статус заказа: %s\n", response.OrderStatus)
}
