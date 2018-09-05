cmake_minimum_required(VERSION 3.5)
SET(GCC_COVERAGE_COMPILE_FLAGS "-std=c99 -Wall -Wno-unused -g -Wno-dev")
SET(GCC_INCLUDE_PATH "-I../../../ -I../../../../../third_party/pjproject-2.7.2/pjlib/include -I../../../../../third_party/pjproject-2.7.2/pjlib-util/include -I../../../../../third_party/pjproject-2.7.2/pjsip/include -I../../../../../third_party/pjproject-2.7.2/pjnath/include -I../../../../../third_party/pjproject-2.7.2/pjmedia/include")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../../../third_party/pjproject-2.7.2/prefix/$ENV{ARCH}/lib)

aux_source_directory(. LIB_SIP_TEST)
add_executable(max_ongoing_uac-$ENV{LIBPREFIX} ${LIB_SIP_TEST})

target_link_libraries(max_ongoing_uac-$ENV{LIBPREFIX} sip-$ENV{LIBPREFIX} pjsip-ua-$ENV{LIBPREFIX}.a pjsip-$ENV{LIBPREFIX} pjmedia-$ENV{LIBPREFIX} pjlib-util-$ENV{LIBPREFIX} pj-$ENV{LIBPREFIX} pthread m)
