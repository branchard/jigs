# @see: https://stackoverflow.com/a/70550568
MAKEFLAGS += --no-print-directory

##@ Global
help: ## Show this help
	@# @see: https://www.avonture.be/blog/makefile-help/
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } ' $(MAKEFILE_LIST)

dev: ## Start a development container
	docker compose run --rm --remove-orphans dev sh -c "make help; sh"

build: ## Build the CLI tool
	go build -o ./build/jigs ./cmd/jigs

test: ## Run unit tests
	go test ./...
