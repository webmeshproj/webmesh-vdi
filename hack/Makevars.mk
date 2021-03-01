# Go options
GO111MODULE ?= on
GOROOT      ?= `go env GOROOT`
GOPATH      ?= $(shell go env GOPATH)
GOBIN       ?= $(GOPATH)/bin

GIT_COMMIT ?= `git rev-parse HEAD`

# Image Names
BASE_IMAGE              ?= ${REPO}/${NAME}:build-base-${VERSION}
MANAGER_IMAGE           ?= ${REPO}/${NAME}:manager-${VERSION}
APP_IMAGE               ?= ${REPO}/${NAME}:app-${VERSION}
KVDI_PROXY_IMAGE        ?= ${REPO}/${NAME}:kvdi-proxy-${VERSION}
UBUNTU_BASE_IMAGE       ?= ${REPO}/${NAME}:ubuntu-base-latest
APP_PROFILE_BASE_IMAGE  ?= ${REPO}/${NAME}:app-base-latest
DOSBOX_IMAGE            ?= ${REPO}/${NAME}:dosbox-latest
QEMU_IMAGE              ?= ${REPO}/${NAME}:qemu-latest

# K3d Options
HELM_VERSION ?= v3.1.2
CLUSTER_NAME ?= kvdi
KUBERNETES_VERSION ?= v1.20.2
KUBECTL_DOWNLOAD_URL ?= https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_VERSION}/bin/$(shell uname | tr A-Z a-z)/amd64/kubectl
HELM_DOWNLOAD_URL ?= https://get.helm.sh/helm-${HELM_VERSION}-$(shell uname | tr A-Z a-z)-amd64.tar.gz
CLUSTER_KUBECONFIG ?= bin/kubeconfig.yaml
KUBECTL ?= bin/kubectl
HELM ?= bin/helm

# Refdocs
REFDOCS ?= bin/refdocs
REFDOCS_CLONE ?= $(dir ${REFDOCS})/gen-crd-api-reference-docs

###

# Functions

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
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--build-arg LDFLAGS="$(LDFLAGS)"
endef

define load_image
	$(K3D) image import --cluster=$(CLUSTER_NAME) $(1)
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

define VAULT_POLICY
path "kvdi/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
}
endef

export VAULT_POLICY