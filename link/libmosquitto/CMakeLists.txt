SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/libmosquitto-1.4.8/lib/")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_MOSQUITTO_SRC)

add_library (mos ${LIB_MOSQUITTO_SRC})

add_subdirectory(test)
