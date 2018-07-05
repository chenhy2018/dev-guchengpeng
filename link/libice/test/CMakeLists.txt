SET(GCC_COVERAGE_COMPILE_FLAGS "-std=gnu99 -Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../ -I../../../third_party/pjproject-2.7.2/pjlib/include -I../../../third_party/pjproject-2.7.2/pjlib-util/include -I../../../third_party/pjproject-2.7.2/pjnath/include -I../../../third_party/pjproject-2.7.2/pjmedia/include")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../third_party/pjproject-2.7.2/prefix/lib)

aux_source_directory(. LIB_ICE_TEST)
add_executable(ice_test-$ENV{LIBPREFIX} ${LIB_ICE_TEST})

target_link_libraries(ice_test-$ENV{LIBPREFIX} ice-$ENV{LIBPREFIX} pjnath-$ENV{LIBPREFIX} pjsip-ua-$ENV{LIBPREFIX}.a pjsip-$ENV{LIBPREFIX} pjmedia-$ENV{LIBPREFIX} pjlib-util-$ENV{LIBPREFIX} pj-$ENV{LIBPREFIX}  pthread m)
