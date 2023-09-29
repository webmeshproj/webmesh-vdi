REPO    ?= ghcr.io/webmeshproj
VERSION ?= latest

CRD_OPTIONS ?= "crd:preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.11.1

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

TOOLCHAINS := $(LOCALBIN)/toolchains

.PHONY: toolchains
toolchains: $(TOOLCHAINS)
$(TOOLCHAINS): $(LOCALBIN)
	mkdir -p $(TOOLCHAINS)
	curl -SsL https://musl.cc/aarch64-linux-musl-cross.tgz | tar -C $(TOOLCHAINS) -xz
	curl -SsL https://musl.cc/x86_64-linux-musl-cross.tgz | tar -C $(TOOLCHAINS) -xz

clean-toolchains:
	rm -rf $(TOOLCHAINS)
	$(MAKE) $(TOOLCHAINS)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef


## BEGIN CUSTOM TARGETS

# includes
-include hack/Makevars.mk
-include hack/MakeDesktops.mk

# specifically needed for tests in github actions
# uses modified shell that doesn't support pipefail
SHELL := /bin/bash

BUNDLE = $(CURDIR)/deploy/bundle.yaml
# Create a single yaml file bundle
bundle: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${MANAGER_IMAGE}
	$(KUSTOMIZE) build config/crd > "$(BUNDLE)"
	$(KUSTOMIZE) build config/default >> "$(BUNDLE)"


##
## # Building Images
##
GO ?= go
GORELEASER ?= $(GO) run github.com/goreleaser/goreleaser@latest
BUILD_ARGS ?= --snapshot --clean
LDFLAGS ?= -s -w \
			-X 'github.com/kvdi/kvdi/pkg/version.Version=$(VERSION)' \
			-X 'github.com/kvdi/kvdi/pkg/version.GitCommit=$(shell git rev-parse HEAD)'

## make                    # Alias to `make build-all`.
## make build
.PHONY: build
build: build-all

## make build-all          # Build the manager, app, and nonvnc-proxy images.
build-all: build-manager build-app build-proxy

## make build-manager      # Build the manager docker image.
build-manager:
	VERSION=$(VERSION) $(GORELEASER) build --single-target --id manager $(BUILD_ARGS)
	$(call build_docker,manager,${MANAGER_IMAGE})

## make build-app          # Build the app docker image.
build-app:
	VERSION=$(VERSION) $(GORELEASER) build --single-target --id app $(BUILD_ARGS)
	$(call build_docker,app,${APP_IMAGE})

## make build-proxy  # Build the proxy image without audio support.
build-proxy:
	VERSION=$(VERSION) $(GORELEASER) build --single-target --id proxy $(BUILD_ARGS)
	$(call build_docker,proxy,${KVDI_PROXY_IMAGE})

## make build-audio-proxy # Build the proxy image with audio support.
build-audio-proxy:
	$(call build_docker,audio-proxy,${KVDI_PROXY_IMAGE})

build-kvdictl:
	cp deploy/bundle.yaml pkg/cmd/
	cd cmd/kvdictl && \
		go build -ldflags="$(LDFLAGS)" -o $(GOBIN)/kvdictl .

kvdictl-docs:
	go run hack/gen-kvdictl-docs.go

GOX ?= $(GOBIN)/gox
$(GOX):
	GO111MODULE=off go get github.com/mitchellh/gox

