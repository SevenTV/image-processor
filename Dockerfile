# The base image used to build all other images
ARG BASE_IMG=ubuntu:22.04
# The tag to use for golang image
ARG GOLANG_TAG=1.18.1

#
# Download and install all deps required to build/run the application
#
FROM $BASE_IMG as deps-builder
    WORKDIR /tmp/build

    ENV PATH="/root/.cargo/bin:${PATH}"

    COPY cpp/third-party third-party

    # Install all libs with dev files, including rust.
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
        curl https://sh.rustup.rs -sSf | bash -s -- -y && \
        cd third-party && \
        make clean && \
        rm -rf **/.git && \
        make && \
        apt-get remove -y ca-certificates \
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
        rustup self uninstall -y && \
        apt-get autoremove -y && \
        apt-get clean -y && \
        rm -rf /var/cache/apt/archives /var/lib/apt/lists/* && \
        cd / && \
        rm -rf /tmp/build

#
# Squash the deps-builder image into a single layer
#
FROM $BASE_IMG as deps
    WORKDIR /tmp/build

    # install all the final libs for ffmpeg.
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

    # copy compiled libs/binaries from the builders.
    COPY --from=deps-builder /usr/local /usr/local

    # Required since we are moving libs.
    RUN ldconfig

#
# CPP Source Code
#
FROM $BASE_IMG as cpp-src
    WORKDIR /tmp/src

    COPY cpp .

    RUN rm -rf third-party

#
# Build the cpp application
#
FROM deps as cpp-builder
    WORKDIR /tmp/build

    COPY --from=cpp-src /tmp/src .

    RUN apt-get update && \
        apt-get install -y \
            build-essential \
            cmake \
            make \
            ninja-build && \
        make && \
        apt-get remove -y \
            build-essential \
            cmake \
            make \
            ninja-build && \
        apt-get autoremove -y && \
        apt-get clean -y && \
        rm -rf /var/cache/apt/archives /var/lib/apt/lists/* && \
        mv out /tmp && \
        cd /tmp && \
        rm -rf /tmp/build && \
        mkdir /tmp/build && \
        mv out /tmp/build

#
# Download and install all deps required to run tests and build the go application
#
FROM golang:$GOLANG_TAG as go-deps
    WORKDIR /tmp/build

    COPY go/go.mod .
    COPY go/go.sum .
    COPY go/Makefile .

    # update the apt repo and install any deps we might need.
    RUN apt-get update && \
        apt-get install -y \
            make \
            git && \
        make deps && \
        apt-get autoremove -y && \
        apt-get clean -y && \
        rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

#
# Build the go application
#
FROM go-deps as go-builder
    WORKDIR /tmp/build

    ARG BUILDER
    ARG VERSION

    ENV IMAGES_BUILDER=${BUILDER}
    ENV IMAGES_VERSION=${VERSION}

    COPY go .

    RUN make

#
# Run the go tests
#
FROM go-deps as tests
    WORKDIR /tmp/build
    
    COPY assets /tmp/assets
    COPY --from=cpp-builder /tmp/build/out /usr/local/bin

    COPY go .

    RUN make test

    CMD ["make", "test"]

#
# final squashed image
#
FROM deps as final
    WORKDIR /app

    COPY --from=cpp-builder /tmp/build/out /usr/local/bin
    COPY --from=go-builder /tmp/build/out .
