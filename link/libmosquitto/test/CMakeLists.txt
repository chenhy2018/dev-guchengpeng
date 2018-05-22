SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../ ")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../../third_party/mosquitto-1.4.8/lib/)
aux_source_directory(. LIB_MOSQUITTO_TEST)
add_executable(mos_test ${LIB_MOSQUITTO_TEST})

target_link_libraries(mos_test mos mosquitto pthread)
