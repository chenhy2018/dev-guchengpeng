SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/pjproject-2.7.2/pjlib/include -I../../third_party/pjproject-2.7.2/pjlib-util/include -I../../third_party/pjproject-2.7.2/pjnath/include -I../../third_party/pjproject-2.7.2/pjmedia/include")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_ICE_SRC)

add_library (ice ${LIB_ICE_SRC})

add_subdirectory(test)
