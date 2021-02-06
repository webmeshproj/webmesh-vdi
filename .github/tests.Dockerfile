FROM ubuntu:20.10

RUN mkdir -p /build \
    && apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y \
        golang git make curl pkg-config build-essential \
        libpulse-dev libgstreamer1.0 libgstreamer1.0-dev \
        libgstreamer-plugins-bad1.0-dev libgstreamer-plugins-base1.0-dev
        # gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad \
        # gstreamer1.0-plugins-ugly gstreamer1.0-libav gstreamer1.0-tools