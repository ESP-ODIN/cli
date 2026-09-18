.PHONY: run build test vet fmt docker-build docker-dev

run:
	go run .

build:
	go build -o bin/ .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal ui main.go

docker-build:
	docker build -t cli .

# Lance le CLI dans un conteneur avec hot-reload (bind mount du code).
# Pas de port : cli n'expose aucun serveur réseau.
docker-dev:
	docker run --rm -it \
		-v "$(PWD)":/app \
		-v /app/tmp \
		cli
