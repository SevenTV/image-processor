project(7tv_apps)

option(CONVERT_PNG "Build the convert_png binary" ON)
if(CONVERT_PNG)
  add_subdirectory(convert_png)
endif(CONVERT_PNG)

option(RESIZE_PNG "Build the resize_png binary" ON)
if(RESIZE_PNG)
  add_subdirectory(resize_png)
endif(RESIZE_PNG)

option(DUMP_PNG "Build the dump_png binary" ON)
if(DUMP_PNG)
  add_subdirectory(dump_png)
endif(DUMP_PNG)
