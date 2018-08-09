cmake_minimum_required (VERSION 2.8)

project(tsuploader)

include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/curl-7.61.0/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/c-sdk")

link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavformat")
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavcodec")
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavutil")
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/curl-7.61.0/lib/.libs")

set(DISABLE_OPENSSL OFF)
if(NOT WITH_OPENSSL)
        message("WITHOUT_OPENSSL")
        set(DISABLE_OPENSSL ON)
        add_definitions(-DDISABLE_OPENSSL)
endif()

add_subdirectory(c-sdk)

add_library(tsuploader STATIC
	resource.c
	resource.h
	uploader.h
	uploader.c
	base.h
	adts.h
	adts.c
	queue.c
	queue.h
	log.h
	log.c
	servertime.h
	servertime.c
	tsmuxuploader.c
	tsmuxuploader.h
	tsuploaderapi.c
	tsuploaderapi.h
)

add_executable(testupload
        test.c
)

if(APPLE)
	set(CMAKE_EXE_LINKER_FLAGS
    		"-framework AudioToolbox -framework VideoToolbox -framework CoreGraphics -framework QuartzCore -framework CoreFoundation -framework CoreMedia -framework Security")
	if(DISABLE_OPENSSL)
        	target_link_libraries(testupload tsuploader avformat avcodec avutil qiniu curl bz2 lzma iconv z)
        else()
        	target_link_libraries(testupload tsuploader avformat avcodec avutil qiniu curl crypto bz2 lzma iconv z)
        endif()
else()
        if(DISABLE_OPENSSL)
        	target_link_libraries(testupload tsuploader avformat avcodec avutil qiniu curl m pthread)
        else()
        	target_link_libraries(testupload tsuploader avformat avcodec avutil qiniu curl crypto m pthread)
        endif()
endif()

