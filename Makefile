###
# Params.
###

APP_NAME := "committer"
BIN_NAME := $(APP_NAME)
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(BIN_NAME)

HAS_GODOC := $(shell command -v godoc;)
HAS_GOLANGCI := $(shell command -v golangci-lint;)
HAS_GORELEASER := $(shell command -v goreleaser;)

default: ci

###
# Entries.
###

build:
	@go build -o $(BIN_PATH) && echo "Build OK"

build-debug:
	@go build -gcflags="all=-N -l" -o $(BIN_PATH) && echo "Build OK"

ci: lint test coverage
ci-integration: lint test-integration coverage

coverage:
	@go tool cover -func=coverage.out && echo "Coverage OK"

doc:
ifndef HAS_GODOC
	@echo "Could not find godoc, installing it"
	@go install golang.org/x/tools/cmd/godoc@latest
endif
	@echo "Open http://localhost:6060/pkg/github.com/thalesfsp/committer/ in your browser\n"
	@godoc -http :6060

lint:
ifndef HAS_GOLANGCI
	@echo "Could not find golangci-list, installing it"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.1
endif
	@golangci-lint run -v -c .golangci.yml && echo "Lint OK"

release-local:
ifndef HAS_GORELEASER
	@echo "Could not find goreleaser, installing it"
	@go install github.com/goreleaser/goreleaser@v1.11.5
endif
	@goreleaser build --clean --snapshot && echo "Local release OK"

test:
	@VENDOR_ENVIRONMENT="testing" go test -timeout 30s -short -v -race -cover \
	-coverprofile=coverage.out ./... && echo "Test OK"

test-integration:
	@VENDOR_ENVIRONMENT="testing-integration" go test -timeout 120s -v -race \
	-cover -coverprofile=coverage.out ./... && echo "Integration test OK"

.PHONY: build
	ci \
	ci-integration \
	coverage \
	doc \
	lint \
	release-local \
	test \
	test-integration
