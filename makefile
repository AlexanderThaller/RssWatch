NAME = RssWatch

all:
	make format
	make style
	make test
	make build

format:
	find . -name "*.go" -not -path './Godeps/*' -type f -exec gofmt -s=true -w=true {} \;
	find . -name "*.go" -not -path './Godeps/*' -type f -exec goimports -w=true {} \;

style:
	find . -name "*.go" -not -path './Godeps/*' -type f -exec golint {} \;

test:
	go test ./...

build:
	go build -ldflags "-X main.buildTime `date +%s` -X main.buildVersion `git describe --always`"

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

dependencies_save:
	godep save ./...

dependencies_update:
	godep update ./...

dependencies_restore:
	godep restore ./...
