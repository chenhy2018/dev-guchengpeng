SET(GCC_COVERAGE_COMPILE_FLAGS "-std=gnu99 -Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/pjproject-2.7.2/pjlib/include -I../../third_party/pjproject-2.7.2/pjlib-util/include -I../../third_party/pjproject-2.7.2/pjnath/include -I../../third_party/pjproject-2.7.2/pjmedia/include -I../../third_party/pjproject-2.7.2/pjsip/include/")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

set(LIB_SRC
	heap.c
	JitterBuffer.c
	MediaStream.c
	PeerConnection.c)
set(LIB_HEADER
	MediaStream.h
	JitterBuffer.h
	qrtc.h
	heap.h
	PeerConnection.h)
set(TEST_SRC
	rtptest.c)

link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/pjproject-2.7.2/prefix/lib)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/openssl-1.1.0h/prefix/lib)

add_library (rtp-$ENV{LIBPREFIX} ${LIB_HEADER} ${LIB_SRC})
add_executable(rtptest-$ENV{LIBPREFIX} ${TEST_SRC})

target_link_libraries(rtptest-$ENV{LIBPREFIX} rtp-$ENV{LIBPREFIX} pjmedia-audiodev-$ENV{LIBPREFIX} pjnath-$ENV{LIBPREFIX} pjsip-ua-$ENV{LIBPREFIX}.a pjsip-$ENV{LIBPREFIX} pjmedia-$ENV{LIBPREFIX} pjlib-util-$ENV{LIBPREFIX} pj-$ENV{LIBPREFIX} pjmedia-$ENV{LIBPREFIX} pjmedia-codec-$ENV{LIBPREFIX}  srtp-$ENV{LIBPREFIX}  ssl crypto pthread m)
#target_link_libraries(rtptest rtp pjmedia-audiodev-x86_64-unknown-linux-gnu pjmedia-codec-x86_64-unknown-linux-gnu pjmedia-x86_64-unknown-linux-gnu speex-x86_64-unknown-linux-gnu ilbccodec-x86_64-unknown-linux-gnu gsmcodec-x86_64-unknown-linux-gnu srtp-x86_64-unknown-linux-gnu pjsdp-x86_64-unknown-linux-gnu pjnath-x86_64-unknown-linux-gnu pjsip-x86_64-unknown-linux-gnu pjlib-util-x86_64-unknown-linux-gnu pj-x86_64-unknown-linux-gnu asound ssl crypto pthread m)
