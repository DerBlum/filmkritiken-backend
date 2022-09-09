build:
	go build -v ./cmd/backend/main.go
test:
	go test -v ./...
test-coverage:
	go test -short -json -coverprofile=test-coverage.out ./... > ./sonar-report.json
run:
	bash -c "set -a; source ./config/local.env; set +a && go run cmd/backend/main.go"
run-docker:
	docker network create filmkritiken || true
	docker-compose up -d
	docker stop filmkritiken-backend || true
	docker rm filmkritiken-backend || true
	docker rmi filmkritiken-backend || true
	docker build -o output -f Dockerfile_build .
	docker build -t filmkritiken-backend .
	docker run --rm -it \
		--network filmkritiken \
		--name filmkritiken-backend \
		-p 8080:8080 \
		--env-file ./config/local-docker.env \
		filmkritiken-backend
