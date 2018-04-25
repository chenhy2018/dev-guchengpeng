include(ExternalProject)
ExternalProject_Add(pjsip
        SOURCE_DIR pjproject-2.7.2
        BUILD_COMMAND make dep COMMAND make
        CONFIGURE_COMMAND ./configure
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)

include(ExternalProject)
ExternalProject_Add(x264
        SOURCE_DIR x264
        BUILD_COMMAND make
        CONFIGURE_COMMAND ./configure --enable-shared --enable-static --disable-opencl --libdir=./prefix --includedir=./prefix --prefix=./prefix
        INSTALL_COMMAND make install
        BUILD_IN_SOURCE 1
)

SET(FDKAAC_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/fdkaac-0.1.6")
include(ExternalProject)
ExternalProject_Add(fdkaac
        SOURCE_DIR fdkaac-0.1.6
        CONFIGURE_COMMAND ./autogen.sh COMMAND ./configure --program-prefix=${FDKAAC_PREFIX_PATH}/prefix --libdir=${FDKAAC_PREFIX_PATH}/prefix --includedir=${FDKAAC_PREFIX_PATH}/prefix --bindir=${FDKAAC_PREFIX_PATH}/prefix
        BUILD_COMMAND make
        INSTALL_COMMAND make install
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
