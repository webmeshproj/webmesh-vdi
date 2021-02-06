REPO ?= ghcr.io/tinyzimmer
NAME = kvdi
VERSION ?= latest

# includes
-include hack/Makevars.mk
-include hack/MakeDesktops.mk

# specifically needed for tests in github actions
# uses modified shell that doesn't support pipefail
SHELL := /bin/bash

## # Building Images
##

## make                    # Alias to `make build-all`.
## make build
.PHONY: build
build: build-all

## make build-all          # Build the manager, app, and nonvnc-proxy images.
build-all: build-manager build-app build-kvdi-proxy

## make build-base         # Builds the base image that contains common go build dependies.
build-base:
	$(call build_docker,base,$(BASE_IMAGE))

## make build-manager      # Build the manager docker image.
build-manager: build-base
	$(call build_docker,manager,${MANAGER_IMAGE})

## make build-app          # Build the app docker image.
build-app: build-base
	$(call build_docker,app,${APP_IMAGE})

## make build-kvdi-proxy  # Build the kvdi-proxy image.
build-kvdi-proxy: build-base
	$(call build_docker,kvdi-proxy,${KVDI_PROXY_IMAGE})

##
## # Pushing images
##

## make push               # Alias to make push-all.
push: build-manager push-manager push-kvdi-proxy

## make push-all           # Push the manager, app, and kvdi-proxy images.
push-all: push-manager push-app push-kvdi-proxy

## make push-manager       # Push the manager docker image.
push-manager: build-manager
	docker push ${MANAGER_IMAGE}

## make push-app           # Push the app docker image.
push-app: build-app
	docker push ${APP_IMAGE}

## make push-kvdi-proxy  # Push the kvdi-proxy docker image.
push-kvdi-proxy: build-kvdi-proxy
	docker push ${KVDI_PROXY_IMAGE}

##
## # Helm Chart Functions
##

## make chart-yaml     # Generate the Chart.yaml from the template in hack/Makevars.mk.
chart-yaml:
	echo "$$CHART_YAML" > deploy/charts/kvdi/Chart.yaml

## make package-chart  # Packages the helm chart.
package-chart: ${HELM} chart-yaml
	cd deploy/charts && helm package kvdi

## make package-index  # Create the helm repo package index.
package-index:
	cd deploy/charts && helm repo index .

##
## # Codegen Functions
##

${OPERATOR_SDK}:
	$(call download_bin,${OPERATOR_SDK},${OPERATOR_SDK_URL})

## make generate            # Generates deep copy code for the k8s apis.
generate: ${OPERATOR_SDK}
	GOROOT=${GOROOT} ${OPERATOR_SDK} generate k8s --verbose

## make manifests           # Generates CRD manifest.
manifests: ${OPERATOR_SDK}
	${OPERATOR_SDK} generate crds --verbose

##
## # Linting and Testing
##

GOLANGCI_LINT    ?= $(GOBIN)/golangci-lint
GOLANGCI_VERSION ?= v1.33.0
$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_VERSION)

## make lint   # Lint files
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run -v --timeout 600s

## make test   # Run unit tests
GO_PACKAGES ?= $(shell go list ./... | grep -v 'pkg/apis' | xargs | sed -e 's/ /,/g')
TEST_FLAGS ?= -v -cover -coverpkg="$(GO_PACKAGES)" -coverprofile=profile.cov 
test:
	@set -o pipefail && go test ${TEST_FLAGS} ./... \
		 | sed ''/PASS/s//$$(printf "\033[32mPASS\033[0m")/'' \
		 | sed ''/FAIL/s//$$(printf "\033[31mFAIL\033[0m")/''
	@go tool cover -func profile.cov
	@rm profile.cov

##
## # Local Testing with k3d
##

K3D ?= $(GOBIN)/k3d
$(K3D):
	curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | K3D_INSTALL_DIR=$(GOBIN) bash -s -- --no-sudo

# Ensures a repo-local installation of kubectl
${KUBECTL}:
	$(call download_bin,${KUBECTL},${KUBECTL_DOWNLOAD_URL})

# Ensures a repo-local installation of helm
${HELM}:
	$(call get_helm)

