project(resize_png)

find_package(OpenCV REQUIRED)

add_executable(resize_png resize_png.cpp)

target_include_directories(
  resize_png
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${OPENCV_INCLUDE_DIRS})

target_link_libraries(resize_png ${OpenCV_LIBS})

install(TARGETS resize_png)
