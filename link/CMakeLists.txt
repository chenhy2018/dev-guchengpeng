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

add_subdirectory(libsip)
add_subdirectory(libmqtt)
add_subdirectory(libsdk)
