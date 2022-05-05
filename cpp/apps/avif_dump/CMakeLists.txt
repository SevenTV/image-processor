project(avif_dump)

list(APPEND CMAKE_MODULE_PATH ${CMAKE_SOURCE_DIR}/cmake)

find_package(OpenCV REQUIRED)
find_package(libavif REQUIRED)

add_executable(avif_dump avif_dump.cpp)

target_include_directories(
  avif_dump
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${OPENCV_INCLUDE_DIRS})

target_link_libraries(avif_dump avif ${OpenCV_LIBS})

install(TARGETS avif_dump)
