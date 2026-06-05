
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f order-service market-service

proto:
	protoc 	\
    	--go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    	--validate_out="lang=go,paths=source_relative:." \
   		-I Proto \
    	Proto/*.proto

order:
	go run .App/OrderService/OrderServiceCMD

market:
	go run .App/SpotInstrumentService/MarketServiceCMD

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/order-service  ./App/OrderService/OrderServiceCMD
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/market-service ./App/SpotInstrumentService/MarketServiceCMD

tidy:
	go mod tidy