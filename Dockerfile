# The base image used to build all other images
ARG BASE_IMG=ubuntu:22.04
# The tag to use for golang image
ARG GOLANG_TAG=1.18.1

#
# builder for all C/C++ libs and applications
#
FROM $BASE_IMG as builder-cpp

WORKDIR /tmp/app

ENV PATH="/root/.cargo/bin:${PATH}"

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
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

# copy third party deps
COPY cpp/third-party third-party

# build 3rd party deps, we remove any .git folders because we dont want the 3rd party deps to know they are apart of a git repo.
RUN cd third-party && make clean && rm -rf **/.git && make

# copy c++ code now
COPY cpp .

# build our c++ apps
RUN make

#
# builder for all golang applications.
#
FROM golang:$GOLANG_TAG as builder-go

WORKDIR /tmp/images

# update the apt repo and install any deps we might need.
RUN apt-get update && \
    apt-get install -y \
        build-essential \
        make \
        git && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

COPY go .

ARG BUILDER
ARG VERSION

ENV IMAGES_BUILDER=${BUILDER}
ENV IMAGES_VERSION=${VERSION}

# download all the golang modules, we do this step here so we can cache the build layers.
RUN make deps

# build the golang app
RUN make

#
# final squashed image
#
FROM $BASE_IMG

WORKDIR /app

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
COPY --from=builder-cpp /usr/local /usr/local
COPY --from=builder-go /tmp/images/bin/images_processor .
