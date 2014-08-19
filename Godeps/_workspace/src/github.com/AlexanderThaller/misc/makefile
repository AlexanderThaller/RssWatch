default:
	make format
	make test

format:
	gofmt -s=true -w=true *.go
	goimports -w=true *.go
	golint *.go

test:
	go test

bench:
	go test -test.bench .

coverage:
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out
