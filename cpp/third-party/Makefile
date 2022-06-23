.PHONY: all clean

all: _libavif _libwebp _gifski _opencv _ffmpeg

_build:
	mkdir -p ../out/lib
	mkdir -p ../out/include
	mkdir -p ../out/bin
	mkdir -p build

_dav1d: _build
	cd build && \
	mkdir -p dav1d && \
	cd dav1d && \
	meson \
		--buildtype release \
		--default-library=shared \
		--prefix $$(realpath $$(pwd)/../../../out) \
		--includedir $$(realpath $$(pwd)/../../../out/include) \
		--libdir $$(realpath $$(pwd)/../../../out/lib) \
		--bindir $$(realpath $$(pwd)/../../../out/bin) \
		../../dav1d && \
	ninja && \
	ninja install

_rav1e: _build
	cd rav1e && \
	cargo install cargo-c && \
	CARGO_TARGET_DIR=$$(realpath $$(pwd)/../build/rav1e) cargo cinstall --release --prefix=$$(realpath $$(pwd)/../../out) --library-type=cdylib

_libavif: _build _dav1d _rav1e
	cd build && \
	mkdir -p libavif && \
	cd libavif && \
	cmake ../../libavif \
			-G Ninja \
			-DAVIF_CODEC_DAV1D=ON \
			-DAVIF_CODEC_RAV1E=ON \
			-DCMAKE_PREFIX_PATH=$$(realpath $$(pwd)/../../../out) \
			-D_DAV1D_INCLUDEDIR=$$(realpath $$(pwd)/../../../out/include) \
			-D_DAV1D_LIBDIR=$$(realpath $$(pwd)/../../../../out/lib) \
			-D_RAV1E_INCLUDEDIR=$$(realpath $$(pwd)/../../../../out/include) \
			-D_RAV1E_LIBDIR=$$(realpath $$(pwd)/../../../../out/lib) \
			-DCMAKE_INSTALL_BINDIR=$$(realpath $$(pwd)/../../../out/bin) \
			-DCMAKE_INSTALL_LIBDIR=$$(realpath $$(pwd)/../../../out/lib) \
			-DCMAKE_INSTALL_INCLUDEDIR=$$(realpath $$(pwd)/../../../out/include) && \
	ninja && \
	ninja install

_libwebp: _build
	cd build && \
	mkdir -p libwebp && \
	cd libwebp && \
	cmake ../../libwebp \
		-G Ninja \
		-DBUILD_SHARED_LIBS=ON \
		-DCMAKE_INSTALL_PREFIX=$$(realpath $$(pwd)/../../../out) && \
	ninja && \
	ninja install

_gifski: _build
	cd gifski && \
	CARGO_TARGET_DIR=$$(realpath $$(pwd)/../build/gifski) cargo build --release --lib && \
	cp $$(realpath $$(pwd)/../build/gifski/release/libgifski.so) $$(realpath $$(pwd)/../../out/lib) && \
	cp gifski.h $$(realpath $$(pwd)/../../out/include)

