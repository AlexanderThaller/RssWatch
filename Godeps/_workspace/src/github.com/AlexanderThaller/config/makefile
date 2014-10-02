NAME = PROJECTNAME

all:
	make format
	make test

format:
	find . -name "*.go" -type f -exec gofmt -s=true -w=true {} \;
	find . -name "*.go" -type f -exec goimports -w=true {} \;

test:
	go test ./...

build:
	go build -ldflags "-X main.buildtime `date +%s` -X main.version `git describe --always`"

clean:
	rm -f "$(NAME)"
	rm -f *.pprof
	rm -f *.pdf

install:
	cp "$(NAME)" /usr/local/bin

graphs:
	make callgraph
	make memograph

callgraph:
	go tool pprof --pdf "$(NAME)" cpu.pprof > callgraph.pdf

memograph:
	go tool pprof --pdf "$(NAME)" mem.pprof > memograph.pdf

coverage:
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

benchmark:
	go test -test.benchmem=true -test.bench . 2> /dev/null
