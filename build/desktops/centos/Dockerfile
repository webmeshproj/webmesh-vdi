FROM centos:latest as base-system

RUN curl -o /etc/yum.repos.d/winswitch.repo https://winswitch.org/downloads/CentOS/winswitch.repo

# Core Dependencies
RUN \
    yum upgrade -y \
    && yum install --allowerasing -y \
        epel-release sudo net-tools alsa-utils tigervnc-server xpra pulseaudio pavucontrol vim \
    && yum --allowerasing -y groupinstall workstation \
    && cd /usr/lib/systemd/system/sysinit.target.wants/ \
    && ls | grep -v systemd-tmpfiles-setup | xargs rm -f $1 \
    && rm -f /usr/lib/systemd/system/multi-user.target.wants/* \
        /etc/systemd/system/*.wants/* \
        /usr/lib/systemd/system/local-fs.target.wants/* \
        /usr/lib/systemd/system/sockets.target.wants/*udev* \
        /usr/lib/systemd/system/sockets.target.wants/*initctl* \
        /usr/lib/systemd/system/basic.target.wants/* \
        /usr/lib/systemd/system/anaconda.target.wants/* \
        /usr/lib/systemd/system/systemd-update-utmp* \
    && yum clean packages \
    && yum clean metadata \
    && yum clean all
    

# Filesystem and systemd setup
COPY rootfs /
RUN chmod +x /usr/local/sbin/init && chmod +x /usr/local/sbin/fakegetty \
  && systemctl --user --global enable display.service \
  && systemctl --user --global enable pulseaudio \
  && systemctl enable user-init \
  && systemctl set-default graphical.target

# Entry
WORKDIR /root
VOLUME ["/sys/fs/cgroup"]
ENTRYPOINT ["/usr/local/sbin/init"]
