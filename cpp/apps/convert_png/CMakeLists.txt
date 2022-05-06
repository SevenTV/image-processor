project(convert_png)

list(APPEND CMAKE_MODULE_PATH ${CMAKE_SOURCE_DIR}/cmake)

find_package(Gifski REQUIRED)
find_package(WebP REQUIRED webpmux webp)
find_package(OpenCV REQUIRED)
find_package(libavif REQUIRED)

add_executable(convert_png convert_png.cpp)

target_include_directories(
  convert_png
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${GIFSKI_INCLUDE_DIR} ${WEBP_INCLUDE_DIR} ${OPENCV_INCLUDE_DIRS})

target_link_libraries(convert_png ${GIFSKI_LIBRARIES} WebP::libwebpmux
                      WebP::webp avif ${OpenCV_LIBS})

install(TARGETS convert_png)
