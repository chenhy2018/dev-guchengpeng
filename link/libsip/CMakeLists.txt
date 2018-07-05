SET(GCC_COVERAGE_COMPILE_FLAGS "-std=gnu99 -g")
SET(GCC_INCLUDE_PATH "-I../util/ -I../../third_party/pjproject-2.7.2/pjlib/include -I../../third_party/pjproject-2.7.2/pjlib-util/include -I../../third_party/pjproject-2.7.2/pjsip/include -I../../third_party/pjproject-2.7.2/pjmedia/include -I../../third_party/pjproject-2.7.2/pjnath/include")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_SIP_SRC)
aux_source_directory(../util/ UTIL)

add_library (sip--$ENV{LIBPREFIX} ${LIB_SIP_SRC} ${UTIL})
add_subdirectory(test/)
