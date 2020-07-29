# Go options
GO111MODULE ?= auto
CGO_ENABLED ?= 0
GOROOT ?= `go env GOROOT`

GIT_COMMIT ?= `git rev-parse HEAD`

# Golang CI Options
GOLANGCI_VERSION ?= 1.23.8
GOLANGCI_LINT ?= _bin/golangci-lint
GOLANGCI_DOWNLOAD_URL ?= https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_VERSION}/golangci-lint-${GOLANGCI_VERSION}-$(shell uname | tr A-Z a-z)-amd64.tar.gz

# Operator SDK options
SDK_VERSION ?= v0.16.0

UNAME := $(shell uname)
ifeq ($(UNAME), Linux)
SDK_PLATFORM := linux-gnu
endif
ifeq ($(UNAME), Darwin)
SDK_PLATFORM := apple-darwin
endif

# Image Options
MANAGER_IMAGE ?= ${REPO}/${NAME}:manager-${VERSION}
APP_IMAGE ?= ${REPO}/${NAME}:app-${VERSION}
NOVNC_PROXY_IMAGE ?= ${REPO}/${NAME}:novnc-proxy-${VERSION}
UBUNTU_BASE_IMAGE ?= ${REPO}/${NAME}:ubuntu-base-${VERSION}
ARCH_BASE_IMAGE ?= ${REPO}/${NAME}:arch-base-${VERSION}
APP_PROFILE_BASE_IMAGE ?= ${REPO}/${NAME}:app-base-${VERSION}

# Operator SDK
OPERATOR_SDK ?= _bin/operator-sdk
OPERATOR_SDK_URL ?= https://github.com/operator-framework/operator-sdk/releases/download/${SDK_VERSION}/operator-sdk-${SDK_VERSION}-x86_64-${SDK_PLATFORM}

# Kind Options
KIND_VERSION ?= v0.7.0
KUBERNETES_VERSION ?= v1.18.4
METALLB_VERSION ?= v0.9.3
HELM_VERSION ?= v3.1.2
HELM_DOCS_VERSION ?= 0.13.0

CLUSTER_NAME ?= vdi
KIND_DOWNLOAD_URL ?= https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-$(shell uname)-amd64
KUBECTL_DOWNLOAD_URL ?= https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/$(shell uname | tr A-Z a-z)/amd64/kubectl
HELM_DOWNLOAD_URL ?= https://get.helm.sh/helm-${HELM_VERSION}-$(shell uname | tr A-Z a-z)-amd64.tar.gz
HELM_DOCS_DOWNLOAD_URL ?= https://github.com/norwoodj/helm-docs/releases/download/v${HELM_DOCS_VERSION}/helm-docs_${HELM_DOCS_VERSION}_$(shell uname)_x86_64.tar.gz
KIND_KUBECONFIG ?= _bin/kubeconfig.yaml
KIND ?= _bin/kind
KUBECTL ?= _bin/kubectl
HELM ?= _bin/helm
HELM_DOCS ?= _bin/helm-docs

# Gendocs
REFDOCS ?= _bin/refdocs
REFDOCS_CLONE ?= $(dir ${REFDOCS})/gen-crd-api-reference-docs

define download_bin
	mkdir -p $(dir $(1))
	curl -JL -o $(1) $(2)
	chmod +x $(1)
endef

define get_helm
	mkdir -p $(dir ${HELM})
	curl -JL $(HELM_DOWNLOAD_URL) | tar xzf - --to-stdout $(shell uname | tr A-Z a-z)-amd64/helm > $(HELM)
	chmod +x $(HELM)
endef

define get_helm_docs
	mkdir -p $(dir ${HELM_DOCS})
	curl -JL $(HELM_DOCS_DOWNLOAD_URL) | tar xzf - --to-stdout helm-docs > $(HELM_DOCS)
	chmod +x $(HELM_DOCS)
endef

define build_docker
	docker build . \
		-f build/Dockerfile.$(1) \
		-t $(2) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(shell git rev-parse HEAD)
endef

define load_image
	$(KIND) load --name $(CLUSTER_NAME) docker-image $(1)
endef

define CHART_YAML
apiVersion: v2
name: kvdi
description: A Kubernetes-Native Virtual Desktop Infrastructure
type: application
version: ${VERSION}
appVersion: ${VERSION}
endef
export CHART_YAML
