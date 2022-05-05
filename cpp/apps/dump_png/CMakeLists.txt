project(webp_dump)

list(APPEND CMAKE_MODULE_PATH ${CMAKE_SOURCE_DIR}/cmake)

find_package(OpenCV REQUIRED)
find_package(WebP REQUIRED)
find_package(libavif REQUIRED)

add_executable(dump_png dump_png.cpp)

target_include_directories(
  dump_png
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${WEBP_INCLUDE_DIR} ${OPENCV_INCLUDE_DIRS})

target_link_libraries(dump_png ${WEBP_LIBRARIES} ${OpenCV_LIBS} avif)

install(TARGETS dump_png)
