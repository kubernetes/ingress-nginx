all: update test check

update:
	go get -u -t . ./_examples

test:
	go test . ./_examples

check:
	golangci-lint run ./...

fmt:
	gofmt -s -w . ./_examples
