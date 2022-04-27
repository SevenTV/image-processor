FROM ubuntu:22.04 as builder-cpp

WORKDIR /tmp/app

ENV PATH="/root/.cargo/bin:${PATH}"

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        ca-certificates \
        build-essential \
        curl \
        ninja-build \
        meson \
        git \
        nasm \
        openssl \
        pkg-config \
        libssl-dev \
        cmake \
        libpng-dev \
        zlib1g-dev \
        libx264-dev \
        libx265-dev \
        libvpx-dev \
        libopenjp2-7-dev \
        libssl-dev && \
    curl https://sh.rustup.rs -sSf | bash -s -- -y

COPY cpp cpp
COPY .git .git

RUN cd cpp && make external_clean && make external

RUN make -C cpp

RUN apt-get remove -y \
        build-essential \
        curl \
        ninja-build \
        meson \
        git \
        nasm \
        openssl \
        pkg-config \
        libssl-dev \
        cmake \
        libpng-dev \
        zlib1g-dev && \
    apt-get autoremove -y && \
    apt-get install -y libpng16-16 && \
    apt-get clean -y && \
    rustup self uninstall -y && \
    cp /tmp/app/cpp/out/* /usr/local/bin && \
    rm -rf /tmp/app && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

FROM golang:1.18.1 as builder-go

WORKDIR /tmp/images

COPY go .

ARG BUILDER
ARG VERSION

ENV IMAGES_BUILDER=${BUILDER}
ENV IMAGES_VERSION=${VERSION}

RUN apt-get update && \
    apt-get install -y \
        make \
        git && \
    apt-get clean && \
    make build && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

FROM ubuntu:22.04

WORKDIR /app

COPY --from=builder-cpp /usr/local /usr/local
COPY --from=builder-go /tmp/images/bin/images .

RUN apt-get update && \
    apt-get install -y \
        libpng16-16 \
        libvpx7 \
        libx264-163 \
        libx265-199 \
        libopenjp2-7 \
        openssl && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*
