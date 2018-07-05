SET(GCC_INCLUDE_PATH "-I../util/ -I../libsip -I ../libmqtt -I ../librtp   -I../../third_party/pjproject-2.7.2/pjlib/include -I../../third_party/pjproject-2.7.2/pjlib-util/include -I../../third_party/pjproject-2.7.2/pjsip/include -I../../third_party/pjproject-2.7.2/pjmedia/include -I../../third_party/pjproject-2.7.2/pjnath/include")

SET(GCC_COVERAGE_COMPILE_FLAGS "-std=gnu99 -Wall -Wno-unused -g")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_SDK_SRC)

add_library (sdk-$ENV{LIBPREFIX} ${LIB_SDK_SRC})
