build:
	go build -o bin/exchange

run: build
	./bin/exchange

test:
	go test -v ./...

curl:
	curl -X GET "https://api.example.com/data" -H "Authorization: Bearer YOUR_TOKEN" -H "Accept: application/json"
	curl -X POST "https://api.example.com/submit" -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{"key1": "value1", "key2": "value2"}'
