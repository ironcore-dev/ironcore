
# Image URL to use all building/pushing image targets
CONTROLLER_IMG ?= controller:latest
APISERVER_IMG ?= apiserver:latest
MACHINEPOOLLET_IMG ?= machinepoollet:latest
MACHINEBROKER_IMG ?= machinebroker:latest
VOLUMEPOOLLET_IMG ?= volumepoollet:latest
VOLUMEBROKER_IMG ?= volumebroker:latest

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
FILE="config/machinepoollet-broker/broker-rbac/role.yaml"
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	# onmetal-controller-manager
	$(CONTROLLER_GEN) rbac:roleName=manager-role webhook paths="./onmetal-controller-manager/controllers/...;./api/..." output:rbac:artifacts:config=config/controller/rbac

	# machinepoollet-broker
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths="./poollet/machinepoollet/controllers/..." output:rbac:artifacts:config=config/machinepoollet-broker/rbac
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths="./broker/machinebroker/..." output:rbac:artifacts:config=config/machinepoollet-broker/broker-rbac
	./hack/replace.sh config/machinepoollet-broker/broker-rbac/role.yaml 's/ClusterRole/Role/g;s/manager-role/broker-role/g'

	# volumepoollet-broker
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths="./poollet/volumepoollet/controllers/..." output:rbac:artifacts:config=config/volumepoollet-broker/rbac
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths="./broker/volumebroker/..." output:rbac:artifacts:config=config/volumepoollet-broker/broker-rbac
	./hack/replace.sh config/volumepoollet-broker/broker-rbac/role.yaml 's/ClusterRole/Role/g;s/manager-role/broker-role/g'

.PHONY: generate
generate: vgopath models-schema deepcopy-gen client-gen lister-gen informer-gen defaulter-gen conversion-gen openapi-gen applyconfiguration-gen
	VGOPATH=$(VGOPATH) \
	MODELS_SCHEMA=$(MODELS_SCHEMA) \
	DEEPCOPY_GEN=$(DEEPCOPY_GEN) \
	CLIENT_GEN=$(CLIENT_GEN) \
	LISTER_GEN=$(LISTER_GEN) \
	INFORMER_GEN=$(INFORMER_GEN) \
	DEFAULTER_GEN=$(DEFAULTER_GEN) \
	CONVERSION_GEN=$(CONVERSION_GEN) \
	OPENAPI_GEN=$(OPENAPI_GEN) \
	APPLYCONFIGURATION_GEN=$(APPLYCONFIGURATION_GEN) \
	./hack/update-codegen.sh

.PHONY: proto
proto: vgopath protoc-gen-gogo
	VGOPATH=$(VGOPATH) \
	PROTOC_GEN_GOGO=$(PROTOC_GEN_GOGO) \
	./hack/update-proto.sh

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
	rm -rf client-go/applyconfigurations
	rm -rf client-go/informers
	rm -rf client-go/listers
	rm -rf client-go/onmetalapi
	rm -rf client-go/openapi

.PHONY: add-license
add-license: addlicense ## Add license headers to all go files.
	find . -name '*.go' -exec $(ADDLICENSE) -c 'OnMetal authors' {} +

.PHONY: check-license
check-license: addlicense ## Check that every file has a license header present.
	find . -name '*.go' -exec $(ADDLICENSE) -check -c 'OnMetal authors' {} +

.PHONY: check
check: manifests generate add-license lint test # Generate manifests, code, lint, add licenses, test

.PHONY: docs
docs: gen-crd-api-reference-docs ## Run go generate to generate API reference documentation.
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir ./api/common/v1alpha1 -config ./hack/api-reference/common-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/common.md
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir ./api/compute/v1alpha1 -config ./hack/api-reference/compute-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/compute.md
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir ./api/storage/v1alpha1 -config ./hack/api-reference/storage-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/storage.md
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir ./api/networking/v1alpha1 -config ./hack/api-reference/networking-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/networking.md
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir ./api/ipam/v1alpha1 -config ./hack/api-reference/ipam-config.json -template-dir ./hack/api-reference/template -out-file ./docs/api-reference/ipam.md

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
docker-build: test docker-build-machinepoollet docker-build-machinebroker ## Build docker image with the manager.
	docker build --target apiserver -t ${APISERVER_IMG} .
	docker build --target manager -t ${CONTROLLER_IMG} .

.PHONY: docker-build-machinepoollet
docker-build-machinepoollet: ## Build machinepoollet image.
	docker build --target machinepoollet -t ${MACHINEPOOLLET_IMG} .

.PHONY: docker-build-machinebroker
docker-build-machinebroker: ## Build machinebroker image.
	docker build --target machinebroker -t ${MACHINEBROKER_IMG} .

.PHONY: docker-build-volumepoollet
docker-build-volumepoollet: ## Build volumepoollet image.
	docker build --target volumepoollet -t ${VOLUMEPOOLLET_IMG} .

.PHONY: docker-build-volumebroker
docker-build-volumebroker: ## Build volumebroker image.
	docker build --target volumebroker -t ${VOLUMEBROKER_IMG} .

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

.PHONY: kind-load-machinepoollet
kind-load-machinepoollet:
	kind load docker-image ${MACHINEPOOLLET_IMG}

.PHONY: kind-load-machinebroker
kind-load-machinebroker:
	kind load docker-image ${MACHINEBROKER_IMG}

.PHONY: kind-load-volumepoollet
kind-load-volumepoollet:
	kind load docker-image ${VOLUMEPOOLLET_IMG}

