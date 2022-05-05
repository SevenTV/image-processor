.PHONY: all clean

all: _libavif _libwebp _gifski _opencv _ffmpeg

_dav1d:
	cd dav1d && \
	(stat build || (mkdir build && cd build && meson --buildtype release --default-library=shared --libdir /usr/local/lib --includedir /usr/local/include ..)) && \
	cd build && ninja && \
	(sudo ninja install || ninja install)

_rav1e:
	cd rav1e && \
	cargo install cargo-c && \
	cargo cinstall --release --prefix=/usr/local --destdir ./out --library-type=cdylib && \
	(sudo cp -r out/usr/local/* /usr/local || cp -r out/usr/local/* /usr/local) && \
	(sudo ldconfig || ldconfig)

_libavif: _dav1d _rav1e
	cd libavif && \
	((stat build && cd build && ninja) || (mkdir -p build && cd build && cmake .. \
		-G Ninja \
		-DAVIF_CODEC_DAV1D=ON \
		-DAVIF_CODEC_RAV1E=ON && \
		ninja)) && \
	cd build && \
	(sudo ninja install || ninja install)

_libwebp:
	cd libwebp && \
	((stat build && cd build && ninja) || (mkdir -p build && cd build && cmake .. \
		-G Ninja \
		-DBUILD_SHARED_LIBS=ON && \
	ninja)) && \
	cd build && \
	(sudo ninja install || ninja install)

_gifski:
	cd gifski && \
	cargo build --release --lib && \
	(sudo cp target/release/libgifski.so /usr/local/lib || cp target/release/libgifski.so /usr/local/lib) && \
	(sudo cp gifski.h /usr/local/include || cp gifski.h /usr/local/include) && \
	(sudo ldconfig || ldconfig)

_opencv:
	cd opencv && \
	((stat build && cd build && ninja) || mkdir -p build && cd build && cmake .. -G Ninja \
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
	ninja) && \
	cd build && \
	(sudo ninja install || ninja install)

_ffmpeg:
	cd FFmpeg && \
	stat build || (mkdir build && cd build && \
	../configure \
		--extra-libs="-lpthread -lm" \
		--ld="g++" \
		--prefix=/usr/local \
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
		--disable-static) && \
	cd build && \
	make -j$$(nproc) && \
	(sudo make install || make install)

clean:
	rm -rf dav1d/build libavif/build libwebp/build gifski/target opencv/build rav1e/target FFmpeg/build
