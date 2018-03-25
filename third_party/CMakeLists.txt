include(ExternalProject)
ExternalProject_Add(pjsip
        SOURCE_DIR pjproject-2.7.2
        BUILD_COMMAND make dep COMMAND make
        CONFIGURE_COMMAND ./configure
        INSTALL_COMMAND make
        BUILD_IN_SOURCE 1 
)
