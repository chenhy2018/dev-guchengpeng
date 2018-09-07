cmake_minimum_required (VERSION 2.8)

project(testupload1)

include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../")


link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../../third_party/curl-7.61.0/lib/.libs")

if(CMAKE_TOOLCHAIN_FILE)
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../demo/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../")
endif()

set(LIBFFMPEG avformat avcodec swresample avutil)

if ($ENV{WITH_FFMPEG} STREQUAL "OFF")
        message("WITHOUT_FFMPEG")
        add_definitions(-DUSE_OWN_TSMUX)
else()
        message("WITH_FFMPEG")
endif()

if (NOT TEST_WITHOUT_FFMPEG) 
        message("TEST_WITH_FFMPEG")
        add_definitions(-DTEST_WITH_FFMPEG)
	#include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/")
	#link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavformat")
	#link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavcodec")
	#link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavutil")
else()
        message("TEST_WITHOUT_FFMPEG")
        set(LIBFFMPEG "")
endif()

add_executable(testupload1
    test.c
    flag.h
    flag.c
)

if(APPLE)
	set(CMAKE_EXE_LINKER_FLAGS
    		"-framework AudioToolbox -framework VideoToolbox -framework CoreGraphics -framework QuartzCore -framework CoreFoundation -framework CoreMedia -framework Security")
	if(DISABLE_OPENSSL)
        	target_link_libraries(testupload1 tsuploader ${LIBFFMPEG} qiniu curl bz2 lzma iconv z)
        else()
        	target_link_libraries(testupload1 tsuploader ${LIBFFMPEG} qiniu curl crypto bz2 lzma iconv z)
        endif()
else()
    if(CMAKE_TOOLCHAIN_FILE)
        if(DISABLE_OPENSSL)
        	target_link_libraries(testupload1 tsuploader ${LIBFFMPEG} qiniu curl m pthread devsdk tools)
        else()
        	target_link_libraries(testupload1 tsuploader ${LIBFFMPEG} qiniu curl crypto m pthread devsdk tools)
        endif()
    else()
        if(DISABLE_OPENSSL)
        	target_link_libraries(testupload1 tsuploader ${LIBFFMPEG} qiniu curl m pthread z)
        else()
        	target_link_libraries(testupload1 tsuploader ${LIBFFMPEG} avutil qiniu curl crypto m pthread z)
        endif()
    endif()
endif()

