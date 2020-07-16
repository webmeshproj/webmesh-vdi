-include hack/Makevars.mk
-include hack/Manifests.mk
-include hack/MakeDesktops.mk

###
# Building
###

.PHONY: build
build: build-all

# Build all images
build-all: build-manager build-app build-novnc-proxy

build-manager:
	$(call build_docker,manager,${MANAGER_IMAGE})

build-app:
	$(call build_docker,app,${APP_IMAGE})

build-novnc-proxy:
	$(call build_docker,novnc-proxy,${NOVNC_PROXY_IMAGE})


###
# Push images
###

push: build-manager push-manager push-novnc-proxy

push-all: push-manager push-app push-novnc-proxy

push-manager: build-manager
	docker push ${MANAGER_IMAGE}

push-app: build-app
	docker push ${APP_IMAGE}

push-novnc-proxy: build-novnc-proxy
	docker push ${NOVNC_PROXY_IMAGE}

chart-yaml:
	echo "$$CHART_YAML" > deploy/charts/kvdi/Chart.yaml

package-chart: ${HELM} chart-yaml
	cd deploy/charts && helm package kvdi

package-index:
	cd deploy/charts && helm repo index .

###
# Codegen
###

# Ensures a local copy of the manager-sdk
${OPERATOR_SDK}:
	$(call download_bin,${OPERATOR_SDK},${OPERATOR_SDK_URL})

# Generates deep copy code
generate: ${OPERATOR_SDK}
	GOROOT=${GOROOT} ${OPERATOR_SDK} generate k8s --verbose

# Generates CRD manifest
manifests: ${OPERATOR_SDK}
	${OPERATOR_SDK} generate crds --verbose

api-docs: ${REFDOCS}
	go mod vendor
	bash hack/update-api-docs.sh

###
# Linting
###

${GOLANGCI_LINT}:
	mkdir -p $(dir ${GOLANGCI_LINT})
	cd $(dir ${GOLANGCI_LINT}) && curl -JL ${GOLANGCI_DOWNLOAD_URL} | tar xzf -
	chmod +x $(dir ${GOLANGCI_LINT})golangci-lint-${GOLANGCI_VERSION}-$(shell uname | tr A-Z a-z)-amd64/golangci-lint
	ln -s golangci-lint-${GOLANGCI_VERSION}-$(shell uname | tr A-Z a-z)-amd64/golangci-lint ${GOLANGCI_LINT}

# Lint files
lint: ${GOLANGCI_LINT}
	${GOLANGCI_LINT} run -v --timeout 300s

# Tests
TEST_FLAGS ?= -v -cover -race -coverpkg=./... -coverprofile=profile.cov
test:
	go test ${TEST_FLAGS} ./...
	go tool cover -func profile.cov
	rm profile.cov

###
# Kind helpers for local testing
###

# Ensures a repo-local installation of kind
${KIND}:
	$(call download_bin,${KIND},${KIND_DOWNLOAD_URL})

${KUBECTL}:
	$(call download_bin,${KUBECTL},${KUBECTL_DOWNLOAD_URL})

${HELM}:
	$(call get_helm)

# Make a local test cluster and load a pre-baked emulator image into it
test-cluster: ${KIND}
	echo -e "$$KIND_CLUSTER_MANIFEST"
	echo "$$KIND_CLUSTER_MANIFEST" | ${KIND} \
			create cluster \
			--config - \
			--image kindest/node:${KUBERNETES_VERSION} \
			--name ${CLUSTER_NAME} \
			--kubeconfig ${KIND_KUBECONFIG}

# Loads the manager image into the local kind cluster
load: load-manager

load-all: load-manager load-app load-novnc-proxy

load-manager: ${KIND} build-manager
	$(call load_image,${MANAGER_IMAGE})

load-app: ${KIND} build-app
	$(call load_image,${APP_IMAGE})

load-novnc-proxy: ${KIND} build-novnc-proxy
	$(call load_image,${NOVNC_PROXY_IMAGE})

KUBECTL_KIND = ${KUBECTL} --kubeconfig ${KIND_KUBECONFIG}
HELM_KIND = ${HELM} --kubeconfig ${KIND_KUBECONFIG}

# Deploys metallb load balancer to the kind cluster
test-ingress: ${KUBECTL}
	${KUBECTL_KIND} apply -f https://raw.githubusercontent.com/google/metallb/${METALLB_VERSION}/manifests/namespace.yaml
	${KUBECTL_KIND} apply -f https://raw.githubusercontent.com/google/metallb/${METALLB_VERSION}/manifests/metallb.yaml
	${KUBECTL_KIND} create secret generic -n metallb-system memberlist --from-literal=secretkey="`openssl rand -base64 128`" || echo
	echo "$$METALLB_CONFIG" | ${KUBECTL_KIND} apply -f -