## make test-cluster           # Make a local k3d cluster for testing.
test-cluster: $(K3D)
	$(K3D) cluster create $(CLUSTER_NAME) \
		--update-default-kubeconfig=false \
		--k3s-server-arg="--disable=traefik" \
		--volume="/dev/shm:/dev/shm@server[0]" \
		-p 443:443@loadbalancer -p 5556:5556@loadbalancer
	$(K3D) kubeconfig get $(CLUSTER_NAME) > $(CLUSTER_KUBECONFIG)

##
## make load-all               # Load all the docker images into the local k3d cluster.
load-all: load-manager load-app load-kvdi-proxy

## make load-manager
load-manager: $(K3D) build-manager
	$(call load_image,${MANAGER_IMAGE})

## make load-app
load-app: $(K3D) build-app
	$(call load_image,${APP_IMAGE})

## make load-kvdi-proxy
load-kvdi-proxy: $(K3D) build-kvdi-proxy
	$(call load_image,${KVDI_PROXY_IMAGE})

KUBECTL_K3D = ${KUBECTL} --kubeconfig ${CLUSTER_KUBECONFIG}
HELM_K3D = ${HELM} --kubeconfig ${CLUSTER_KUBECONFIG}

## make test-vault             # Deploys a vault instance into the k3d cluster.
test-vault: ${KUBECTL} ${HELM}
	${HELM} repo add hashicorp https://helm.releases.hashicorp.com
	${HELM_K3D} upgrade --install vault hashicorp/vault \
		--set server.dev.enabled=true \
		--wait
	${KUBECTL_K3D} wait --for=condition=ready pod vault-0 --timeout=300s
	${KUBECTL_K3D} exec -it vault-0 -- vault auth enable kubernetes
	${KUBECTL_K3D} \
		config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | \
		base64 --decode > ca.crt
	${KUBECTL_K3D} exec -it vault-0 -- vault write auth/kubernetes/config \
		token_reviewer_jwt=`${KUBECTL_K3D} exec -it vault-0 -- cat /var/run/secrets/kubernetes.io/serviceaccount/token` \
		kubernetes_host=https://kubernetes.default:443 \
		kubernetes_ca_cert="`cat ca.crt`"
	rm ca.crt
	echo "$$VAULT_POLICY" | ${KUBECTL_K3D} exec -it vault-0 -- vault policy write kvdi -
	${KUBECTL_K3D} exec -it vault-0 -- vault secrets enable --path=kvdi/ kv
	${KUBECTL_K3D} exec -it vault-0 -- vault write auth/kubernetes/role/kvdi \
	    bound_service_account_names=kvdi-app,kvdi-manager \
	    bound_service_account_namespaces=default \
	    policies=kvdi \
	    ttl=1h

## make get-vault-token        # Returns a token that can be used to login to vault from the CLI or UI.
get-vault-token:
	${KUBECTL_K3D} exec -it vault-0 -- vault token create | grep token | head -n1

## make test-ldap              # Deploys a test LDAP server into the k3d cluster.
test-ldap:
	${KUBECTL_K3D} apply -f hack/glauth.yaml

## make test-oidc              # Deploys a test OIDC provider using dex
test-oidc:
	${KUBECTL_K3D} apply -f hack/oidc.yaml

##
## make deploy                 # Deploys kVDI into the local k3d cluster.
.PHONY: deploy
HELM_ARGS ?=
deploy: ${HELM} chart-yaml
	${HELM_K3D} upgrade --install ${NAME} deploy/charts/kvdi --wait ${HELM_ARGS}

helm-template: ${HELM} chart-yaml
	${HELM_K3D} template ${NAME} deploy/charts/kvdi ${HELM_ARGS}

## make deploy-with-vault      # Deploys kVDI into the k3d cluster with a vault configuration for the product of `test-vault`.
deploy-with-vault:
	$(MAKE) deploy HELM_ARGS="-f deploy/examples/example-vault-helm-values.yaml"

## make deploy-with-ldap       # Deploys kVDI into the k3d cluster with an LDAP configuration for the product of `test-ldap`.
deploy-with-ldap:
	$(MAKE) deploy HELM_ARGS="-f deploy/examples/example-ldap-helm-values.yaml"
	${KUBECTL_K3D} apply -f hack/glauth-role.yaml

