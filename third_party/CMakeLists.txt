if ($ENV{WITH_OPENSSL} STREQUAL "ON")
SET(OPENSSL_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/openssl-1.1.0h/prefix/$ENV{ARCH}")
include(ExternalProject)
ExternalProject_Add(openssl-1.1.0h
        SOURCE_DIR openssl-1.1.0h
        BUILD_COMMAND make COMMAND make install
        CONFIGURE_COMMAND ./Configure --prefix=${OPENSSL_PREFIX_PATH} my-gcc;
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)
endif()

SET(PJ_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/pjproject-2.7.2/prefix/$ENV{ARCH}")
include(ExternalProject)

if ($ENV{WITH_OPENSSL} STREQUAL "ON")
ExternalProject_Add(pjsip
        DEPENDS openssl-1.1.0h
        SOURCE_DIR pjproject-2.7.2
        CONFIGURE_COMMAND ./configure  --host=$ENV{HOST}  --with-ssl=${OPENSSL_PREFIX_PATH} --prefix=${PJ_PREFIX_PATH} --disable-sdl --disable-ffmpeg --disable-v4l2 --disable-openh264 --disable-libwebrtc --disable-sound --disable-oss --disable-video --disable-speex-aec --disable-g711-code --disable-l16-codec --disable-gsm-codec --disable-g722-codec --disable-g7221-codec --disable-speex-codec --disable-ilbc-codec --disable-bcg729 --disable-libyuv
        BUILD_COMMAND make dep COMMAND make
        INSTALL_COMMAND make install
        BUILD_IN_SOURCE 1
)
else()
add_definitions(-DPJMEDIA_HAS_SRTP=0)
ExternalProject_Add(pjsip
        SOURCE_DIR pjproject-2.7.2
        CONFIGURE_COMMAND ./configure  --host=$ENV{HOST}  --prefix=${PJ_PREFIX_PATH} --disable-ssl --disable-sdl --disable-ffmpeg --disable-v4l2 --disable-openh264 --disable-libwebrtc --disable-sound --disable-oss --disable-video --disable-speex-aec --disable-g711-code --disable-l16-codec --disable-gsm-codec --disable-g722-codec --disable-g7221-codec --disable-speex-codec --disable-ilbc-codec --disable-bcg729 --disable-libyuv
        BUILD_COMMAND make dep COMMAND make
        INSTALL_COMMAND make install
        BUILD_IN_SOURCE 1
)
endif()

SET(PJ_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/curl-7.61.0/prefix/$ENV{ARCH}")
include(ExternalProject)
ExternalProject_Add(curl-7.61.0
        SOURCE_DIR curl-7.61.0
        CONFIGURE_COMMAND ./configure --enable-shared=no --without-zlib --without-libidn2 --disable-libtool-lock --enable-http --disable-ftp --disable-file --disable-ldap --disable-ldaps --disable-rtsp --disable-proxy --disable-dict --disable-telnet --disable-tftp --disable-pop3 --disable-imap --disable-smb --disable-smtp --disable-gopher --disable-manual --disable-libcurl-option --enable-ipv6 --disable-largefile --disable-sspi --disable-ntlm-wb --disable-unix-sockets --disable-cookies --disable-crypto-auth --disable-tls-srp --without-ssl 
        BUILD_COMMAND make
        BUILD_IN_SOURCE 1
)

#SET(X264_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/x264/prefix")
#include(ExternalProject)
#ExternalProject_Add(x264
#        SOURCE_DIR x264
#        BUILD_COMMAND make
#        CONFIGURE_COMMAND ./configure --enable-shared --enable-static --disable-opencl --libdir=${X264_PREFIX_PATH} --includedir=${X264_PREFIX_PATH} --prefix=${X264_PREFIX_PATH}
#        INSTALL_COMMAND make install
#        BUILD_IN_SOURCE 1
#)

#SET(FDKAAC_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/fdkaac-0.1.6/prefix")
#include(ExternalProject)
#ExternalProject_Add(fdkaac
#        SOURCE_DIR fdkaac-0.1.6
#        CONFIGURE_COMMAND ./autogen.sh COMMAND ./configure --prefix=${FDKAAC_PREFIX_PATH} --libdir=${FDKAAC_PREFIX_PATH} --includedir=${FDKAAC_PREFIX_PATH} --bindir=${FDKAAC_PREFIX_PATH}
#        BUILD_COMMAND make
#        INSTALL_COMMAND make install
#        BUILD_IN_SOURCE 1
#)

include(ExternalProject)
ExternalProject_Add(ffmpeg
        SOURCE_DIR ffmpeg-4.0
        CONFIGURE_COMMAND PKG_CONFIG_PATH=${X264_PREFIX_PATH}/pkgconfig:${FDKAAC_PREFIX_PATH}/pkgconfig ./configure --disable-x86asm --disable-zlib --enable-small --disable-programs --disable-doc --disable-avdevice --disable-swresample --disable-swscale --disable-postproc --disable-avfilter --disable-network --disable-dct --disable-dwt --disable-error-resilience --disable-lsp --disable-lzo --disable-mdct --disable-rdft --disable-fft --disable-faan --disable-pixelutils --disable-everything --enable-muxer=mpegts --bindir=./bin --datadir=./data --docdir=./doc --libdir=./lib
        BUILD_COMMAND make
        INSTALL_COMMAND ""
        BUILD_IN_SOURCE 1
)

#include(ExternalProject)
#ExternalProject_Add(librtmp
#        SOURCE_DIR rtmpdump-2.4
#        CONFIGURE_COMMAND ""
#        BUILD_COMMAND make
#        INSTALL_COMMAND ""
#        BUILD_IN_SOURCE 1
#)

#add_dependencies(ffmpeg fdkaac x264)
