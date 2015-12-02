PROGRAM = "scribe"
PREFIX  = "bin"

default: test install

build: generate
	@echo -n "Building bin for $(PROGRAM)... "
	@godep go build -o $(PREFIX)/$(PROGRAM)
	@echo "done"

dist: generate
	@echo -n "Building static bin for $(PROGRAM)... "
	@CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags "-s -X github.com/olark/scribe/version.GitCommit=$$(git rev-parse HEAD)" -o $(PREFIX)/$(PROGRAM)
	@echo "done"

install: build
	@cp $(PREFIX)/$(PROGRAM) $(GOPATH)/bin/ || true

test: generate
	@godep go test -cover -v ./... -race --timeout=40s

generate:
	@godep go generate ./...

clean:
	@rm -rf bin/*

.PHONY: default build generate test clean install
