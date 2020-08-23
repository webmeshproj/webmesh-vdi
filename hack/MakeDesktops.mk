# Builders

build-ubuntu-base:
	cd build/desktops/ubuntu && docker build . \
		-f Dockerfile.base \
		-t ${UBUNTU_BASE_IMAGE}

build-arch-base:
	cd build/desktops/arch && docker build . \
		-f Dockerfile.base \
		-t ${ARCH_BASE_IMAGE}

build-app-base:
	cd build/desktops/app-profiles && docker build . \
		-f Dockerfile.base \
		-t ${APP_PROFILE_BASE_IMAGE}

build-ubuntu-%:
	cd build/desktops/ubuntu && docker build . \
		-f Dockerfile.desktop \
		--build-arg BASE_IMAGE=${UBUNTU_BASE_IMAGE} \
		--build-arg DESKTOP_PACKAGE=$* \
		-t ${REPO}/${NAME}:ubuntu-$*-${VERSION}

build-arch-%:
	cd build/desktops/arch && docker build . \
		-f Dockerfile.$* \
		--build-arg BASE_IMAGE=${ARCH_BASE_IMAGE} \
		-t ${REPO}/${NAME}:arch-$*-${VERSION}

build-app-%:
	cd build/desktops/app-profiles && docker build . \
		-f Dockerfile.$* \
		--build-arg BASE_IMAGE=${APP_PROFILE_BASE_IMAGE} \
		-t ${REPO}/${NAME}:app-$*-${VERSION}

# Pushers

push-ubuntu-base: build-ubuntu-base
	docker push ${UBUNTU_BASE_IMAGE}

push-arch-base: build-arch-base
	docker push ${ARCH_BASE_IMAGE}

push-ubuntu-%: build-ubuntu-%
	docker push ${REPO}/${NAME}:ubuntu-$*-${VERSION}

push-arch-%: build-arch-%
	docker push ${REPO}/${NAME}:arch-$*-${VERSION}

push-app-%: build-app-%
	docker push ${REPO}/${NAME}:app-$*-${VERSION}

# Loaders

load-ubuntu-%: ${KIND} build-ubuntu-%
	$(call load_image,${REPO}/${NAME}:ubuntu-$*-${VERSION})

load-arch-%: ${KIND} build-arch-%
	$(call load_image,${REPO}/${NAME}:arch-$*-${VERSION})

load-app-%: ${KIND} build-app-%
	$(call load_image,${REPO}/${NAME}:app-$*-${VERSION})

#
# For building demo environment

demo-init:
	cd deploy/terraform && terraform init

demo-fmt:
	terraform fmt deploy/terraform

get_ext_ip = $(shell curl https://ifconfig.me 2> /dev/null)
TF_ARGS ?=
demo-%:
	cd deploy/terraform && \
		terraform $* \
			-var ext_ip=$(call get_ext_ip) \
			`[[ "$*" != "plan" ]] && echo -auto-approve` \
			$(TF_ARGS)