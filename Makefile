-include Makevars.mk
-include Manifests.mk

###
# Building
###

.PHONY: build
build: build-all

# Build all images
build-all: build-manager build-app build-novnc-proxy build-desktop-lxde

build-manager:
	$(call build_docker,manager,${MANAGER_IMAGE})

build-app:
	$(call build_docker,app,${APP_IMAGE})

build-novnc-proxy:
	$(call build_docker,novnc-proxy,${NOVNC_PROXY_IMAGE})

build-desktop-base:
	cd build/desktops/ubuntu && docker build . \
		-f Dockerfile.base \
		-t ${DESKTOP_BASE_IMAGE}

build-desktop-%:
	cd build/desktops/ubuntu && docker build . \
		-f Dockerfile.desktop \
		--build-arg BASE_IMAGE=${DESKTOP_BASE_IMAGE} \
		--build-arg DESKTOP_PACKAGE=$* \
		-t ${REPO}/${NAME}:$*-${VERSION}

ENTRYPOINT ?= /startup.sh
run-desktop:
	docker run \
		--rm -it \
		-e USER=ubuntu \
		-v /dev/shm:/dev/shm \
		-v /dev/snd:/dev/snd \
		-p 5900:5900 \
		--privileged \
		--entrypoint ${ENTRYPOINT} \
		${DESKTOP_IMAGE}

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

push-desktop-base: build-desktop-base
	docker push ${DESKTOP_BASE_IMAGE}

push-desktop-%: build-desktop-%
	docker push ${REPO}/${NAME}:$*-${VERSION}

chart-yaml:
	echo "$$CHART_YAML" > deploy/charts/kvdi/Chart.yaml

package-chart: ${HELM} chart-yaml
	cd deploy/charts && helm package kvdi
	rm deploy/charts/kvdi/Chart.yaml

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

load-all: load-manager load-app load-novnc-proxy load-desktop-lxde

load-manager: ${KIND} build-manager
	$(call load_image,${MANAGER_IMAGE})

load-app: ${KIND} build-app
	$(call load_image,${APP_IMAGE})

load-novnc-proxy: ${KIND} build-novnc-proxy
	$(call load_image,${NOVNC_PROXY_IMAGE})

load-desktop-%: ${KIND} build-desktop-%
	$(call load_image,${REPO}/${NAME}:$*-${VERSION})

# Deploys metallb load balancer to the kind cluster
test-ingress: ${KUBECTL}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} apply -f https://raw.githubusercontent.com/google/metallb/${METALLB_VERSION}/manifests/namespace.yaml
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} apply -f https://raw.githubusercontent.com/google/metallb/${METALLB_VERSION}/manifests/metallb.yaml
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} create secret generic -n metallb-system memberlist --from-literal=secretkey="`openssl rand -base64 128`" || echo
	echo "$$METALLB_CONFIG" | ${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} apply -f -

test-certmanager: ${KUBECTL} ${HELM}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.crds.yaml
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} create namespace cert-manager
	${HELM} repo add jetstack https://charts.jetstack.io && ${HELM} repo update
	${HELM} --kubeconfig ${KIND_KUBECONFIG} install \
		cert-manager jetstack/cert-manager \
		--namespace cert-manager \
		--version ${CERT_MANAGER_VERSION} \
		--set extraArgs[0]="--enable-certificate-owner-ref=true" \
		--wait

example-vdi-templates: ${KUBECTL}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} apply \
		-f deploy/examples/example-desktop-templates.yaml

full-test-cluster: test-cluster test-ingress test-certmanager

restart-manager: ${KUBECTL}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} delete pod -l component=kvdi-manager

restart-app: ${KUBECTL}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} delete pod -l vdiComponent=app

restart: restart-manager restart-app

clean-cluster: ${KUBECTL} ${HELM}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} delete --ignore-not-found -f deploy/examples
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} delete --ignore-not-found certificate --all
	${HELM} del kvdi

remove-cluster: ${KIND}
	${KIND} delete cluster --name ${CLUSTER_NAME}
	rm -f ${KIND_KUBECONFIG}

forward-app: ${KUBECTL}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} get pod | grep app | awk '{print$$1}' | xargs -I% ${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} port-forward % 8443

get-app-secret: ${KUBECTL}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} get secret kvdi-app-client -o json | jq -r '.data["ca.crt"]' | base64 -d > _bin/ca.crt
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} get secret kvdi-app-client -o json | jq -r '.data["tls.crt"]' | base64 -d > _bin/tls.crt
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} get secret kvdi-app-client -o json | jq -r '.data["tls.key"]' | base64 -d > _bin/tls.key

get-admin-password: ${KUBECTL}
	${KUBECTL} --kubeconfig ${KIND_KUBECONFIG} get secret kvdi-admin-secret -o json | jq -r .data.password | base64 -d && echo

# Builds and deploys the manager into a local kind cluster, requires helm.
.PHONY: deploy
deploy: ${HELM} package-chart
	${HELM} upgrade --install --kubeconfig ${KIND_KUBECONFIG} ${NAME} deploy/charts/${NAME}-${VERSION}.tgz ${HELM_ARGS} --wait

## Doc generation

${REFDOCS_CLONE}:
	mkdir -p $(dir ${REFDOCS})
	git clone https://github.com/ahmetb/gen-crd-api-reference-docs "${REFDOCS_CLONE}"

${REFDOCS}: ${REFDOCS_CLONE}
	cd "${REFDOCS_CLONE}" && go build .
	mv "${REFDOCS_CLONE}/gen-crd-api-reference-docs" "${REFDOCS}"
