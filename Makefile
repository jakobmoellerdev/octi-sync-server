ifeq ($(OS),Windows_NT)
    $(error This Makefile does not support windows)
endif

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Image URL to use all building/pushing image targets
IMG := ghcr.io/jakob-moeller-cloud/octi-sync-server:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build and Development

.PHONY: build
build: ## Build manager binary.
	go build -o octi-sync-server main.go

.PHONY: test
test: ## Run tests.
	go test ./... -covermode atomic -cover

.PHONY: lint
lint: golangci-lint ## Run golangci-lint against code.
	$(LOCALBIN)/golangci-lint run

.PHONY: run
run: ## Run a controller from your host.
	go run ./main.go

.PHONY: generate
generate: mockgen ## Generate code
	PATH=${PATH}:${ROOT_DIR}/bin go generate ./...

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} .

##@ Deployment

.PHONY: deploy
deploy: kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build deploy/kustomize | kubectl apply -f -

.PHONY: deploy-dry-run
deploy-dry-run: kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build deploy/kustomize

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

##@ Tool Binaries

KUSTOMIZE_VERSION ?= v5.1.1
KUSTOMIZE ?= $(LOCALBIN)/kustomize
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

GOLANG_CI_LINT_VERSION ?= v1.54.2
GOLANG_CI_LINT = $(LOCALBIN)/golangci-lint
.PHONY: golangci-lint
golangci-lint: $(GOLANG_CI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANG_CI_LINT): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)

MOCKGEN_VERSION ?= $(shell awk '/go\.uber\.org\/mock/ {print substr($$2, 2)}' go.mod)
MOCKGEN = $(LOCALBIN)/mockgen
.PHONY: mockgen
mockgen: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install go.uber.org/mock/mockgen@v$(MOCKGEN_VERSION)

OAPI_CODEGEN_VERSION ?= $(shell awk '/github\.com\/deepmap\/oapi-codegen/ {print substr($$2, 2)}' go.mod)
OAPI_CODEGEN = $(LOCALBIN)/oapi-codegen
.PHONY: oapi-codegen
oapi-codegen: $(OAPI_CODEGEN) ## Download mockgen locally if necessary.
$(OAPI_CODEGEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v$(OAPI_CODEGEN_VERSION)
