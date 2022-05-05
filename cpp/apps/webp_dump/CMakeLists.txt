project(webp_dump)

list(APPEND CMAKE_MODULE_PATH ${CMAKE_SOURCE_DIR}/cmake)

find_package(OpenCV REQUIRED)
find_package(WebP REQUIRED)

add_executable(webp_dump webp_dump.cpp)

target_include_directories(
  webp_dump
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${WEBP_INCLUDE_DIR} ${OPENCV_INCLUDE_DIRS})

target_link_libraries(webp_dump ${WEBP_LIBRARIES} ${OpenCV_LIBS})

install(TARGETS webp_dump)
