BINARY := tkr
PKG := github.com/darkfactoryhq/tkr/cmd/tkr

.PHONY: build test lint clean install

build:
	go build -o $(BINARY) $(PKG)

test:
	go test -race ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

install:
	go install $(PKG)
