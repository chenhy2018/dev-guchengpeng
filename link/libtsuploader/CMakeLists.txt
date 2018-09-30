cmake_minimum_required (VERSION 2.8)

project(tsuploader)

#add_definitions("-Wall -g")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/curl-7.61.0/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/c-sdk")

include_directories("${CMAKE_CURRENT_SOURCE_DIR}/demo/include")
include_directories("${CMAKE_CURRENT_SOURCE_DIR}/")

if ( $ENV{ARCH} STREQUAL mstar )
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/libs/mstar")
elseif ( $ENV{ARCH} STREQUAL hi )
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/libs/hi")
set( MOCK_ANJIA_API ON)
add_definitions(-DMOCK_ANJIA )
elseif ( $ENV{ARCH} STREQUAL 80386 )
set( MOCK_ANJIA_API ON)
add_definitions(-DMOCK_ANJIA )
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/libs/80386")
else()
set( MOCK_ANJIA_API ON)
add_definitions(-DMOCK_ANJIA )
link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/curl-7.61.0/lib/.libs")
endif()

set(DISABLE_OPENSSL OFF)
if ($ENV{WITH_OPENSSL} STREQUAL "OFF")
        message("WITHOUT_OPENSSL")
        set(DISABLE_OPENSSL ON)
        add_definitions(-DDISABLE_OPENSSL)
else()
        set(OPENSSL_ROOT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/openssl-1.1.0h")
        link_directories("${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/openssl-1.1.0h")
        message("WITH_OPENSSL")
endif()

set(LIBFFMPEG "")

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
        localkey.h
)

if ( $ENV{ARCH} STREQUAL x86 )
        if ($ENV{WITH_FFMPEG} STREQUAL "ON")
            add_subdirectory(pcdemo)
        endif()
endif()

add_executable(ipc-testupload
    ./demo/ipc_test.c
    ./demo/socket_logging.c
    ./demo/dbg.c
    ./demo/log2file.c
    ./demo/queue.c
    ./demo/cfg_parse.c
    ./demo/mymalloc.c
    ./demo/mock_anjia.c
)

if ( $ENV{ARCH} STREQUAL mstar )
    add_definitions(-D__ARM)
endif()

if(DISABLE_OPENSSL)
    if(MOCK_ANJIA_API)
        message("mock an-jia")
        target_link_libraries(ipc-testupload tsuploader ${LIBFFMPEG} qiniu curl m pthread )
    else()
        target_link_libraries(ipc-testupload tsuploader ${LIBFFMPEG} qiniu curl m pthread devsdk tools)
    endif()
else()
    target_link_libraries(ipc-testupload tsuploader ${LIBFFMPEG} qiniu curl crypto m pthread devsdk tools)
endif()

