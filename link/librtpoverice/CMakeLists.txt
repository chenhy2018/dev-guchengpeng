SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/pjproject-2.7.2/pjlib/include -I../../third_party/pjproject-2.7.2/pjlib-util/include -I../../third_party/pjproject-2.7.2/pjnath/include -I../../third_party/pjproject-2.7.2/pjmedia/include -I../../third_party/pjproject-2.7.2/pjsip/include/")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_RTP_OVER_ICE_SRC)

link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/pjproject-2.7.2/pjnath/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/pjproject-2.7.2/pjmedia/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/pjproject-2.7.2/pjlib-util/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/pjproject-2.7.2/pjlib/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/pjproject-2.7.2/pjsip/lib/)
link_directories(${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/pjproject-2.7.2/third_party/lib/)

#add_library (rtpoverice ${LIB_RTP_OVER_ICE_SRC})
add_executable(icetest ${LIB_RTP_OVER_ICE_SRC})
target_link_libraries(icetest ice pjmedia-audiodev-x86_64-unknown-linux-gnu pjmedia-codec-x86_64-unknown-linux-gnu pjmedia-x86_64-unknown-linux-gnu speex-x86_64-unknown-linux-gnu ilbccodec-x86_64-unknown-linux-gnu gsmcodec-x86_64-unknown-linux-gnu srtp-x86_64-unknown-linux-gnu pjsdp-x86_64-unknown-linux-gnu pjnath-x86_64-unknown-linux-gnu pjsip-x86_64-unknown-linux-gnu pjlib-util-x86_64-unknown-linux-gnu pj-x86_64-unknown-linux-gnu asound ssl crypto pthread m)
