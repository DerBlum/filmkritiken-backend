build:
	go build -o ./build/filmkritiken-backend ./cmd/backend/main.go
test:
	go test ./...
test-coverage:
	go test -short -json -coverprofile=test-coverage.out ./...  > ./build/sonar-report.json
run:
	go run ./cmd/backend/main.go
