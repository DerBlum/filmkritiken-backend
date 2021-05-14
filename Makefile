build:
	go build -v ./cmd/backend/main.go
test:
	go test -v ./...
run:
	go run ./cmd/backend/main.go
run-docker:
	docker network create filmkritiken || true
	docker-compose up -d
	docker stop filmkritiken-backend || true
	docker rm filmkritiken-backend || true
	docker rmi filmkritiken-backend || true
	docker build -t filmkritiken-backend .
	docker run --rm -it \
		--network filmkritiken \
		--name filmkritiken-backend \
		-p 8080:8080 \
		filmkritiken-backend
