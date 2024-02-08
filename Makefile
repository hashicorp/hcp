ROOT_DIR = $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

# MOCKERY_DIRS are directories that contain a .mockery.yaml file. When go/mocks
# is run, the existing mocks will be deleted and new ones will be generated. If
# the mock file is generated in the same package as the actual implementation,
# store the mock file in MOCKERY_OUTPUT_FILES.
MOCKERY_DIRS=./ internal/commands/auth/ internal/pkg/api/iampolicy
MOCKERY_OUTPUT_DIRS=internal/pkg/api/mocks internal/commands/auth/mocks
MOCKERY_OUTPUT_FILES=internal/pkg/api/iampolicy/mock_setter.go \
					 internal/pkg/api/iampolicy/mock_resource_updater.go

default: help

.PHONY: go/lint
go/lint: ## Run the Go Linter
	@golangci-lint run

.PHONY: go/mocks
go/mocks: ## Generates Go mock files.
	@for dir in $(MOCKERY_OUTPUT_DIRS); do \
		rm -rf $$dir; \
    done

	@for file in $(MOCKERY_OUTPUT_FILES); do \
		rm -f $$file; \
    done

	@for dir in $(MOCKERY_DIRS); do \
		cd $(ROOT_DIR); \
		cd $$dir; \
		mockery; \
    done

.PHONY: test
test: ## Run the unit tests
	@go test -v -cover ./...

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""
