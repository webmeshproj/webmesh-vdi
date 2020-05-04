#!/bin/bash

export HOME="/home/${USER}"

echo "** Setting up user account: ${USER}"
useradd --uid 9000 --no-create-home --home-dir "${HOME}" --shell /bin/bash --user-group --groups adm ${USER}
mkdir -p "${HOME}" && chown ${USER}: "${HOME}"

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

cp -r /root/{.gtkrc-2.0,.asoundrc} "${HOME}"

find /etc/supervisor/conf.d/ -type f -exec \
    sed -i -e "s|%USER%|${USER}|g" -e "s|%HOME%|${HOME}|g" {} +
sed -i -e "s|%UNIX_SOCK%|${VNC_SOCK_ADDR}|g" /etc/supervisor/conf.d/supervisord.conf

mkdir -p /var/log/supervisord
exec /bin/tini -- supervisord -n -c /etc/supervisor/supervisord.conf
