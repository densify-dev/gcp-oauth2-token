TRIM_PATH ?= $(shell realpath ..)
LINTER_VERSION ?= v2.6.2
TARGETARCH ?= amd64

all: clean tools lint test build

clean:
	@rm -rf build/

lint: tools
	@golangci-lint run ./...

generate:
	@go generate ./...

build: clean lint build-dir
	@mkdir -p build/$(TARGETARCH)
	@GOOS=linux GOARCH=$(TARGETARCH) CGO_ENABLED=0 go build -trimpath -gcflags=-trimpath="${TRIM_PATH}" -asmflags=-trimpath="${TRIM_PATH}" -ldflags="-w -s"  -o build/$(TARGETARCH)/gcp-oauth2-token ./cmd

tools:
	@echo "Installing tools..."
	@go mod tidy
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(LINTER_VERSION)

build-dir:
	@if [ ! -d build/ ]; then echo "Creating build directory..."; mkdir -p build/; fi

test: build-dir
	@echo "Running tests..."
	@go test -v -race $(shell go list ./...) --coverprofile=build/coverage.out
	@go tool cover -html=build/coverage.out -o build/coverage.html
