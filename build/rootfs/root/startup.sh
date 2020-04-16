#!/bin/bash

echo "** Setting up user account: ${USER}"
useradd --create-home --shell /bin/bash --user-group --groups adm ${USER}

if [[ "${ENABLE_ROOT}" == "true" ]] ; then
  echo "** Allowing ${USER} to use root!"
  echo "${USER} ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
fi

if [[ -z "${VNC_SOCK_ADDR}" ]] ; then
  export VNC_SOCK_ADDR="/tmp/vnc.sock"
fi

if [[ -d "/dev/snd" ]] ; then
  echo "** Sound device is mounted inside the desktop, granting permissions"
  chgrp -R adm /dev/snd
  modprobe snd-aloop index=2
  export ALSADEV="hw:2,0"
fi

export HOME="/home/${USER}"

cp -r /root/{.gtkrc-2.0,.asoundrc} ${HOME}

sed -i -e "s|%USER%|${USER}|g" -e "s|%HOME%|${HOME}|g" /etc/supervisor/conf.d/desktop.conf
sed -i -e "s|%UNIX_SOCK%|${VNC_SOCK_ADDR}|g" /etc/supervisor/conf.d/supervisord.conf

mkdir -p /var/log/supervisord
exec /bin/tini -- supervisord -n -c /etc/supervisor/supervisord.conf
