cmake_minimum_required(VERSION 3.13)

set(CMAKE_POSITION_INDEPENDENT_CODE ON)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_SOURCE_DIR}/out)

project(convert_png)

list(APPEND CMAKE_MODULE_PATH ${CMAKE_SOURCE_DIR}/cmake)
find_package(gifski REQUIRED)
find_package(WebP REQUIRED)
find_package(libavif REQUIRED)
find_package(OpenCV REQUIRED)

add_executable(convert_png convert_png.cpp)
target_include_directories(
  convert_png
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${GIFSKI_INCLUDE_DIR} ${WEBP_INCLUDE_DIR} ${OpenCV_INCLUDE_DIRS}
          ${LIBAVIF_INCLUDE_DIR})
target_link_libraries(convert_png ${GIFSKI_LIBRARIES} ${WEBP_LIBRARIES} avif
                      ${OpenCV_LIBS})
