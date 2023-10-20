# Builders

build-ubuntu-base:
	cd build/desktops/ubuntu && docker build . \
		-f Dockerfile.base \
		-t ${UBUNTU_BASE_IMAGE}

build-app-base:
	cd build/desktops/app-profiles && docker build . \
		-f Dockerfile.base \
		-t ${APP_PROFILE_BASE_IMAGE}

build-qemu:
	cd build/desktops/qemu && docker build . \
		-f Dockerfile.base \
		-t ${QEMU_IMAGE}

build-dosbox:
	cd build/desktops/dosbox && docker build . \
		-f Dockerfile.base \
		-t ${DOSBOX_IMAGE}

build-ubuntu-%: build-ubuntu-base
	cd build/desktops/ubuntu && docker build . \
		-f Dockerfile.desktop \
		--build-arg DESKTOP_PACKAGE=$* \
		-t ${REPO}/desktop-ubuntu-base-$*:latest

build-app-%:
	cd build/desktops/app-profiles && docker build . \
		-f Dockerfile.$* \
		-t ${REPO}/desktop-app-$*:latest

# Pushers

push-ubuntu-base: build-ubuntu-base
	docker push ${UBUNTU_BASE_IMAGE}

push-arch-base: build-arch-base
	docker push ${ARCH_BASE_IMAGE}

push-dosbox: build-dosbox
	docker push ${DOSBOX_IMAGE}

push-qemu: build-qemu
	docker push ${QEMU_IMAGE}

push-ubuntu-%: build-ubuntu-%
	docker push ${REPO}/desktop-ubuntu-base-$*:latest

push-app-%: build-app-%
	docker push ${REPO}/app-$*:latest

# Loaders

load-ubuntu-%: $(K3D) build-ubuntu-%
	$(K3D) image import --cluster=$(CLUSTER_NAME) $(REPO)/desktop-ubuntu-base-$*:latest

load-app-%: $(K3D) build-app-%
	$(K3D) image import --cluster=$(CLUSTER_NAME) $(REPO)/app-$*:latest

load-dosbox: $(K3D) build-dosbox
	$(K3D) image import --cluster=$(CLUSTER_NAME) $(DOSBOX_IMAGE)

load-qemu: $(K3D) build-qemu
	$(K3D) image import --cluster=$(CLUSTER_NAME) $(QEMU_IMAGE)
