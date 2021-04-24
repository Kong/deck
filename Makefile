.DEFAULT_GOAL := test-all

.PHONY: test-all
test-all: lint test

.PHONY: test
test:
	go test -race ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: build
build:
	CGO_ENABLED=0 go build -o deck main.go

.PHONY: verify-codegen
verify-codegen:
	./scripts/verify-codegen.sh
	./scripts/verify-deepcopy-gen.sh

.PHONY: update-codegen
update-codegen:
	go generate ./...
	./scripts/update-deepcopy-gen.sh

.PHONY: test-coverage
test-coverage:
	go test -race -v -count=1 -coverprofile=coverage.out.tmp ./...
	# ignoring generated code for coverage
	grep -E -v 'generated.deepcopy.go' coverage.out.tmp > coverage.out
	rm -f coverage.out.tmp 

.PHONY: integration-test-coverage
test-coverage:
	go test -tags=integration -race -v -count=1 -coverprofile=coverage.out.tmp ./...
	# ignoring generated code for coverage
	grep -E -v 'generated.deepcopy.go' coverage.out.tmp > coverage.out
	rm -f coverage.out.tmp 

.PHONY: setup-kong
setup-kong:
	bash .ci/setup_kong.sh