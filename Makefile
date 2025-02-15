run:
	@go run ./cmd/app/main.go 

build:
	@go build -o ./bin/main.out ./cmd/app/main.go