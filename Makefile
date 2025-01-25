build:
	@go build -o ./bin/fileserver

run : build
	@./bin/fileserver

test:
	@go test ./... -v

clean:
	@rm -rf ./bin
	@rm -rf ./test_store_root
	@go clean -i ./...

format:
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: build run test clean
