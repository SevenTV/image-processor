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
        cd .. && \
        cp -r out / && \
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

#
# CPP Source Code
#
FROM $BASE_IMG as cpp-src
    WORKDIR /tmp/src

    COPY cpp .

    RUN rm -rf third-party out
    COPY --from=deps-builder /out out

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
            make

#
# Download and install all deps required to run tests and build the go application
#
FROM golang:$GOLANG_TAG as go

FROM deps as go-builder
    WORKDIR /tmp/build

    # update the apt repo and install any deps we might need.
    RUN apt-get update && \
        apt-get install -y \
            build-essential \
            make \
            git && \
        apt-get autoremove -y && \
        apt-get clean -y && \
        rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

    ENV PATH /usr/local/go/bin:$PATH
    ENV GOPATH /go
    ENV PATH $GOPATH/bin:$PATH
    COPY --from=go /usr/local /usr/local
    COPY --from=go /go /go

    COPY go/go.mod .
    COPY go/go.sum .
    COPY go/Makefile .

    RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH" && \
        make deps

    COPY go .

    COPY assets /tmp/assets

    COPY --from=cpp-builder /tmp/build/out /usr/local

    RUN ldconfig && make test

    ARG BUILDER
    ARG VERSION

    ENV IMAGES_BUILDER=${BUILDER}
    ENV IMAGES_VERSION=${VERSION}

    RUN make

#
# final squashed image
#
FROM deps as final
    WORKDIR /app

    RUN apt-get update && \
        apt-get install -y \
            ca-certificates && \
        apt-get autoremove -y && \
        apt-get clean -y && \
        rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

    COPY --from=cpp-builder /tmp/build/out /usr/local
    COPY --from=go-builder /tmp/build/out .

    RUN ldconfig

    CMD ./image_processor
