project(convert_png)

find_package(Gifski REQUIRED)
find_package(WebP REQUIRED)
find_package(OpenCV REQUIRED)
find_package(libavif REQUIRED)

add_executable(convert_png convert_png.cpp)

target_include_directories(
  convert_png
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${GIFSKI_INCLUDE_DIR} ${WebP_INCLUDE_DIRS} ${OPENCV_INCLUDE_DIRS})

target_link_libraries(convert_png ${GIFSKI_LIBRARIES} ${WebP_LIBRARIES} avif
                      ${OpenCV_LIBS})

install(TARGETS convert_png)
