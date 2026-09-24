.PHONY: check build vet lint test cover examples

# check runs everything CI runs. Run it before pushing.
check: build vet lint test examples

build:
	go build ./...

vet:
	go vet ./...

# Lints the root module and every example module (each is its own module).
# Requires golangci-lint v2 (v1 cannot read .golangci.yml):
#   brew install golangci-lint
lint:
	golangci-lint run
	@set -e; for d in examples/*/; do \
		echo "== lint $$d =="; \
		(cd $$d && golangci-lint run); \
	done

test:
	go test ./... -race

# Coverage is not a gate. Some locally auto-downloaded toolchains lack `covdata`,
# which -cover needs for packages without tests, so it is kept out of `check`.
cover:
	go test ./... -race -cover

# Each example is its own module, so it is built and vetted separately.
examples:
	@set -e; for d in examples/*/; do \
		echo "== $$d =="; \
		(cd $$d && go build -o /dev/null ./... && go vet ./...); \
	done
