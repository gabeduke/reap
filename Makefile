.PHONY: test build run dev

test:
	go vet ./...
	go test -race ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/reap .

# Images are built and pushed to ghcr.io/gabeduke/reap by CI (.github/workflows/ci.yaml).
# Needs INFLUXDB_TOKEN; BROKER, DB, BUCKET, ORG, TOPIC, IGNORE are optional.
run:
	go run . --debug

dev:
	docker run -d -p 8086:8086 influxdb:2
