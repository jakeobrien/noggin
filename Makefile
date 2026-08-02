# Defensive: on a host where GOROOT is set but wrong (as this sandbox's base
# image once was, see .charm/knowledge/go-toolchain-env.md), an inherited
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

# End-to-end checks against the real binary in a pty. Separate from `check`
# because they need a pty and are slower; they are the only way to verify what
# actually reaches a terminal — colour, cell geometry and startup handshake are
# all invisible to `go test` (see .charm/knowledge/pty-verification.md).
.PHONY: check-pty
check-pty: build
	python3 test/pty/checks.py
