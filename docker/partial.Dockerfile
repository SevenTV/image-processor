ARG BASE_IMG=ubuntu:22.04

FROM $BASE_IMG
WORKDIR /app

# RUN apt-get update && \
#     apt-get install -y \
#         ca-certificates \
#         libpng16-16 \
#         libvpx7 \
#         libx264-163 \
#         libx265-199 \
#         libopenjp2-7 \
#         openssl \
#         libssl3 \
#         gifsicle \
#         optipng \
#         libasound2 \
#         libxcb-xfixes0 \
#         libxcb-shape0 \
#         libxcb-shm0 \
#         libsdl2-2.0-0 \
#         libsndio7.0 \
#         libxv1 \
#         libva2 \
#         libva-drm2 \
#         libva-x11-2 \
#         libvdpau1 && \
#     apt-get autoremove -y && \
#     apt-get clean -y && \
#     rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

RUN apt-get update && \
    apt-get install -y \
        ca-certificates \
        libpng16-16 \
        libvpx7 \
        libx264-163 \
        libx265-199 \
        libopenjp2-7 \
        openssl \
        libssl3 \
        gifsicle \
        optipng && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

COPY cpp/out/lib /usr/local/lib
COPY cpp/out/bin /usr/local/bin
COPY go/out image_processor

RUN ldconfig

CMD ./image_processor
