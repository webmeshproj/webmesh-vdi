FROM ubuntu:latest

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update \
    && apt-get -y -u dist-upgrade \
    && apt-get -y --no-install-recommends install \
        tigervnc-standalone-server xpra lwm supervisor pulseaudio \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && useradd -M -d /var/run/kvdi -u 9000 kvdi \
    && mkdir -p /var/log/supervisor \
    && chown -R kvdi: /var/log/supervisor \
    && echo "load-module module-native-protocol-unix auth-anonymous=1 socket=/var/run/kvdi/pulse-server" >> /etc/pulse/default.pa 

COPY supervisor/ /etc/supervisor/conf.d/
CMD ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/supervisord.conf"]