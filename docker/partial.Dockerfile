ARG BASE_IMG=ubuntu:20.04

FROM $BASE_IMG
WORKDIR /app

RUN apt-get update && \
    apt-get install -y \
        ca-certificates \
        libpng16-16 \
        libvpx6 \
        libx264-155 \
        libx265-179 \
        libopenjp2-7 \
        openssl \
        libssl1.1 \
        gifsicle \
        optipng && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

COPY cpp/out/lib /usr/local/lib
COPY cpp/out/bin /usr/local/bin
COPY go/out/image_processor image_processor

RUN ldconfig

CMD ./image_processor
