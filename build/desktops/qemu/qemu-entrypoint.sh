#!/bin/bash

DEFAULT_CPUS="1"
DEFAULT_MEMORY="1024"
DEFAULT_ARCH="x86_64"
DEFAULT_IMAGE="/disk/boot.img"
DEFAULT_SOCK_ADDR="unix:///tmp/display.sock"
DEFAULT_CLOUD_IMAGE="/tmp/cloud.img"

set -x

export HOME="${HOME:-/home/$USER}"
export CPUS=${CPUS:-$DEFAULT_CPUS}
export MEMORY=${MEMORY:-$DEFAULT_MEMORY}
export ARCH=${ARCH:-$DEFAULT_ARCH}
export BOOT_IMAGE=${BOOT_IMAGE:-$DEFAULT_IMAGE}
export CLOUD_IMAGE=${CLOUD_IMAGE:-$DEFAULT_CLOUD_IMAGE}
export DISPLAY_SOCK_ADDR=${DISPLAY_SOCK_ADDR:-$DEFAULT_SOCK_ADDR}

ENVIRONMENT=$(env | base64 --wrap 0)


if [[ "${DISPLAY_SOCK_ADDR}" =~ ^unix://* ]] ; then
  socket_type="unix"
  socket_address=${DISPLAY_SOCK_ADDR#"unix://"}
else
  socket_type="tcp"
  address=${DISPLAY_SOCK_ADDR#"tcp://"}
  tcp_address=$(echo ${address} | cut -d ":" -f1)
  tcp_port=$(echo ${address} | cut -d ":" -f2)
fi

if [[ -n "${SPICE_DISPLAY}" ]] && [[ "${SPICE_DISPLAY}" == "true" ]] ; then
  spice_opts="disable-ticketing,image-compression=lz,jpeg-wan-compression=always,zlib-glz-wan-compression=always,playback-compression=off,streaming-video=filter,seamless-migration=on"
  if [[ "${socket_type}" == "unix" ]] ; then
    DISPLAY_ARGS="-vga qxl -spice unix,addr=${socket_address},${spice_opts} \
      -device virtio-serial -chardev spicevmc,id=vdagent,debug=0,name=vdagent"
  else
    DISPLAY_ARGS="-vga qxl -spice addr=${tcp_address},port=${tcp_port},${spice_opts} \
      -device virtio-serial -chardev spicevmc,id=vdagent,debug=0,name=vdagent"
  fi
else
  if [[ "${socket_type}" == "unix" ]] ; then
    DISPLAY_ARGS="-vnc unix:${socket_address}"
  else
    DISPLAY_ARGS="-vnc ${tcp_address}:${tcp_port}"
  fi
fi

if [[ ! -f "${CLOUD_IMAGE}" ]] ; then cat << EOF | cloud-localds ${CLOUD_IMAGE} /dev/stdin
#cloud-config

growpart:
  mode: auto
  devices: ["/"]
  ignore_growroot_disabled: false

mounts:
  - ["kvdi_run", "/run/kvdi", "9p", "trans=virtio,rw,msize=104857600,nodevmap,access=client,_netdev"]
  - ["home", "${HOME}", "9p", "trans=virtio,rw,dfltuid=${UID},dfltgid=${UID},msize=104857600,access=client,_netdev"]

write_files:

  - path: /etc/environment
    permissions: "0644"
    encoding: b64
    content: ${ENVIRONMENT}

  - path: "${HOME}/.config/gnome-initial-setup-done"
    permissions: "0644"
    content: "yes"

  - path: /etc/lightdm/lightdm.conf.d/12-autologin.conf
    permissions: "0644"
    content: |
      [SeatDefaults]
      autologin-user=${USER}
      autologin-session=xfce
  
  - path: /usr/share/gnome-session/sessions/gdm-shell.session
    permissions: "0644"
    content: |
      [GNOME Session]
      Name=KVDI

  - path: /etc/gdm3/custom.conf
    permissions: "0644"
    content: |
      [daemon]
      AutomaticLoginEnable=true
      AutomaticLogin=${USER}

      [security]

      [xdmcp]

      [chooser]

      [debug]

groups:
  - autologin
  - nopasswdlogin

users:
  - name: ${USER}
    sudo: ALL=(ALL) NOPASSWD:ALL
    uid: "${UID}"
    groups: wheel, autologin, nopasswdlogin
    homedir: ${HOME}
    shell: /bin/bash
EOF
fi

# use overlay - https://wiki.archlinux.org/index.php/QEMU#Overlay_storage_images
# -cpu host should be dynamic according to arch or other options

qemu-system-${ARCH} \
  -enable-kvm \
  -bios /usr/share/qemu/OVMF.fd \
	${DISPLAY_ARGS} \
	-cpu host -smp ${CPUS} -m ${MEMORY} \
  -usb -device usb-tablet \
	-device virtio-blk,drive=image -drive if=none,id=image,file="${BOOT_IMAGE}" \
	-device virtio-blk,drive=cloud -drive if=none,id=cloud,format=raw,file="${CLOUD_IMAGE}" \
  -fsdev local,id=home,path=${HOME},security_model=none -device virtio-9p,fsdev=home,mount_tag=home \
  -fsdev local,id=kvdi_run,path=/run,security_model=none -device virtio-9p,fsdev=kvdi_run,mount_tag=kvdi_run \
	-device virtio-net,netdev=user -netdev user,id=user \
  -monitor unix:/run/qemu-monitor.sock,server,nowait \
  -chardev socket,path=/run/qga.sock,server,nowait,id=qga0 \
  -device virtio-serial \
  -device virtserialport,chardev=qga0,name=org.qemu.guest_agent.0

# {"execute":"guest-exec", "arguments":{"path":"/bin/bash","arg": ["-c", "touch /tmp/test.txt"]}}