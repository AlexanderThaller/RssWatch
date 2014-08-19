default:
	make format
	make test
	make style
	make coverage

format:
	gofmt -s=true -w=true *.go
	goimports -w=true *.go

style:
	golint *.go

test:
	go test -test.v=true

coverage:
	make clean
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

clean:
	rm -f *.out

bench:
	go test -test.benchmem=true -test.bench . 2> /dev/null
