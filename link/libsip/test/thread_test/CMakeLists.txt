cmake_minimum_required(VERSION 3.5)
SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../../ -I../../../../third_party/pjproject-2.7.2/pjlib/include -I../../../../third_party/pjproject-2.7.2/pjlib-util/include -I../../../../third_party/pjproject-2.7.2/pjsip/include -I../../../../third_party/pjproject-2.7.2/pjnath/include -I../../../../third_party/pjproject-2.7.2/pjmedia/include")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")


link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../../third_party/pjproject-2.7.2/pjsip/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../../third_party/pjproject-2.7.2/pjmedia/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../../third_party/pjproject-2.7.2/pjlib-util/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../../third_party/pjproject-2.7.2/pjlib/lib/)

aux_source_directory(. LIB_SIP_TEST)
add_executable(thread_test ${LIB_SIP_TEST})

target_link_libraries(thread_test sip pjsip-ua-x86_64-unknown-linux-gnu.a pjsip-x86_64-unknown-linux-gnu pjmedia-x86_64-unknown-linux-gnu pjlib-util-x86_64-unknown-linux-gnu pj-x86_64-unknown-linux-gnu pthread m)
