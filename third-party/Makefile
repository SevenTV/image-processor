.PHONY: all clean

all: _libavif _libwebp _gifski _opencv

_libavif:
	cd libavif/ext && \
	(stat dav1d || ./dav1d.cmd) && stat dav1d && \
	(stat rav1e || ./rav1e.cmd) && stat rav1e && \
	cd .. && \
	((stat build && cd build && ninja) || (mkdir -p build && cd build && cmake .. -G Ninja -DAVIF_CODEC_DAV1D=ON -DAVIF_CODEC_RAV1E=ON -DAVIF_LOCAL_DAV1D=ON -DAVIF_LOCAL_RAV1E=ON && ninja)) && \
	cd build && \
	(sudo ninja install || ninja install)

_libwebp:
	cd libwebp && \
	((stat build && cd build && ninja) || (mkdir -p build && cd build && cmake .. -G Ninja -DWEBP_LINK_STATIC=OFF -DBUILD_SHARED_LIBS=ON && ninja)) && \
	cd build && \
	(sudo ninja install || ninja install)

_gifski:
	cd gifski && \
	cargo build --release --lib && \
	(sudo cp target/release/libgifski.so /usr/local/lib || cp target/release/libgifski.so /usr/local/lib) && \
	(sudo cp gifski.h /usr/local/include || cp gifski.h /usr/local/include)

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
		-DBUILD_KOTLIN_EXTENSIONS=OFF \
		 && ninja) && \
	cd build && \
	(sudo ninja install || ninja install)

clean:
	rm -rf libavif/ext/dav1d libavif/ext/rav1e libavif/build libwebp/build gifski/target opencv/build
