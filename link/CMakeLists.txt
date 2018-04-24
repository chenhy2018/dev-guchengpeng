include(ExternalProject)
ExternalProject_Add(stun
        SOURCE_DIR stun
        BUILD_COMMAND make
        CONFIGURE_COMMAND ""
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)

