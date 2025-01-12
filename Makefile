build:
	go build -o bin/exchange

run:
	./bin/exchange

test:
	go test -v ./...