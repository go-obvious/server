default: menu
.PHONY: default

menu:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(lastword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: menu

test:  ## Runs the unit tests
	@go test -v ./... -cover
.PHONY: test

clean: ## clean up 
	@go clean -cache
.PHONY: clean

fmt: ## Run go fmt against code
	@go fmt ./...
.PHONY: fmt

lint: ## Run the linter 
	@golangci-lint run
.PHONY: lint

vet: ## Run go vet against code
	@go vet ./...
.PHONY: vet

#######################################################################
# DYNAMIC TARGETS
# -----------------------------------------------
# For each cmd directory, make the make goals
$(applications):
	@$(MAKE) -C $@ $(MAKECMDGOALS)
.PHONY: $(applications)


