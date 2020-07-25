#!/bin/bash

export HOME="/home/${USER}"
export USER_ID="9000"

# simple function for ensuring a basedir then writing a file from stdin
write () {
  mkdir -p "$(dirname ${1})"
  cp /dev/stdin "${1}"
}

echo "** Setting up user account: ${USER}"
useradd --uid 9000 --no-create-home --home-dir "${HOME}" --shell /bin/bash --user-group --groups adm ${USER}
passwd -d ${USER}
mkdir -p "${HOME}" && chown ${USER}: "${HOME}"
mkdir -p /run/user/${USER_ID} && chmod 0700 /run/user/${USER_ID} && chown ${USER}: /run/user/${USER_ID}

# This could be more cleverly done inside the image itself as opposed to at boot.
if [[ "${ENABLE_ROOT}" == "true" ]] ; then
  echo "** Allowing ${USER} to use root!"
  echo "${USER} ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
fi

# Set the VNC socket dir
if [[ -z "${VNC_SOCK_ADDR}" ]] ; then
  export VNC_SOCK_ADDR="/tmp/vnc.sock"
fi

# Pre-create the vnc socket directory and give it to the user
mkdir -p "$(dirname ${VNC_SOCK_ADDR})" && chown ${USER}: "$(dirname ${VNC_SOCK_ADDR})"

# This doesn't really work yet
if [[ -d "/dev/snd" ]] ; then
  echo "** Sound device is mounted inside the desktop, granting permissions"
  chgrp -R adm /dev/snd
  modprobe snd-aloop index=2
  export ALSADEV="hw:2,0"
  cp -r /root/.gtkrc-2.0 "${HOME}"
fi

# Iterate all var files and do substitution
find /etc/default -type f -exec \
    sed -i \
      -e "s|%USER%|${USER}|g" \
      -e "s|%UID%|${USER_ID}|g" \
      -e "s|%UNIX_SOCK%|${VNC_SOCK_ADDR}|g" \
      -e "s|%USER_ID%|${USER_ID}|g" \
      -e "s|%HOME%|${HOME}|g" {} +


# cat << EOF | write /etc/systemd/system/desktop.service.d/override.conf
# [Service]
# User=${USER}
# EOF

# Allow an automatic shell at the pts. This will trigger systemd-user as described
# below.
cat << EOF | write /etc/pam.d/login
auth       sufficient   pam_listfile.so item=tty sense=allow file=/etc/securetty onerr=fail apply=${USER}
auth       required     pam_securetty.so
auth       requisite    pam_nologin.so
-session   optional     pam_loginuid.so
-session   optional     pam_systemd.so
EOF

# This will cause pam-systemd to auto launch a systemd --user process which will
# in turn spawn all systemd user units that are part of default.target.
cat << EOF | write /etc/systemd/system/console-getty.service.d/override.conf
[Unit]
ConditionPathExists=

[Service]
ExecStart=
ExecStart=-/root/fakegetty.sh
Environment="USER=${USER}"
EOF
echo pts/9 >> /etc/securetty

export container=docker
exec /usr/lib/systemd/systemd
