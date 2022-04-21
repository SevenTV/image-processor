FROM ubuntu:21.04 as builder

WORKDIR /tmp/app

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
        zlib1g-dev && \
    curl https://sh.rustup.rs -sSf | bash -s -- -y

COPY . .

RUN PATH="/root/.cargo/bin:${PATH}" make external && \
    ldconfig

RUN make

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
    PATH="/root/.cargo/bin:${PATH}" rustup self uninstall -y && \
    mkdir -p /app && \
    cp /tmp/app/out/convert_png /usr/local/bin && \
    rm -rf /tmp/app && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

# this is a trick with docker builds to squash all the compile layers into 2 layers.
FROM ubuntu:21.04

COPY --from=builder /lib /lib
COPY --from=builder /usr /usr
