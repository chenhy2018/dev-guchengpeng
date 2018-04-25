SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -std=c++14 -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/ffmpeg-4.0 -I../../third_party/rtmpdump-2.3/librtmp")
SET(CMAKE_CXX_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_MUXER_SRC)

add_library (libmuxer.a ${LIB_MUXER_SRC})
