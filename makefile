all:
	make format
	make test
	make build

build:
	go build -ldflags "-X main.buildtime `date +%s`"

format:
	find . -name "*.go" -type f -exec gofmt -s=true -w=true {} \;
	find . -name "*.go" -type f -exec goimports -w=true {} \;

bench:
	go test -bench .

test:
	go test
