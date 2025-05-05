.DEFAULT_GOAL := test-all

CLI_DOCS_PATH=docs/cli-docs/
.PHONY: test-all
test-all: lint test

.PHONY: test
test:
	go test -race -count=1 ./...

.PHONY: lint
lint:
	golangci-lint run -v ./...

.PHONY: build
build:
	CGO_ENABLED=0 go build -o deck main.go

.PHONY: coverage
coverage:
	go test -race -v -count=1 -coverprofile=coverage.out.tmp ./...
	# ignoring generated code for coverage
	grep -E -v 'generated.deepcopy.go' coverage.out.tmp > coverage.out
	rm -f coverage.out.tmp

generate-cli-docs:
	mkdir -p $(CLI_DOCS_PATH)
	go run docs/*.go -output-path $(CLI_DOCS_PATH)

.PHONY: setup-kong
setup-kong:
	bash .ci/setup_kong.sh

.PHONY: setup-kong-ee
setup-kong-ee:
	bash .ci/setup_kong_ee.sh

.PHONY: test-integration
test-integration:
	go test -v -count=1 -tags=integration \
		-race \
		$(GOTESTFLAGS) \
		./tests/integration/...

.PHONY: test-performance
test-performance:
	go test -v -count=1 -tags=performance \
		-race \
		$(GOTESTFLAGS) \
		./tests/performance/...

.PHONY: clean
clean:
	bash .ci/clean_kong.sh
