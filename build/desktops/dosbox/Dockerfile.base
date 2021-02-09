FROM ubuntu:latest

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update \
    && apt-get -y -u dist-upgrade \
    && apt-get -y --no-install-recommends install \
        dosbox tigervnc-standalone-server xpra xfonts-base lwm supervisor xdotool pulseaudio libsdl1.2debian \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN echo "Set up dosbox" \
    && mkdir -p /dos/drive_c \
    # Drive K will be the volume used internally by kvdi
    && mkdir -p /dos/drive_k \
    && mv `dosbox -printconf` /dos/dosbox.conf \
    && echo "mount c /dos/drive_c" >> /dos/dosbox.conf \
    # Drive K will be the volume used internally by kvdi
    && echo "mount k /dos/drive_k" >> /dos/dosbox.conf \
    && sed -i 's/usescancodes=true/usescancodes=false/' /dos/dosbox.conf \
    && sed -i 's/fullscreen=false/fullscreen=true/' /dos/dosbox.conf \
    && sed -i 's/fulldouble=false/fulldouble=true/' /dos/dosbox.conf \
    && sed -i 's/fullresolution=original/fullresolution=auto/' /dos/dosbox.conf \
    && sed -i 's/windowresolution=original/windowresolution=auto/' /dos/dosbox.conf \
    && sed -i 's/output=surface/output=opengl/' /dos/dosbox.conf \
    && sed -i 's/midiconfig=/midiconfig=128:0/' /dos/dosbox.conf \
    && echo "load-module module-native-protocol-unix auth-anonymous=1 socket=/var/run/kvdi/pulse-server" >> /etc/pulse/default.pa \
    && chmod 777 /var/log/supervisor \
    && find /dos -type d -exec chmod 777 {} \; \
    && find /dos -type f -exec chmod 666 {} \;

COPY supervisor/ /etc/supervisor/conf.d/
CMD ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/supervisord.conf"]