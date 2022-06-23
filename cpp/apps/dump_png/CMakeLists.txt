project(webp_dump)

find_package(OpenCV REQUIRED)
find_package(WebP REQUIRED webp webpdemux)
find_package(libavif REQUIRED)

add_executable(dump_png dump_png.cpp)

target_include_directories(
  dump_png
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${WebP_INCLUDE_DIRS} ${OPENCV_INCLUDE_DIRS})

target_link_libraries(dump_png ${WebP_LIBRARIES} ${OpenCV_LIBS} avif)

install(TARGETS dump_png)