## make deploy-with-oidc       # Deploys kVDI into the k3d cluster with an OIDC configuration for the product of `test-oidc`.
##                             # Requires you set kvdi.local to the localhost in /etc/hosts.
deploy-with-oidc:
	$(MAKE) deploy HELM_ARGS="-f deploy/examples/example-oidc-helm-values.yaml"
	${KUBECTL_K3D} apply -f hack/oidc-role.yaml

##
## make example-vdi-templates  # Deploys the example VDITemplates into the k3d cluster.
example-vdi-templates: ${KUBECTL}
	${KUBECTL_K3D} apply \
		-f deploy/examples/example-desktop-templates.yaml

##
## make restart-manager    # Restart the manager pod.
restart-manager: ${KUBECTL}
	${KUBECTL_K3D} delete pod -l component=kvdi-manager

## make restart-app        # Restart the app pod.
restart-app: ${KUBECTL}
	${KUBECTL_K3D} delete pod -l vdiComponent=app

## make restart            # Restart the manager and app pod.
restart: restart-manager restart-app

## make clean-cluster      # Remove all kVDI components from the cluster for a fresh start.
clean-cluster: ${KUBECTL} ${HELM}
	${KUBECTL_K3D} delete --ignore-not-found certificate --all
	${HELM_K3D} del kvdi

## make remove-cluster     # Deletes the k3d cluster.
remove-cluster: $(K3D)
	$(K3D) cluster delete $(CLUSTER_NAME)
	rm -f $(CLUSTER_KUBECONFIG)

##
## # Runtime Helpers
##

## make forward-app         # Run a kubectl port-forward to the app pod.
forward-app: ${KUBECTL}
	${KUBECTL_K3D} port-forward --address 0.0.0.0 svc/kvdi-app 8443:443

## make get-app-secret      # Get the app client TLS certificate for debugging.
get-app-secret: ${KUBECTL}
	${KUBECTL_K3D} get secret kvdi-app-client -o json | jq -r '.data["ca.crt"]' | base64 -d > _bin/ca.crt
	${KUBECTL_K3D} get secret kvdi-app-client -o json | jq -r '.data["tls.crt"]' | base64 -d > _bin/tls.crt
	${KUBECTL_K3D} get secret kvdi-app-client -o json | jq -r '.data["tls.key"]' | base64 -d > _bin/tls.key

## make get-admin-password  # Get the generated admin password for kVDI.
get-admin-password: ${KUBECTL}
	${KUBECTL_K3D} get secret kvdi-admin-secret -o json | jq -r .data.password | base64 -d && echo

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

## make helm-docs           # Generates the helm chart documentation.
helm-docs: ${HELM_DOCS} chart-yaml
	docker run --rm -v "$(PWD)/deploy/charts/kvdi:/helm-docs" -u $(shell id -u) jnorwood/helm-docs:latest


check-release:
	@if [[ "$(VERSION)" == "latest" ]] ; then \
		echo "You must specify a VERSION for release" ; exit 1 ; \
	fi

prep-release: check-release generate manifests api-docs helm-docs package-chart package-index

##
## ######################################################################################
##
## make help                # Print this help message
help:
	@echo "# MAKEFILE USAGE" && echo
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'



TEST_DOCKER_IMAGE ?= "kvdi-tests"

test-docker-build:
	docker build . \
	    -f .github/tests.Dockerfile \
	    -t $(TEST_DOCKER_IMAGE)

TEST_CMD ?= /bin/bash
run-in-docker: test-docker-build
	docker run -it --rm --privileged \
	    -v /lib/modules:/lib/modules:ro \
	    -v /sys:/sys:ro \
	    -v /usr/src:/usr/src:ro \
	    -v "$(PWD)":/workspace \
		-w /workspace \
		-e HOME=/tmp \
	    $(TEST_DOCKER_IMAGE) $(TEST_CMD)

test-in-docker:
	$(MAKE) run-in-docker TEST_CMD="make test"

lint-in-docker:
	$(MAKE) run-in-docker TEST_CMD="make lint"