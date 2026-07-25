.PHONY: build test test-coverage run run-docker docker-up wait-mongo seed

build:
	go build -v ./cmd/backend/main.go

test:
	go test -v ./...

test-coverage:
	go test -short -json -coverprofile=test-coverage.out ./... > ./sonar-report.json

run:
	bash -c "set -a; source ./config/local.env; set +a && go run cmd/backend/main.go"

docker-up:
	docker network create filmkritiken || true
	docker compose up -d

wait-mongo:
	@echo "Warte auf MongoDB readiness..."
	@for i in $$(seq 1 30); do \
		if [ "$$(docker inspect --format='{{.State.Health.Status}}' filmkritiken-mongodb 2>/dev/null)" = "healthy" ]; then \
			echo "MongoDB ist betriebsbereit."; \
			break; \
		fi; \
		sleep 1; \
	done

seed:
	bash -c "set -a; source ./config/local-docker.env; MONGODB_CONNECTION_URI='mongodb://localhost:27017/?directConnection=true'; set +a && go run ./cmd/seed"

run-docker: docker-up wait-mongo
	@$(MAKE) seed || true
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