_opencv: _build
	cd build && \
	mkdir -p opencv && \
	cd opencv && \
	cmake ../../opencv -G Ninja \
		-DCMAKE_INSTALL_PREFIX=$$(realpath $$(pwd)/../../../out) \
		-DWITH_AVFOUNDATION=OFF \
		-DWITH_CAP_IOS=OFF \
		-DWITH_CAROTENE=OFF \
		-DWITH_VTK=OFF \
		-DWITH_CUDA=OFF \
		-DWITH_CUFFT=OFF \
		-DWITH_CUBLAS=OFF \
		-DWITH_CUDNN=OFF \
		-DWITH_NVCUVID=OFF \
		-DWITH_EIGEN=OFF \
		-DWITH_FFMPEG=OFF \
		-DWITH_GSTREAMER=OFF \
		-DWITH_GTK=OFF \
		-DWITH_GTK_2_X=OFF \
		-DWITH_IPP=OFF \
		-DWITH_HALIDE=OFF \
		-DWITH_VULKAN=OFF \
		-DWITH_OPENVINO=OFF \
		-DWITH_WEBNN=OFF \
		-DWITH_JASPER=OFF \
		-DWITH_OPENJPEG=OFF \
		-DWITH_JPEG=OFF \
		-DWITH_WEBP=OFF \
		-DWITH_OPENEXR=OFF \
		-DWITH_OPENGL=OFF \
		-DWITH_OPENVX=OFF \
		-DWITH_OPENNI=OFF \
		-DWITH_OPENNI2=OFF \
		-DWITH_PNG=ON \
		-DWITH_GDCM=OFF \
		-DWITH_PVAPI=OFF \
		-DWITH_ARAVIS=OFF \
		-DWITH_QT=OFF \
		-DWITH_WIN32UI=OFF \
		-DWITH_TBB=OFF \
		-DWITH_HPX=OFF \
		-DWITH_OPENMP=OFF \
		-DWITH_PTHREADS_PF=OFF \
		-DWITH_TIFF=OFF \
		-DWITH_V4L=OFF \
		-DWITH_DSHOW=OFF \
		-DWITH_MSMF=OFF \
		-DWITH_MSMF_DXVA=OFF \
		-DWITH_XIMEA=OFF \
		-DWITH_UEYE=OFF \
		-DWITH_XINE=OFF \
		-DWITH_CLP=OFF \
		-DWITH_OPENCL=OFF \
		-DWITH_OPENCL_SVM=OFF \
		-DWITH_OPENCLAMDFFT=OFF \
		-DWITH_OPENCLAMDBLAS=OFF \
		-DWITH_DIRECTX=OFF \
		-DWITH_OPENCL_D3D11_NV=OFF \
		-DWITH_LIBREALSENSE=OFF \
		-DWITH_VA=OFF \
		-DWITH_VA_INTEL=OFF \
		-DWITH_MFX=OFF \
		-DWITH_GDAL=OFF \
		-DWITH_GPHOTO2=OFF \
		-DWITH_LAPACK=OFF \
		-DWITH_ITT=OFF \
		-DWITH_PROTOBUF=OFF \
		-DWITH_IMGCODEC_HDR=OFF \
		-DWITH_IMGCODEC_SUNRASTER=OFF \
		-DWITH_IMGCODEC_PXM=OFF \
		-DWITH_IMGCODEC_PFM=OFF \
		-DWITH_QUIRC=OFF \
		-DWITH_ANDROID_MEDIANDK=OFF \
		-DWITH_ANDROID_NATIVE_CAMERA=OFF \
		-DWITH_TENGINE=OFF \
		-DWITH_ONNX=OFF \
		-DWITH_TIMVX=OFF \
		-DBUILD_SHARED_LIBS=ON \
		-DBUILD_opencv_apps=OFF \
		-DBUILD_opencv_js=OFF \
		-DBUILD_ANDROID_PROJECTS=OFF \
		-DBUILD_ANDROID_EXAMPLES=OFF \
		-DBUILD_DOCS=OFF \
		-DBUILD_EXAMPLES=OFF \
		-DBUILD_PACKAGE=OFF \
		-DBUILD_PERF_TESTS=OFF \
		-DBUILD_TESTS=OFF \
		-DBUILD_WITH_DEBUG_INFO=OFF \
		-DBUILD_WITH_STATIC_CRT=OFF \
		-DUILD_WITH_DYNAMIC_IPP=OFF \
		-DBUILD_FAT_JAVA_LIB=OFF \
		-DBUILD_ANDROID_SERVICE=OFF \
		-DBUILD_CUDA_STUBS=OFF \
		-DBUILD_JAVA=OFF \
		-DBUILD_OBJC=OFF \
		-DBUILD_KOTLIN_EXTENSIONS=OFF && \
	ninja && \
	ninja install

_ffmpeg: _build
	cd build && \
	mkdir -p FFmpeg && \
	cd FFmpeg && \
	PKG_CONFIG_PATH=$$(realpath $$(pwd)/../../../out/lib/pkgconfig) ../../FFmpeg/configure \
		--extra-libs="-lpthread -lm" \
		--ld="g++" \
		--prefix=$$(realpath $$(pwd)/../../../out) \
		--toolchain=hardened \
		--arch=amd64 \
		--enable-gpl \
		--disable-stripping \
		--enable-libdav1d \
		--enable-nonfree \
		--enable-librav1e \
		--enable-openssl \
		--enable-libopenjpeg \
		--enable-libvpx \
		--enable-libwebp \
		--enable-libx265 \
		--enable-libx264 \
		--enable-shared \
		--disable-static && \
	make -j$$(nproc) && \
	make install

clean: _build
	rm -rf build/*
