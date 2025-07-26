build:
	go build -o bin/banner-rotation.exe cmd/main.go

run:
	docker-compose up --build

down:
	docker-compose down

test:
	go test -race -count=100 -v ./...

lint:
	golangci-lint run