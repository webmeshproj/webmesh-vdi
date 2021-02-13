#!/bin/bash

DEFAULT_CPUS="1"
DEFAULT_MEMORY="1024"
DEFAULT_ARCH="x86_64"
DEFAULT_IMAGE="/disk/boot.img"
DEFAULT_SOCK_ADDR="unix:///tmp/vnc.sock"
DEFAULT_CLOUD_IMAGE="/tmp/cloud.img"

set -x

export CPUS=${CPUS:-$DEFAULT_CPUS}
export MEMORY=${MEMORY:-$DEFAULT_MEMORY}
export ARCH=${ARCH:-$DEFAULT_ARCH}
export BOOT_IMAGE=${BOOT_IMAGE:-$DEFAULT_IMAGE}
export CLOUD_IMAGE=${CLOUD_IMAGE:-$DEFAULT_CLOUD_IMAGE}
export VNC_SOCK_ADDR=${VNC_SOCK_ADDR:-$DEFAULT_SOCK_ADDR}

env 2>&1

if [ ! -f "${CLOUD_IMAGE}" ] ; then
  echo "Generating cloud-init image"
  cat << EOF | cloud-localds ${CLOUD_IMAGE} /dev/stdin
#cloud-config
bootcmd:
  - sed -i 's/%USER%/${USER}/g' /etc/gdm/custom.conf
users:
  - name: ${USER}
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    groups: sudo
    shell: /bin/bash
EOF
fi

# use overlay - https://wiki.archlinux.org/index.php/QEMU#Overlay_storage_images
# -cpu host should be dynamic according to arch or other options

qemu-system-${ARCH} \
  -enable-kvm \
  -serial stdio \
	-display vnc="${VNC_SOCK_ADDR}" \
	-cpu host -smp ${CPUS} -m ${MEMORY} \
  -usb -device usb-tablet \
	-device virtio-blk,drive=image -drive if=none,id=image,file="${BOOT_IMAGE}" \
	-device virtio-blk,drive=cloud -drive if=none,id=cloud,file="${CLOUD_IMAGE}" \
	-device virtio-net,netdev=user -netdev user,id=user \
  -monitor unix:/run/qemu-monitor.sock,server,nowait