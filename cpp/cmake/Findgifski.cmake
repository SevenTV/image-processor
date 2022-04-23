# * Try to find gifski Once done this will define
#
# GIFSKI_FOUND - system has gifski GIFSKI_INCLUDE_DIR - the gifski include
# directory GIFSKI_LIBRARIES - Link these to use gifski
#

find_path(
  GIFSKI_INCLUDE_DIR
  NAMES gifski.h
  PATHS ${_GIFSKI_INCLUDEDIR})

find_library(
  GIFSKI_LIBRARY
  NAMES gifski
  PATHS ${_GIFSKI_LIBDIR})

set(GIFSKI_LIBRARIES ${GIFSKI_LIBRARIES} ${GIFSKI_LIBRARY} ${_GIFSKI_LDFLAGS})

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(
  gifski
  FOUND_VAR GIFSKI_FOUND
  REQUIRED_VARS GIFSKI_LIBRARY GIFSKI_LIBRARIES GIFSKI_INCLUDE_DIR
  VERSION_VAR _GIFSKI_VERSION)

# show the AOM_INCLUDE_DIR, AOM_LIBRARY and AOM_LIBRARIES variables only in the
# advanced view
mark_as_advanced(GIFSKI_INCLUDE_DIR GIFSKI_LIBRARY GIFSKI_LIBRARIES)
