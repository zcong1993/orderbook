generate:
	@go generate ./...

build: generate
	@echo "====> Build orderbook"
	@sh -c ./build.sh
