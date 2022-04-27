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
        cmake \
        libssl-dev \
        libpng-dev \
        zlib1g-dev \
        libx264-dev \
        libx265-dev \
        libvpx-dev \
        libopenjp2-7-dev \
        libssl-dev && \
    curl https://sh.rustup.rs -sSf | bash -s -- -y

COPY cpp/third-party third-party

RUN cd third-party && make clean && make

COPY cpp .

RUN make

FROM golang:1.18.1 as builder-go

WORKDIR /tmp/images

COPY go .

ARG BUILDER
ARG VERSION

ENV IMAGES_BUILDER=${BUILDER}
ENV IMAGES_VERSION=${VERSION}

RUN apt-get update && \
    apt-get install -y \
        build-essential \
        make \
        git

RUN make deps

RUN make

FROM ubuntu:22.04 as ffmpeg

WORKDIR /app

COPY --from=builder-cpp /usr/local /usr/local

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

FROM ffmpeg

COPY --from=builder-go /tmp/images/bin/images_processor .
