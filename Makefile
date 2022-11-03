
# Image URL to use all building/pushing image targets
CONTROLLER_IMG ?= controller:latest
APISERVER_IMG ?= apiserver:latest

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24.1

# Docker image name for the mkdocs based local development setup
IMAGE=onmetal-api/documentation

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

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./onmetal-controller-manager/controllers/...;./apis/..." output:rbac:artifacts:config=config/controller/rbac

.PHONY: generate
generate:
	./hack/update-codegen.sh

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint on the code.
	golangci-lint run ./...

.PHONY: clean
clean: ## Clean any artifacts that can be regenerated.
	rm -rf generated/*

.PHONY: addlicense
addlicense: ## Add license headers to all go files.
	find . -name '*.go' -exec go run github.com/google/addlicense -c 'OnMetal authors' {} +

.PHONY: checklicense
checklicense: ## Check that every file has a license header present.
	find . -name '*.go' -exec go run github.com/google/addlicense  -check -c 'OnMetal authors' {} +

.PHONY: check
check: manifests generate addlicense lint test # Generate manifests, code, lint, add licenses, test

.PHONY: docs
docs: ## Run go generate to generate API reference documentation.
	go run github.com/ahmetb/gen-crd-api-reference-docs -api-dir ./apis/common/v1alpha1 -config ./hack/api-reference/common-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/common.md
	go run github.com/ahmetb/gen-crd-api-reference-docs -api-dir ./apis/compute/v1alpha1 -config ./hack/api-reference/compute-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/compute.md
	go run github.com/ahmetb/gen-crd-api-reference-docs -api-dir ./apis/storage/v1alpha1 -config ./hack/api-reference/storage-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/storage.md
	go run github.com/ahmetb/gen-crd-api-reference-docs -api-dir ./apis/networking/v1alpha1 -config ./hack/api-reference/networking-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/networking.md
	go run github.com/ahmetb/gen-crd-api-reference-docs -api-dir ./apis/ipam/v1alpha1 -config ./hack/api-reference/ipam-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/ipam.md

.PHONY: start-docs
start-docs: ## Start the local mkdocs based development environment.
	docker build -t $(IMAGE) -f docs/Dockerfile .
	docker run -p 8000:8000 -v `pwd`/:/docs $(IMAGE)

.PHONY: clean-docs
clean-docs: ## Remove all local mkdocs Docker images (cleanup).
	docker container prune --force --filter "label=project=onmetal_api_documentation"

.PHONY: test
test: manifests generate fmt vet test-only ## Run tests.

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
.PHONY: test-only
test-only: envtest ## Run *only* the tests - no generation, linting etc.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

.PHONEY: openapi-extractor
extract-openapi: openapi-extractor
	$(OPENAPI_EXTRACTOR) --apiserver-package="github.com/onmetal/onmetal-api/onmetal-apiserver/cmd/apiserver" --apiserver-build-opts=mod --apiservices="./config/apiserver/apiservice/bases" --output="./gen"

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager ./onmetal-controller-manager/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./onmetal-controller-manager/main.go

.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	docker build --target apiserver -t ${CONTROLLER_IMG} .
	docker build --target manager -t ${APISERVER_IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${CONTROLLER_IMG}
	docker push ${APISERVER_IMG}

##@ Deployment

.PHONY: install
install: manifests kustomize ## Install API server & API services into the K8s cluster specified in ~/.kube/config. This requires APISERVER_IMG to be available for the cluster.
	cd config/apiserver/server && $(KUSTOMIZE) edit set image apiserver=${APISERVER_IMG}
	kubectl apply -k config/apiserver/default

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall API server & API services from the K8s cluster specified in ~/.kube/config.
	kubectl delete -k config/apiserver/default

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/controller/manager && $(KUSTOMIZE) edit set image controller=${CONTROLLER_IMG}
	kubectl apply -k config/controller/default

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -k config/controller/default

##@ Kind Deployment plumbing

.PHONY: kind-build-apiserver
kind-build-apiserver: ## Build the apiserver for usage in kind.
	docker build --target apiserver -t apiserver .

.PHONY: kind-build-controller
kind-build-controller: ## Build the controller for usage in kind.
	docker build --target manager -t controller .

.PHONY: kind-build
kind-build: kind-build-apiserver kind-build-controller ## Build the apiserver and controller for usage in kind.

.PHONY: kind-load-apiserver
kind-load-apiserver: ## Load the apiserver image into the kind cluster.
	kind load docker-image apiserver

.PHONY: kind-load-controller
kind-load-controller: ## Load the controller image into the kind cluster.
	kind load docker-image controller

.PHONY: kind-load
kind-load: kind-load-apiserver kind-load-controller ## Load the apiserver and controller in kind.

.PHONY: kind-restart-apiserver
kind-restart-apiserver: ## Restart the apiserver in kind. Useless if the manifests are not in place (deployed e.g. via kind-apply / kind-deploy).
	kubectl -n onmetal-system delete rs -l control-plane=apiserver

.PHONY: kind-restart-controller
kind-restart-controller: ## Restart the controller in kind. Useless if the manifests are not in place (deployed e.g. via kind-apply / kind-deploy).
	kubectl -n onmetal-system delete rs -l control-plane=controller-manager

.PHONY: kind-restart
kind-restart: kind-restart-apiserver kind-restart-controller ## Restart the apiserver and controller in kind. Restart is useless if the manifests are not in place (deployed e.g. via kind-apply / kind-deploy).

.PHONY: kind-build-load-restart-controller
kind-build-load-restart-controller: kind-build-controller kind-load-controller kind-restart-controller ## Build, load and restart the controller in kind. Restart is useless if the manifests are not in place (deployed e.g. via kind-apply / kind-deploy).

.PHONY: kind-build-load-restart-apiserver
kind-build-load-restart-apiserver: kind-build-apiserver kind-load-apiserver kind-restart-apiserver ## Build, load and restart the apiserver in kind. Restart is useless if the manifests are not in place (deployed e.g. via kind-apply / kind-deploy).

.PHONY: kind-build-load-restart
kind-build-load-restart: kind-build-load-restart-apiserver kind-build-load-restart-controller ## Build load and restart the apiserver and controller in kind. Restart is useless if the manifests are not in place (deployed e.g. via kind-apply / kind-deploy).

.PHONY: kind-apply-apiserver
kind-apply-apiserver: manifests kustomize ## Applies the apiserver manifests in kind. Caution, without loading the images, the pods won't come up. Use kind-install / kind-deploy for a deployment including loading the images.
	kubectl apply -k config/apiserver/kind

.PHONY: kind-install
kind-install: kind-build-load-restart-apiserver kind-apply-apiserver ## Build and load and apply apiserver in kind. Restarts apiserver if it was present.

.PHONY: kind-uninstall
kind-uninstall: manifests kustomize ## Uninstall API server & API services from the K8s cluster specified in ~/.kube/config.
	kubectl delete -k config/apiserver/kind

.PHONY: kind-apply
kind-apply: ## Apply the config in kind. Caution: Without loading the images, the pods won't come up. Use kind-deploy for a deployment including loading the images.
	kubectl apply -k config/kind

.PHONY: kind-delete
kind-delete: ## Delete the config from kind.
	kubectl delete -k config/kind

.PHONY: kind-deploy
kind-deploy: kind-build-load-restart kind-apply ## Build and load apiserver and controller into the kind cluster, then apply the config. Restarts apiserver / controller if they were present.

##@ Tools

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
OPENAPI_EXTRACTOR ?= $(LOCALBIN)/openapi-extractor

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.9.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: openapi-extractor
openapi-extractor: $(OPENAPI_EXTRACTOR) ## Download openapi-extractor locally if necessary.
$(OPENAPI_EXTRACTOR): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/onmetal/openapi-extractor/cmd/openapi-extractor@latest
