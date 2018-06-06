include(ExternalProject)
ExternalProject_Add(stun
        SOURCE_DIR stun
        BUILD_COMMAND make
        CONFIGURE_COMMAND ""
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)

add_subdirectory(libmuxer)
add_subdirectory(libsip)
add_subdirectory(libice)
add_subdirectory(libmqtt)
