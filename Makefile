ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.DEFAULT_GOAL := help
YARN ?= corepack yarn
DLV_VERSION ?= v1.27.2

.PHONY: help setup doctor api.deps api.dev api.debug api.vet api.test api.build debugger.deps client.dev docs.dev client.deps client.build client.typecheck client.test check multiuser.check test build

help:
	@echo "make setup           Install project dependencies and create .env"
	@echo "make doctor          Check local prerequisites"
	@echo "make api.dev         Run Go API on 127.0.0.1:4110"
	@echo "make client.dev      Run frontend on http://localhost:5173 (second terminal)"
	@echo "make api.debug       Run API under Delve in the terminal"
	@echo "make check           Run Go tests and frontend type checking"
	@echo "make build           Build API and frontend"
	@echo "make multiuser.check Validate multiuser journeys on disposable databases"

setup: doctor
	@test -f .env || cp .env.example .env
	$(MAKE) api.deps client.deps debugger.deps

doctor:
	go version
	node --version
	corepack --version
	@command -v cc >/dev/null || { echo "Install a C compiler (Xcode Command Line Tools on macOS)"; exit 1; }
	@pkg-config --modversion vips || { echo "Install libvips and pkg-config: brew install vips pkg-config (macOS)"; exit 1; }

api.deps:
	go mod download

debugger.deps:
	GOBIN="$(CURDIR)/.bin" go install github.com/go-delve/delve/cmd/dlv@$(DLV_VERSION)

api.dev:
	CGO_ENABLED=1 go run ./cmd/api

api.debug:
	CGO_ENABLED=1 ./.bin/dlv debug ./cmd/api

api.vet:
	CGO_ENABLED=1 go vet ./...

api.test:
	CGO_ENABLED=1 go test ./... -timeout 60s

api.build:
	CGO_ENABLED=1 go build -ldflags='-s -w' -o koito ./cmd/api/main.go

client.dev:
	cd client && $(YARN) run dev

docs.dev:
	cd docs && $(YARN) dev

client.deps:
	cd client && $(YARN) install --immutable

client.typecheck:
	cd client && $(YARN) run typecheck

client.build: client.deps
	cd client && $(YARN) run build
	find client/build/client -type f \( -name "*.js" -o -name "*.css" -o -name "*.html" -o -name "*.svg" \) -exec gzip -kf -9 {} \;

client.test:
	cd client && $(YARN) run test

check: api.vet api.test client.typecheck client.test

test: api.test

build: api.build client.build

# Starts isolated loopback APIs; never uses the database configured in .env.
multiuser.check: api.build client.build
	python3 scripts/validate_multiuser.py --binary ./koito

# Increment once per new commit, then include the synchronized version files.
.PHONY: version.check version.bump
version.check:
	python3 scripts/version.py

version.bump:
	python3 scripts/version.py --bump
