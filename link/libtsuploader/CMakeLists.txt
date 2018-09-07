cmake_minimum_required (VERSION 2.8)

project(tsuploader)

include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/curl-7.61.0/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/c-sdk")

if(CMAKE_TOOLCHAIN_FILE)
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/demo/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/")
endif()


link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/curl-7.61.0/lib/.libs")
if(CMAKE_TOOLCHAIN_FILE)
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/libs")
endif()

set(DISABLE_OPENSSL OFF)
if(NOT WITH_OPENSSL)
        message("WITHOUT_OPENSSL")
        set(DISABLE_OPENSSL ON)
        add_definitions(-DDISABLE_OPENSSL)
endif()

if(CMAKE_TOOLCHAIN_FILE)
set(LIBFFMPEG "")
else()
set(LIBFFMPEG avformat avcodec swresample avutil)
endif()

if ($ENV{WITH_FFMPEG} STREQUAL "OFF")
        message("WITHOUT_FFMPEG")
        add_definitions(-DUSE_OWN_TSMUX)
else()
        message("WITH_FFMPEG")
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
        tsmux.h
        tsmux.c
        mpegts.c
        mpegts.h
)

add_subdirectory(pcdemo)

if(CMAKE_TOOLCHAIN_FILE)
add_executable(testupload
    ./demo/ipc_test.c
    ./demo/socket_logging.c
    ./demo/dbg.c
    ./demo/log2file.c
)
add_definitions(-D__ARM)
endif()

if(CMAKE_TOOLCHAIN_FILE)
    if(DISABLE_OPENSSL)
    	target_link_libraries(testupload tsuploader ${LIBFFMPEG} qiniu curl m pthread devsdk tools)
    else()
    	target_link_libraries(testupload tsuploader ${LIBFFMPEG} qiniu curl crypto m pthread devsdk tools)
    endif()
endif()

