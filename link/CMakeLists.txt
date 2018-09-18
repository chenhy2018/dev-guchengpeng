if ($ENV{WITH_P2P} STREQUAL "ON")
        if ($ENV{ARCH} STREQUAL "x86") 
        include(ExternalProject)
        ExternalProject_Add(stun
                SOURCE_DIR stun
                BUILD_COMMAND make
                CONFIGURE_COMMAND ""
                INSTALL_COMMAND ""
                BUILD_IN_SOURCE 1
        )
        endif()
add_subdirectory(libice)
add_subdirectory(librtp)
endif()

if ($ENV{ARCH} STREQUAL "x86")
        include(ExternalProject)
        ExternalProject_Add(vod
                SOURCE_DIR vod
                BUILD_COMMAND  make
                CONFIGURE_COMMAND ""
                INSTALL_COMMAND ""
                BUILD_IN_SOURCE 1
        )

        include(ExternalProject)
        ExternalProject_Add(fastfwd
                SOURCE_DIR fastfwd
                BUILD_COMMAND  make
                CONFIGURE_COMMAND ""
                INSTALL_COMMAND ""
                BUILD_IN_SOURCE 1
        )
endif()

if ($ENV{WITH_PJSIP} STREQUAL "ON")
        add_subdirectory(libsip)
        add_subdirectory(libmqtt)
        add_subdirectory(libsdk)
endif()

add_subdirectory(libtsuploader)