test-certmanager: ${KUBECTL} ${HELM}
	${KUBECTL_KIND} apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.crds.yaml
	${KUBECTL_KIND} create namespace cert-manager
	${HELM} repo add jetstack https://charts.jetstack.io && ${HELM} repo update
	${HELM_KIND} install \
		cert-manager jetstack/cert-manager \
		--namespace cert-manager \
		--version ${CERT_MANAGER_VERSION} \
		--set extraArgs[0]="--enable-certificate-owner-ref=true" \
		--wait

test-vault: ${KUBECTL} ${HELM}
	${HELM} repo add hashicorp https://helm.releases.hashicorp.com
	${HELM_KIND} upgrade --install vault hashicorp/vault \
		--set server.dev.enabled=true \
		--wait
	${KUBECTL_KIND} wait --for=condition=ready pod vault-0 --timeout=300s
	${KUBECTL_KIND} exec -it vault-0 -- vault auth enable kubernetes
	${KUBECTL_KIND} \
		config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | \
		base64 --decode > ca.crt
	${KUBECTL_KIND} exec -it vault-0 -- vault write auth/kubernetes/config \
		token_reviewer_jwt=`${KUBECTL_KIND} exec -it vault-0 -- cat /var/run/secrets/kubernetes.io/serviceaccount/token` \
		kubernetes_host=https://kubernetes.default:443 \
		kubernetes_ca_cert="`cat ca.crt`"
	rm ca.crt
	echo "$$VAULT_POLICY" | ${KUBECTL_KIND} exec -it vault-0 -- vault policy write kvdi -
	${KUBECTL_KIND} exec -it vault-0 -- vault secrets enable --path=kvdi/ kv
	${KUBECTL_KIND} exec -it vault-0 -- vault write auth/kubernetes/role/kvdi \
	    bound_service_account_names=kvdi-app,kvdi-manager \
	    bound_service_account_namespaces=default \
	    policies=kvdi \
	    ttl=1h

test-ldap:
	${KUBECTL_KIND} apply -f hack/glauth.yaml

example-vdi-templates: ${KUBECTL}
	${KUBECTL_KIND} apply \
		-f deploy/examples/example-desktop-templates.yaml

full-test-cluster: test-cluster test-ingress test-certmanager

restart-manager: ${KUBECTL}
	${KUBECTL_KIND} delete pod -l component=kvdi-manager

restart-app: ${KUBECTL}
	${KUBECTL_KIND} delete pod -l vdiComponent=app

restart: restart-manager restart-app

clean-cluster: ${KUBECTL} ${HELM}
	${KUBECTL_KIND} delete --ignore-not-found -f deploy/examples
	${KUBECTL_KIND} delete --ignore-not-found certificate --all
	${HELM_KIND} del kvdi

remove-cluster: ${KIND}
	${KIND} delete cluster --name ${CLUSTER_NAME}
	rm -f ${KIND_KUBECONFIG}

forward-app: ${KUBECTL}
	${KUBECTL_KIND} get pod | grep app | awk '{print$$1}' | xargs -I% ${KUBECTL_KIND} port-forward % 8443

get-app-secret: ${KUBECTL}
	${KUBECTL_KIND} get secret kvdi-app-client -o json | jq -r '.data["ca.crt"]' | base64 -d > _bin/ca.crt
	${KUBECTL_KIND} get secret kvdi-app-client -o json | jq -r '.data["tls.crt"]' | base64 -d > _bin/tls.crt
	${KUBECTL_KIND} get secret kvdi-app-client -o json | jq -r '.data["tls.key"]' | base64 -d > _bin/tls.key

get-admin-password: ${KUBECTL}
	${KUBECTL_KIND} get secret kvdi-admin-secret -o json | jq -r .data.password | base64 -d && echo

# Builds and deploys the manager into a local kind cluster, requires helm.
.PHONY: deploy
HELM_ARGS ?=
deploy: ${HELM} package-chart
	${HELM_KIND} upgrade --install ${NAME} deploy/charts/${NAME}-${VERSION}.tgz --wait ${HELM_ARGS}

## Doc generation

${REFDOCS_CLONE}:
	mkdir -p $(dir ${REFDOCS})
	git clone https://github.com/ahmetb/gen-crd-api-reference-docs "${REFDOCS_CLONE}"

${REFDOCS}: ${REFDOCS_CLONE}
	cd "${REFDOCS_CLONE}" && go build .
	mv "${REFDOCS_CLONE}/gen-crd-api-reference-docs" "${REFDOCS}"

${HELM_DOCS}:
	$(call get_helm_docs)

HELM_DOCS_VERSION ?= 0.13.0
helm-docs: ${HELM_DOCS} chart-yaml
	${HELM_DOCS}
