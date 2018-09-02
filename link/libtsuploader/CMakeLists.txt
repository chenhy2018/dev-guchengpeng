cmake_minimum_required (VERSION 2.8)

project(tsuploader)

include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/curl-7.61.0/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/c-sdk")

if(CMAKE_TOOLCHAIN_FILE)
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/include/ffmpeg")
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

set(LIBFFMPEG avformat avcodec avutil)
if(NOT WITH_FFMPEG)
        message("WITHOUT_FFMPEG")
        set(LIBFFMPEG "")
        add_definitions(-DUSE_OWN_TSMUX)
else()
        include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/")
        link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavformat")
        link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavcodec")
        link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/ffmpeg-4.0/libavutil")
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

if(CMAKE_TOOLCHAIN_FILE)
add_executable(testupload
    ipc_test.c
    socket_logging.c
)
add_definitions(-D__ARM)
else()
add_executable(testupload
    test.c
)
endif()

if(APPLE)
	set(CMAKE_EXE_LINKER_FLAGS
    		"-framework AudioToolbox -framework VideoToolbox -framework CoreGraphics -framework QuartzCore -framework CoreFoundation -framework CoreMedia -framework Security")
	if(DISABLE_OPENSSL)
        	target_link_libraries(testupload tsuploader ${LIBFFMPEG} qiniu curl bz2 lzma iconv z)
        else()
        	target_link_libraries(testupload tsuploader ${LIBFFMPEG} qiniu curl crypto bz2 lzma iconv z)
        endif()
else()
    if(CMAKE_TOOLCHAIN_FILE)
        if(DISABLE_OPENSSL)
        	target_link_libraries(testupload tsuploader ${LIBFFMPEG} qiniu curl m pthread devsdk tools)
        else()
        	target_link_libraries(testupload tsuploader ${LIBFFMPEG} qiniu curl crypto m pthread devsdk tools)
        endif()
    else()
        if(DISABLE_OPENSSL)
        	target_link_libraries(testupload tsuploader ${LIBFFMPEG} qiniu curl m pthread)
        else()
        	target_link_libraries(testupload tsuploader avformat avcodec avutil qiniu curl crypto m pthread)
        endif()
    endif()
endif()

