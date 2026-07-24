.PHONY: run verify test fmt clean

run:
	go run ./cmd/tiny

verify:
	go run verify.go

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	go clean ./...
	rm -f ./tmp/*
