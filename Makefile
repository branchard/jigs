# @see: https://stackoverflow.com/a/70550568
MAKEFLAGS += --no-print-directory

##@ Global
help: ## Show this help
	@# @see: https://www.avonture.be/blog/makefile-help/
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } ' $(MAKEFILE_LIST)

dev: ## Start a development container
	UID=$$(id -u) GID=$$(id -g) docker compose run --rm --remove-orphans dev sh -c "make help; PS1='\\w\$$ ' bash --noprofile --norc"

build: ## Build the CLI tool
	go build -o ./build/jigs ./cmd/jigs

run: ## Run the CLI tool
	go run ./cmd/jigs

test: ## Run unit tests and end-to-end test
	# Run Go unit tests
	@go test ./...

	# Generate a .env with dummy values
	@rm -f .env
	@printf "aaa\n\n1234\ntest@test.com\n/tmp" | go run cmd/jigs/main.go examples/.env.dist examples/.env.dev

	# Ensure the .env exists
	@test -s .env || (echo "ERROR: The .env file was not created by the CLI tool" && exit 1)

	# Ensure the generated .env is correct
	@grep -q "^SALT_KEY=aaa" .env  || (echo "ERROR: The generated .env does not contain good SALT_KEY variable" && exit 1)
	@grep -q "^PREPOPULATED=prepopulated" .env  || (echo "ERROR: The generated .env does not contain good PREPOPULATED variable" && exit 1)
	@grep -q "^PASSWORD=1234" .env  || (echo "ERROR: The generated .env does not contain good PASSWORD variable" && exit 1)
	@grep -q "^SMTP_EMAIL=test@test.com" .env  || (echo "ERROR: The generated .env does not contain good SMTP_EMAIL variable" && exit 1)
	@grep -q "^DEPLOY_KEY_LOCATION=/tmp" .env  || (echo "ERROR: The generated .env does not contain good DEPLOY_KEY_LOCATION variable" && exit 1)
	@rm -f .env
	@echo "Done."

.DEFAULT:
	@echo "Command unknown: $@"
	@echo ""
	@$(MAKE) help