include(ExternalProject)
ExternalProject_Add(pjsip
        SOURCE_DIR pjproject-2.7.2
        BUILD_COMMAND make dep COMMAND make
        CONFIGURE_COMMAND ./configure
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)

SET(X264_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/x264/prefix")
include(ExternalProject)
ExternalProject_Add(x264
        SOURCE_DIR x264
        BUILD_COMMAND make
        CONFIGURE_COMMAND ./configure --enable-shared --enable-static --disable-opencl --libdir=${X264_PREFIX_PATH} --includedir=${X264_PREFIX_PATH} --prefix=${X264_PREFIX_PATH}
        INSTALL_COMMAND make install
        BUILD_IN_SOURCE 1
)

SET(FDKAAC_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/fdkaac-0.1.6/prefix")
include(ExternalProject)
ExternalProject_Add(fdkaac
        SOURCE_DIR fdkaac-0.1.6
        CONFIGURE_COMMAND ./autogen.sh COMMAND ./configure --prefix=${FDKAAC_PREFIX_PATH} --libdir=${FDKAAC_PREFIX_PATH} --includedir=${FDKAAC_PREFIX_PATH} --bindir=${FDKAAC_PREFIX_PATH}
        BUILD_COMMAND make
        INSTALL_COMMAND make install
        BUILD_IN_SOURCE 1
)

include(ExternalProject)
ExternalProject_Add(ffmpeg
        SOURCE_DIR ffmpeg-4.0
        CONFIGURE_COMMAND PKG_CONFIG_PATH=${X264_PREFIX_PATH}/pkgconfig:${FDKAAC_PREFIX_PATH}/pkgconfig ./configure --bindir=./bin --datadir=./data --docdir=./doc --libdir=./lib --shlibdir=./shlib --incdir=./include --disable-manpages --enable-gpl --enable-version3 --enable-shared --enable-static --enable-openssl --enable-nonfree --enable-avresample --enable-libx264 --enable-libfdk-aac
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

add_dependencies(ffmpeg fdkaac x264)
