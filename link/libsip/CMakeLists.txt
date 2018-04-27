SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -std=c++14 -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/pjproject-2.7.2/pjsip/include")
SET(CMAKE_CXX_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_SIP_SRC)

add_library (sip ${LIB_SIP_SRC})
