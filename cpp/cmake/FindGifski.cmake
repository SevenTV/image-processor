# * Try to find Gifski Once done this will define
#
# GIFSKI_FOUND - system has Gifski GIFSKI_INCLUDE_DIR - the Gifski include
# directory GIFSKI_LIBRARIES - Link these to use Gifski
#

find_path(
  GIFSKI_INCLUDE_DIR
  NAMES gifski.h
  PATHS ${_GIFSKI_INCLUDEDIR}
)

find_library(
  GIFSKI_LIBRARY
  NAMES gifski
  PATHS ${_GIFSKI_LIBDIR}
)

set(GIFSKI_LIBRARIES ${GIFSKI_LIBRARIES} ${GIFSKI_LIBRARY} ${_GIFSKI_LDFLAGS})

include(FindPackageHandleStandardArgs)
find_package_handle_standard_args(
  Gifski
  FOUND_VAR GIFSKI_FOUND
  REQUIRED_VARS GIFSKI_LIBRARY GIFSKI_LIBRARIES GIFSKI_INCLUDE_DIR
  VERSION_VAR _GIFSKI_VERSION)

# show the GIFSKI_INCLUDE_DIR, GIFSKI_LIBRARY and GIFSKI_LIBRARIES variables
# only in the advanced view
mark_as_advanced(GIFSKI_INCLUDE_DIR GIFSKI_LIBRARY GIFSKI_LIBRARIES)
