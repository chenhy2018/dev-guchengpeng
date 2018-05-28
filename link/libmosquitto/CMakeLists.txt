SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/mosquitto-1.4.8/lib/ -I../../third_party/mosquitto-1.4.8/")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_MOSQUITTO_SRC)
aux_source_directory(../../third_party/mosquitto-1.4.8/lib/ MOSQUITTO_SRC)

add_library (emqtt ${LIB_MOSQUITTO_SRC} ${MOSQUITTO_SRC})
target_link_libraries(emqtt)

add_subdirectory(test)
