ROOT_DIR = $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

# MOCKERY_DIRS are directories that contain a .mockery.yaml file. When go/mocks
# is run, the existing mocks will be deleted and new ones will be generated. If
# the mock file is generated in the same package as the actual implementation,
# store the mock file in MOCKERY_OUTPUT_FILES.
MOCKERY_DIRS=./ internal/commands/auth/ internal/pkg/api/iampolicy internal/pkg/api/releasesapi
MOCKERY_OUTPUT_DIRS=internal/pkg/api/mocks internal/commands/auth/mocks internal/pkg/api/releasesapi/mocks
MOCKERY_OUTPUT_FILES=internal/pkg/api/iampolicy/mock_setter.go \
					 internal/pkg/api/iampolicy/mock_resource_updater.go

default: help

.PHONY: docs/gen
docs/gen: go/build ## Generate the HCP CLI documentation
	@mkdir -p web-docs
	@rm -rf web-docs/*
	@./bin/gendocs -output-dir web-docs/commands --output-nav-json web-docs/nav.json

.PHONY: docs/move
docs/move: go/build ## Copy the generated documentation to the HCP docs repository
	@./bin/mvdocs -generated-commands-dir web-docs/commands --generated-nav-json web-docs/nav.json \
		--dest-commands-dir ../hcp-docs/content/docs/cli/commands --dest-nav-json ../hcp-docs/data/docs-nav-data.json

.PHONY: gen/screenshot
gen/screenshot: go/install ## Create a screenshot of the HCP CLI
	@go run github.com/homeport/termshot/cmd/termshot@v0.2.7 -c -f assets/hcp.png -- hcp

.PHONY: gen/releasesapi
gen/releasesapi: ## Generate the releases API client
	@./hack/gen_releases_client.sh

.PHONY: go/build
go/build: ## Build the HCP CLI binary
	@CGO_ENABLED=0 go build -o bin/ ./...

.PHONY: go/install
go/install: ## Install the HCP CLI binary
	@go install

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

.PHONY: go/test
go/test: ## Run the unit tests
ifeq (, $(shell which gotestfmt))
	@go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
endif
	@go test -json -v  ./... 2>&1 | tee /tmp/gotest.log | gotestfmt -hide all

.PHONY: changelog/build
changelog/build:
ifeq (, $(shell which changelog-build))
	@go install github.com/hashicorp/go-changelog/cmd/changelog-build@latest
endif
ifeq (, $(LAST_RELEASE_GIT_TAG))
	@echo "Please set the LAST_RELEASE_GIT_TAG environment variable to generate a changelog section of notes since the last release."
else
	changelog-build -last-release ${LAST_RELEASE_GIT_TAG} -entries-dir .changelog/ -changelog-template .changelog/changelog.tmpl -note-template .changelog/release-note.tmpl -this-release $(shell git rev-parse HEAD)
endif

.PHONY: changelog/new-entry
changelog/new-entry:
ifeq (, $(shell which changelog-entry))
	@go install github.com/hashicorp/go-changelog/cmd/changelog-entry@latest
endif
ifeq (, $(CHANGELOG_PR))
	@echo "Please set the CHANGELOG_PR environment variable to the PR number to associate with the changelog.\n\nWill attempt to auto-detect. Branch must be already be pushed to GitHub and be part of a PR.\n"
	changelog-entry -dir .changelog -allowed-types-file .changelog/types.txt -pr -1
else
	changelog-entry -dir .changelog -allowed-types-file .changelog/types.txt -pr ${CHANGELOG_PR}
endif

.PHONY: changelog/check
changelog/check:
ifeq (, $(shell which changelog-check))
	@go install github.com/hashicorp/go-changelog/cmd/changelog-check@latest
endif
	@changelog-check

# Docker build and publish variables and targets
REGISTRY_NAME?=docker.io/hashicorp
IMAGE_NAME=hcp
IMAGE_TAG_DEV?=$(REGISTRY_NAME)/$(IMAGE_NAME):latest-$(shell git rev-parse --short HEAD)
DEV_DOCKER_GOOS ?= linux
DEV_DOCKER_GOARCH ?= arm64

.PHONY: docker/dev
# Builds from the locally generated binary in ./bin/
docker/dev: export GOOS=$(DEV_DOCKER_GOOS)
docker/dev: export GOARCH=$(DEV_DOCKER_GOARCH)
docker/dev: go/build
	docker buildx build \
		--load \
		--platform $(DEV_DOCKER_GOOS)/$(DEV_DOCKER_GOARCH) \
		--tag $(IMAGE_TAG_DEV) \
		--target=dev \
		.
	@echo "Successfully built $(IMAGE_TAG_DEV)"

crt-build:
	CGO_ENABLED=0 go build -o ${BIN_PATH} -trimpath -buildvcs=false \
    	-ldflags "-X github.com/hashicorp/hcp/version.GitCommit=${PRODUCT_REVISION}"

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""
