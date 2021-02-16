FROM ubuntu:latest

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update \
    && apt-get dist-upgrade -y \
    && apt-get install -y --no-install-recommends \
        ssh socat ovmf qemu-kvm qemu-utils cloud-image-utils \
    && apt-get autoclean -y \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY qemu-entrypoint.sh /qemu-entrypoint.sh

CMD ["/bin/bash", "/qemu-entrypoint.sh"]