.PHONY: kind-load-volumebroker
kind-load-volumebroker:
	kind load docker-image ${VOLUMEBROKER_IMG}

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
DEEPCOPY_GEN ?= $(LOCALBIN)/deepcopy-gen
CLIENT_GEN ?= $(LOCALBIN)/client-gen
LISTER_GEN ?= $(LOCALBIN)/lister-gen
INFORMER_GEN ?= $(LOCALBIN)/informer-gen
DEFAULTER_GEN ?= $(LOCALBIN)/defaulter-gen
CONVERSION_GEN ?= $(LOCALBIN)/conversion-gen
OPENAPI_GEN ?= $(LOCALBIN)/openapi-gen
APPLYCONFIGURATION_GEN ?= $(LOCALBIN)/applyconfiguration-gen
VGOPATH ?= $(LOCALBIN)/vgopath
GEN_CRD_API_REFERENCE_DOCS ?= $(LOCALBIN)/gen-crd-api-reference-docs
ADDLICENSE ?= $(LOCALBIN)/addlicense
PROTOC_GEN_GOGO ?= $(LOCALBIN)/protoc-gen-gogo
MODELS_SCHEMA ?= $(LOCALBIN)/models-schema

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CODE_GENERATOR_VERSION ?= v0.24.3
CONTROLLER_TOOLS_VERSION ?= v0.9.0
VGOPATH_VERSION ?= v0.0.2
GEN_CRD_API_REFERENCE_DOCS_VERSION ?= v0.3.0
ADDLICENSE_VERSION ?= v1.1.0
PROTOC_GEN_GOGO_VERSION ?= v1.3.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: deepcopy-gen
deepcopy-gen: $(DEEPCOPY_GEN) ## Download deepcopy-gen locally if necessary.
$(DEEPCOPY_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/deepcopy-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(CODE_GENERATOR_VERSION)

.PHONY: client-gen
client-gen: $(CLIENT_GEN) ## Download client-gen locally if necessary.
$(CLIENT_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/client-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/client-gen@$(CODE_GENERATOR_VERSION)

.PHONY: lister-gen
lister-gen: $(LISTER_GEN) ## Download lister-gen locally if necessary.
$(LISTER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/lister-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/lister-gen@$(CODE_GENERATOR_VERSION)

.PHONY: informer-gen
informer-gen: $(INFORMER_GEN) ## Download informer-gen locally if necessary.
$(INFORMER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/informer-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/informer-gen@$(CODE_GENERATOR_VERSION)

.PHONY: defaulter-gen
defaulter-gen: $(DEFAULTER_GEN) ## Download defaulter-gen locally if necessary.
$(DEFAULTER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/defaulter-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/defaulter-gen@$(CODE_GENERATOR_VERSION)

.PHONY: conversion-gen
conversion-gen: $(CONVERSION_GEN) ## Download conversion-gen locally if necessary.
$(CONVERSION_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/conversion-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/conversion-gen@$(CODE_GENERATOR_VERSION)

.PHONY: openapi-gen
openapi-gen: $(OPENAPI_GEN) ## Download openapi-gen locally if necessary.
$(OPENAPI_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/openapi-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/openapi-gen@$(CODE_GENERATOR_VERSION)

.PHONY: applyconfiguration-gen
applyconfiguration-gen: $(APPLYCONFIGURATION_GEN) ## Download applyconfiguration-gen locally if necessary.
$(APPLYCONFIGURATION_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/applyconfiguration-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/applyconfiguration-gen@$(CODE_GENERATOR_VERSION)

.PHONY: vgopath
vgopath: $(VGOPATH) ## Download vgopath locally if necessary.
$(VGOPATH): $(LOCALBIN)
	test -s $(LOCALBIN)/vgopath || GOBIN=$(LOCALBIN) go install github.com/onmetal/vgopath@$(VGOPATH_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: openapi-extractor
openapi-extractor: $(OPENAPI_EXTRACTOR) ## Download openapi-extractor locally if necessary.
$(OPENAPI_EXTRACTOR): $(LOCALBIN)
	test -s $(LOCALBIN)/openapi-extractor || GOBIN=$(LOCALBIN) go install github.com/onmetal/openapi-extractor/cmd/openapi-extractor@latest

.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: $(GEN_CRD_API_REFERENCE_DOCS) ## Download gen-crd-api-reference-docs locally if necessary.
$(GEN_CRD_API_REFERENCE_DOCS): $(LOCALBIN)
	test -s $(LOCALBIN)/gen-crd-api-reference-docs || GOBIN=$(LOCALBIN) go install github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_REFERENCE_DOCS_VERSION)

.PHONY: addlicense
addlicense: $(ADDLICENSE) ## Download addlicense locally if necessary.
$(ADDLICENSE): $(LOCALBIN)
	test -s $(LOCALBIN)/addlicense || GOBIN=$(LOCALBIN) go install github.com/google/addlicense@$(ADDLICENSE_VERSION)

.PHONY: protoc-gen-gogo
protoc-gen-gogo: $(PROTOC_GEN_GOGO) ## Download protoc-gen-gogo locally if necessary.
$(PROTOC_GEN_GOGO): $(LOCALBIN)
	test -s $(LOCALBIN)/protoc-gen-gogo || GOBIN=$(LOCALBIN) go install github.com/gogo/protobuf/protoc-gen-gogo@$(PROTOC_GEN_GOGO_VERSION)

.PHONY: models-schema
models-schema: $(MODELS_SCHEMA) ## Install models-schema locally if necessary.
$(MODELS_SCHEMA): $(LOCALBIN)
	test -s $(LOCALBIN)/models-schema || GOBIN=$(LOCALBIN) go install github.com/onmetal/onmetal-api/models-schema
