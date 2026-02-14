.PHONY: test coverage vet example

test:
	go test -race -timeout 60s ./...

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

vet:
	go vet ./...

example:
	go run ./examples/basic/main.go
