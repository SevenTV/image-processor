# C++

## 1. To build this repo, you must first install all required dependancies

Update your apt repo

```bash
sudo apt-get update
```

Install all these packages used for building.

```bash
sudo apt-get install \
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
    libssl-dev \
    gifsicle \
    optipng
```

You will also need the rust compiler

```bash
curl https://sh.rustup.rs -sSf | bash -s -- -y
```

## 2. You can now start builing all submodules.

To fetch them do

```bash
git submodule sync && git submodule update --init
```

Then run to build all the submodules, this will take a long time.

```bash
make external
```

## 3. Build the application

Now that everything else is installed you can simply run this to build the application.

```bash
make
```

The output folder should be created in `./out`

## 4. Clean up

If you wish to clean up you can run

```bash
make clean
```

this will clean the application build, if you want to clean all submodule build you must run

```bash
make external_clean
```

## 5. Formatting

The formatters we use are `clang-format` and `cmake-format`

### Installing Clang-Format

```bash
sudo apt-get install clang-format
```

### Installing CMake-Format (requires python3)

```bash
pip3 install cmake-format
```
