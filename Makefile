# Defensive: on a host where GOROOT is set but wrong (as this sandbox's base
# image once was), an inherited GOROOT can break every `go` invocation. The
# Go binary infers its own root from its own location, so it's safe to
# unexport here regardless of whether the host actually needs it.
unexport GOROOT

# Install prefix for `make install`. bin/noggin ends up at $(PREFIX)/bin/noggin.
PREFIX ?= ~/.local

.PHONY: build vet fmt check test run install

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

# Copies the already-built bin/noggin into place; deliberately does not depend
# on `build` so `make install PREFIX=...` never triggers a rebuild on its own.
install:
	install -d $(PREFIX)/bin
	install bin/noggin $(PREFIX)/bin/noggin

# End-to-end checks against the real binary in a pty. Separate from `check`
# because they need a pty and are slower; they are the only way to verify what
# actually reaches a terminal — colour, cell geometry and startup handshake are
# all invisible to `go test` (see docs/dev/pty-harness.md).
.PHONY: check-pty
check-pty: build
	python3 test/pty/checks.py
