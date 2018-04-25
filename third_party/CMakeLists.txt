include(ExternalProject)
ExternalProject_Add(pjsip
        SOURCE_DIR pjproject-2.7.2
        BUILD_COMMAND make dep COMMAND make
        CONFIGURE_COMMAND ./configure
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)

include(ExternalProject)
ExternalProject_Add(ffmpeg
        SOURCE_DIR ffmpeg-4.0
        CONFIGURE_COMMAND ./configure  --bindir=./bin --datadir=./data --docdir=./doc --libdir=./lib --shlibdir=./shlib --incdir=./include --disable-manpages --enable-gpl --enable-version3 --enable-shared --enable-static --enable-openssl --enable-nonfree --enable-avresample
        BUILD_COMMAND make
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)

include(ExternalProject)
ExternalProject_Add(librtmp
        SOURCE_DIR rtmpdump-2.4
        CONFIGURE_COMMAND ""
        BUILD_COMMAND make
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)
