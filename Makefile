# Defensive: on a host where GOROOT is set but wrong (as this sandbox's base
# image once was, see .charm/cycle/findings/go-toolchain-env.md), an inherited
# GOROOT can break every `go` invocation. The Go binary infers its own root
# from its own location, so it's safe to unexport here regardless of whether
# the host actually needs it.
unexport GOROOT

.PHONY: build vet fmt check test run

build:
	go build -o bin/ ./...

vet:
	go vet ./...

fmt:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then echo "$$out"; echo "gofmt: files need formatting"; exit 1; fi

test:
	go test ./...

check: build vet fmt test

run:
	go run ./cmd/noggin
