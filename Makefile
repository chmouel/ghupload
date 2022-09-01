BINARY := ghupload
MD_FILES := $(shell find . -type f -regex ".*md"  -not -regex '^./vendor/.*'  -not -regex '^./.vale/.*' -not -regex "^./.git/.*" -print)

LDFLAGS := -s -w
FLAGS += -ldflags "$(LDFLAGS)" -buildvcs=true

all: test lint build

test:
	@go test ./... -v

clean:
	@rm -rf bin/$(BINARY)

build: clean
	@echo "building."
	@mkdir -p bin/
	@go build  -v $(FLAGS)  -o bin/$(BINARY) $(BINARY).go

lint: lint-go

lint-go:
	@echo "linting."
	@golangci-lint run --disable gosimple --disable staticcheck --disable structcheck --disable unused

fmt:
	@go fmt `go list ./... | grep -v /vendor/`

fumpt:
	@gofumpt -w *.go

.PHONY: vendor
vendor:
	@go mod tidy
	@go mod vendor
