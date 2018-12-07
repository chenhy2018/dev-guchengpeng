
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

if ($ENV{WITH_PJSIP} STREQUAL "ON")
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
endif()

SET(WOLFSSL_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/wolfssl/prefix/$ENV{ARCH}")
include(ExternalProject)
ExternalProject_Add(wolfssl
        SOURCE_DIR wolfssl

        CONFIGURE_COMMAND ./configure --prefix=${WOLFSSL_PREFIX_PATH} --enable-static --disable-shared
        BUILD_COMMAND make
        BUILD_IN_SOURCE 1
)

SET(WOLFMQTT_PREFIX_PATH "${CMAKE_CURRENT_SOURCE_DIR}/wolfMQTT/prefix/$ENV{ARCH}")

include(ExternalProject)
ExternalProject_Add(wolfMQTT
        SOURCE_DIR wolfMQTT
        CONFIGURE_COMMAND  ./autogen.sh  && ./configure --prefix=${WOLFMQTT_PREFIX_PATH} --enable-static --disable-shared --disable-examples --with-lib-prefix=${WOLFSSL_PREFIX_PATH}
        BUILD_COMMAND make
        BUILD_IN_SOURCE 1
)

include(ExternalProject)
ExternalProject_Add(curl-7.61.0
        SOURCE_DIR curl-7.61.0
        CONFIGURE_COMMAND ./configure --enable-shared=no --without-zlib --without-libidn2 --disable-libtool-lock --enable-http --disable-ftp --disable-file --disable-ldap --disable-ldaps --disable-rtsp --disable-proxy --disable-dict --disable-telnet --disable-tftp --disable-pop3 --disable-imap --disable-smb --disable-smtp --disable-gopher --disable-manual --disable-libcurl-option --enable-ipv6 --disable-largefile --disable-sspi --disable-ntlm-wb --disable-unix-sockets --disable-cookies --disable-crypto-auth --disable-tls-srp --without-ssl  --host=$ENV{HOST}
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

if ($ENV{ARCH} STREQUAL "x86")
SET(FASTFWD_RELEASE_PATH  "${CMAKE_CURRENT_SOURCE_DIR}/../release/fastfwd")
include(ExternalProject)
ExternalProject_Add(ffmpeg
        SOURCE_DIR ffmpeg-4.0
	CONFIGURE_COMMAND ./configure --libdir=${FASTFWD_RELEASE_PATH}/lib/x86_64/ffmpeg --incdir=${FASTFWD_RELEASE_PATH}/include --enable-gpl --enable-version3 --enable-nonfree --disable-shared --disable-shared --enable-static --enable-small --disable-doc --disable-programs --disable-debug --disable-avdevice --enable-avcodec --enable-avformat --enable-swresample --enable-swscale --disable-postproc --enable-avfilter --disable-avresample --disable-encoders --disable-decoders --enable-decoder=aac --enable-decoder=aac_latm --enable-decoder=h264 --enable-decoder=mpeg4 --disable-hwaccels --enable-muxers --enable-demuxers --enable-parsers --enable-bsfs --enable-protocols --disable-devices --disable-filters --disable-cuda --disable-cuvid --disable-nvenc --disable-v4l2_m2m
        BUILD_COMMAND make
        INSTALL_COMMAND make install
        BUILD_IN_SOURCE 1
)
endif()

#include(ExternalProject)
#ExternalProject_Add(librtmp
#        SOURCE_DIR rtmpdump-2.4
#        CONFIGURE_COMMAND ""
#        BUILD_COMMAND make
#        INSTALL_COMMAND ""
#        BUILD_IN_SOURCE 1
#)
#add_dependencies(ffmpeg fdkaac x264)
