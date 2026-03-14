# @see: https://stackoverflow.com/a/70550568
MAKEFLAGS += --no-print-directory
.PHONY: help dev build run test

##@ Global
help: ## Show this help
	@# @see: https://www.avonture.be/blog/makefile-help/
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } ' $(MAKEFILE_LIST)

dev: ## Start a development container
	UID=$$(id -u) GID=$$(id -g) docker compose run --rm --remove-orphans dev sh -c "make help; PS1='\\w\$$ ' bash --noprofile --norc"

build: ## Build the CLI tool
	go build -ldflags "-X main.version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o ./build/jigs ./cmd/jigs

run: ## Run the CLI tool (usage: `make run ARGS="./e2e/.env.dist"`)
	go run ./cmd/jigs $(ARGS)

test: ## Run unit and end-to-end tests
	go test ./...

.DEFAULT:
	@echo "Command unknown: $@"
	@echo ""
	@$(MAKE) help
