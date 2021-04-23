build:
	go build -v ./cmd/backend/main.go
test:
	go test -v ./...
run:
	go run ./cmd/backend/main.go