DIST ?= $(PWD)/dist
COMPILE_TARGETS ?= "darwin/amd64 linux/amd64 linux/arm linux/arm64 windows/amd64"
COMPILE_OUTPUT  ?= "$(DIST)/{{.Dir}}_{{.OS}}_{{.Arch}}"
dist-kvdictl: $(GOX)
	mkdir -p dist
	cp deploy/bundle.yaml pkg/cmd/
	cd cmd/kvdictl && \
		CGO_ENABLED=0 $(GOX) -osarch=$(COMPILE_TARGETS) -output=$(COMPILE_OUTPUT) -ldflags="$(LDFLAGS)"
	upx -9 $(DIST)/*

license-headers:
	for i in `find cmd/ -name '*.go'` ; do if ! grep -q Copyright $$i ; then cat hack/boilerplate.go.txt $$i > $$i.new && mv $$i.new $$i ; fi ; done
	for i in `find apis/ -name '*.go' -not -name zz_generated.deepcopy.go` ; do if ! grep -q Copyright $$i ; then cat hack/boilerplate.go.txt $$i > $$i.new && mv $$i.new $$i ; fi ; done
	for i in `find pkg/ -name '*.go' -not -name zz_generated.deepcopy.go` ; do if ! grep -q Copyright $$i ; then cat hack/boilerplate.go.txt $$i > $$i.new && mv $$i.new $$i ; fi ; done
	for i in `find ui/app/src -name '*.js'` ; do if ! grep -q Copyright $$i ; then cat hack/boilerplate.go.txt $$i > $$i.new && mv $$i.new $$i ; fi ; done
	for i in `find ui/app/src -name '*.vue'` ; do if ! grep -q Copyright $$i ; then cat hack/boilerplate.vue.txt $$i > $$i.new && mv $$i.new $$i ; fi ; done

##
## # Pushing images
##

## make push               # Alias to make push-all.
push: build-manager push-manager push-proxy

## make push-all           # Push the manager, app, and proxy images.
push-all: push-manager push-app push-proxy

## make push-manager       # Push the manager docker image.
push-manager: build-manager
	docker push ${MANAGER_IMAGE}

## make push-app           # Push the app docker image.
push-app: build-app
	docker push ${APP_IMAGE}

## make push-proxy  # Push the proxy docker image.
push-proxy: build-proxy
	docker push ${KVDI_PROXY_IMAGE}

##
## # Linting and Testing
##

GOLANGCI_LINT    ?= $(GOBIN)/golangci-lint
GOLANGCI_VERSION ?= v1.53.3
$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_VERSION)

## make lint   # Lint files
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run -v --timeout 600s

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

##
## # Local Testing with k3d
##

K3D ?= $(GOBIN)/k3d
$(K3D):
	curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | K3D_INSTALL_DIR=$(GOBIN) bash -s -- --no-sudo

# Ensures a repo-local installation of kubectl
$(KUBECTL):
	$(call download_bin,$(KUBECTL),${KUBECTL_DOWNLOAD_URL})

# Ensures a repo-local installation of helm
$(HELM):
	$(call get_helm)

## make test-cluster           # Make a local k3d cluster for testing.
test-cluster: $(K3D)
	$(K3D) cluster create $(CLUSTER_NAME) \
		--kubeconfig-update-default=false \
		--k3s-arg --disable=traefik@server:0 \
		--volume="/dev/shm:/dev/shm@server:0" \
		--volume="/dev/kvm:/dev/kvm@server:0" \
		-p 443:443@loadbalancer -p 5556:5556@loadbalancer
	mkdir -p $(shell dirname $(CLUSTER_KUBECONFIG))
	$(K3D) kubeconfig get $(CLUSTER_NAME) > $(CLUSTER_KUBECONFIG)

##
## make load-all               # Load all the docker images into the local k3d cluster.
load-all: load-manager load-app load-proxy

## make load-manager
load-manager: $(K3D) build-manager
	$(call load_image,${MANAGER_IMAGE})

## make load-app
load-app: $(K3D) build-app
	$(call load_image,${APP_IMAGE})

## make load-proxy
load-proxy: $(K3D) build-proxy
	$(call load_image,${KVDI_PROXY_IMAGE})

KUBECTL_K3D = $(KUBECTL) --kubeconfig ${CLUSTER_KUBECONFIG}
HELM_K3D = $(HELM) --kubeconfig ${CLUSTER_KUBECONFIG}

## make test-vault             # Deploys a vault instance into the k3d cluster.
test-vault: $(KUBECTL) $(HELM)
	$(HELM) repo add hashicorp https://helm.releases.hashicorp.com
	$(HELM_K3D) upgrade --install vault hashicorp/vault \
		--set server.dev.enabled=true \
		--wait
	$(KUBECTL_K3D) wait --for=condition=ready pod vault-0 --timeout=300s
	$(KUBECTL_K3D) exec -it vault-0 -- vault auth enable kubernetes
	$(KUBECTL_K3D) \
		config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | \
		base64 --decode > ca.crt
	$(KUBECTL_K3D) exec -it vault-0 -- vault write auth/kubernetes/config \
		token_reviewer_jwt=`$(KUBECTL_K3D) exec -it vault-0 -- cat /var/run/secrets/kubernetes.io/serviceaccount/token` \
		kubernetes_host=https://kubernetes.default:443 \
		kubernetes_ca_cert="`cat ca.crt`"
	rm ca.crt
	echo "$$VAULT_POLICY" | $(KUBECTL_K3D) exec -it vault-0 -- vault policy write kvdi -
	$(KUBECTL_K3D) exec -it vault-0 -- vault secrets enable --path=kvdi/ kv
	$(KUBECTL_K3D) exec -it vault-0 -- vault write auth/kubernetes/role/kvdi \
	    bound_service_account_names=kvdi-app,kvdi-manager \
	    bound_service_account_namespaces=default \
	    policies=kvdi \
	    ttl=1h

## make get-vault-token        # Returns a token that can be used to login to vault from the CLI or UI.
get-vault-token:
	$(KUBECTL_K3D) exec -it vault-0 -- vault token create | grep token | head -n1

## make test-ldap              # Deploys a test LDAP server into the k3d cluster.
test-ldap:
	$(KUBECTL_K3D) apply -f hack/glauth.yaml

## make test-oidc              # Deploys a test OIDC provider using dex
test-oidc:
	$(KUBECTL_K3D) apply -f hack/oidc.yaml

## make test-image-populator   # Deploys the imagepopulator CSI plugin
test-image-populator:
	$(KUBECTL_K3D) apply -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-image-populator/master/deploy/kubernetes-1.16/csi-image-daemonset.yaml
	$(KUBECTL_K3D) apply -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-image-populator/master/deploy/kubernetes-1.16/csi-image-csidriverinfo.yaml

##
## make deploy                 # Deploys kVDI into the local k3d cluster.
.PHONY: deploy
HELM_ARGS ?=
REPO_URL  ?= https://kvdi.github.io/helm-charts/charts
deploy: $(HELM)
	$(HELM_K3D) repo add kvdi $(REPO_URL)
	$(HELM_K3D) upgrade --install kvdi kvdi/kvdi \
		--set manager.image.repository=$(REPO)/vdi-manager \
		--set manager.image.tag=$(VERSION) \
		--set vdi.spec.app.image=$(APP_IMAGE) \
		--wait ${HELM_ARGS}

## make deploy-crds            # Deploys just the CRDs into the k3d cluster.
deploy-crds: manifests kustomize
	$(KUSTOMIZE) build config/crd | $(KUBECTL_K3D) apply -f -

## make deploy-with-vault      # Deploys kVDI into the k3d cluster with a vault configuration for the product of `test-vault`.
deploy-with-vault:
	$(MAKE) deploy HELM_ARGS="-f deploy/examples/example-vault-helm-values.yaml"

## make deploy-with-ldap       # Deploys kVDI into the k3d cluster with an LDAP configuration for the product of `test-ldap`.
deploy-with-ldap:
	$(MAKE) deploy HELM_ARGS="-f deploy/examples/example-ldap-helm-values.yaml"
	$(KUBECTL_K3D) apply -f hack/glauth-role.yaml

## make deploy-with-oidc       # Deploys kVDI into the k3d cluster with an OIDC configuration for the product of `test-oidc`.
##                             # Requires you set kvdi.local to the localhost in /etc/hosts.
deploy-with-oidc:
	$(MAKE) deploy HELM_ARGS="-f deploy/examples/example-oidc-helm-values.yaml"
	$(KUBECTL_K3D) apply -f hack/oidc-role.yaml

##
## make example-vdi-templates  # Deploys the example VDITemplates into the k3d cluster.
example-vdi-templates: $(KUBECTL)
	$(KUBECTL_K3D) apply \
		-f deploy/examples/example-desktop-templates.yaml

##
## make restart-manager    # Restart the manager pod.
restart-manager: $(KUBECTL)
	$(KUBECTL_K3D) delete pod -l app.kubernetes.io/name=kvdi

## make restart-app        # Restart the app pod.
restart-app: $(KUBECTL)
	$(KUBECTL_K3D) delete pod -l vdiComponent=app

## make restart            # Restart the manager and app pod.
restart: restart-manager restart-app

## make clean-cluster      # Remove all kVDI components from the cluster for a fresh start.
clean-cluster: $(KUBECTL) $(HELM)
	$(KUBECTL_K3D) delete --ignore-not-found certificate --all
	$(HELM_K3D) del kvdi

## make remove-cluster     # Deletes the k3d cluster.
remove-cluster: $(K3D)
	$(K3D) cluster delete $(CLUSTER_NAME)
	rm -f $(CLUSTER_KUBECONFIG)

##
## # Runtime Helpers
##

## make forward-app         # Run a kubectl port-forward to the app pod.
forward-app: $(KUBECTL)
	$(KUBECTL_K3D) port-forward --address 0.0.0.0 svc/kvdi-app 8443:443

## make get-app-secret      # Get the app client TLS certificate for debugging.
get-app-secret: $(KUBECTL)
	$(KUBECTL_K3D) get secret kvdi-app-client -o json | jq -r '.data["ca.crt"]' | base64 -d > _bin/ca.crt
	$(KUBECTL_K3D) get secret kvdi-app-client -o json | jq -r '.data["tls.crt"]' | base64 -d > _bin/tls.crt
	$(KUBECTL_K3D) get secret kvdi-app-client -o json | jq -r '.data["tls.key"]' | base64 -d > _bin/tls.key

## make get-admin-password  # Get the generated admin password for kVDI.
get-admin-password: $(KUBECTL)
	$(KUBECTL_K3D) get secret kvdi-admin-secret -o json | jq -r .data.password | base64 -d && echo

##
## # Doc generation
##

${REFDOCS_CLONE}:
	mkdir -p $(dir ${REFDOCS})
	git clone https://github.com/ahmetb/gen-crd-api-reference-docs "${REFDOCS_CLONE}"

${REFDOCS}: ${REFDOCS_CLONE}
	cd "${REFDOCS_CLONE}" && go build .
	mv "${REFDOCS_CLONE}/gen-crd-api-reference-docs" "${REFDOCS}"

## make api-docs            # Generate the CRD API documentation.
api-docs: ${REFDOCS}
	go mod vendor
	bash hack/update-api-docs.sh


##
## ######################################################################################
##
## make help                # Print this help message
help:
	@echo "# MAKEFILE USAGE" && echo
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

check-release:
	@if [[ "$(VERSION)" == "latest" ]] ; then \
		echo "You must specify a VERSION for release" ; exit 1 ; \
	fi

prep-release: check-release generate manifests api-docs kvdictl-docs bundle