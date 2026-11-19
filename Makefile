.PHONY: run build test clean docker-build

run:
	go run ./cmd/server/main.go

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/backoffice-api.exe ./cmd/server/main.go

test:
	go test ./... -v

clean:
	rm -rf bin/

docker-build:
	docker build -t backoffice-service:latest .

