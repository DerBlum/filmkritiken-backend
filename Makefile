build:
	go build -o ./build/filmkritiken-backend ./cmd/backend/main.go
test:
	go test -v ./...
run:
	go run ./cmd/backend/main.go
pipeline-build:
	go build -v -o ./build/filmkritiken-backend ./cmd/backend/main.go
pipeline-test:
	go test -v ./...
