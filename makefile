all:
	make format
	make test
	make build

build:
	go build -ldflags "-X main.buildtime `date +%s`"

format:
	gofmt -s=true -w=true *.go

bench:
	go test -bench .

test:
	go test
