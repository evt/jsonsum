run:
	go run --race ./cmd/...

test:
	go test -v ./...

dc:
	docker compose up --build --remove-orphans
