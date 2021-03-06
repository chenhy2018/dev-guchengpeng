option(WITH_THREADING "Include client library threading support?" ON)
if (${WITH_THREADING} STREQUAL ON)
        add_definitions("-DWITH_THREADING")
        if (WIN32)
                if (CMAKE_CL_64)
                        set (PTHREAD_LIBRARIES C:\\pthreads\\Pre-built.2\\lib\\x64\\pthreadVC2.lib)
                else (CMAKE_CL_64)
                        set (PTHREAD_LIBRARIES C:\\pthreads\\Pre-built.2\\lib\\x86\\pthreadVC2.lib)
                endif (CMAKE_CL_64)
                set (PTHREAD_INCLUDE_DIR C:\\pthreads\\Pre-built.2\\include)
        else (WIN32)
                find_library(LIBPTHREAD pthread)
                if (LIBPTHREAD)
                        set (PTHREAD_LIBRARIES pthread)
                else (LIBPTHREAD)
                        set (PTHREAD_LIBRARIES "")
                endif()
                set (PTHREAD_INCLUDE_DIR "")
        endif (WIN32)
else (${WITH_THREADING} STREQUAL ON)
        set (PTHREAD_LIBRARIES "")
        set (PTHREAD_INCLUDE_DIR "")
endif (${WITH_THREADING} STREQUAL ON)

if ($ENV{WITH_OPENSSL} STREQUAL "ON")
option(WITH_TLS
        "Include SSL/TLS support?" ON)
else()
option(WITH_TLS
        "Include SSL/TLS support?" OFF)
endif()

option(WITH_TLS_PSK
        "Include TLS-PSK support (requires WITH_TLS)?" OFF)
option(WITH_EC
        "Include Elliptic Curve support (requires WITH_TLS)?" OFF)

option(WITH_UUID
        "Build using libuuid for clientid generation (Linux only - please report if
supported on your platform)." OFF)


if (${WITH_TLS} STREQUAL ON)
        add_definitions("-DWITH_TLS")
        message("DWITH_TLS")

        if (${WITH_TLS_PSK} STREQUAL ON)
                add_definitions("-DWITH_TLS_PSK")
        endif (${WITH_TLS_PSK} STREQUAL ON)

        if (${WITH_EC} STREQUAL ON)
                add_definitions("-DWITH_EC")
        endif (${WITH_EC} STREQUAL ON)
else (${WITH_TLS} STREQUAL ON)
        set (OPENSSL_INCLUDE_DIR "")
endif (${WITH_TLS} STREQUAL ON)

option(WITH_SOCKS "Include SOCKS5 support?" OFF)
if (${WITH_SOCKS} STREQUAL ON)
        add_definitions("-DWITH_SOCKS")
endif (${WITH_SOCKS} STREQUAL ON)

if (${WITH_UUID} STREQUAL ON)
        add_definitions("-DWITH_UUID")
endif (${WITH_UUID} STREQUAL ON)

option(WITH_SRV "Include SRV lookup support?" OFF)

if ($ENV{WITH_MOSQUITTO} STREQUAL "ON")
        add_definitions("-DWITH_MOSQUITTO")
        if ($ENV{WITH_OPENSSL} STREQUAL "ON")
        SET(GCC_INCLUDE_PATH "-I../../third_party/mosquitto-1.5/lib/ -I../../third_party/mosquitto-1.5/ -I ../../third_party/openssl-1.1.0h/prefix/$ENV{ARCH}/include/ -I ./mosquitto")
        else()
        SET(GCC_INCLUDE_PATH "-I../../third_party/mosquitto-1.5/lib/ -I../../third_party/mosquitto-1.5/ -I ./mosquitto")
        endif()
else()
        add_definitions("-DWITH_WOLFMQTT")
	SET(GCC_INCLUDE_PATH "-I../../third_party/wolfMQTT/prefix/$ENV{ARCH}/include  -I../../third_party/wolfMQTT/prefix/$ENV{ARCH}/include/wolfmqtt -I../../third_party/wolfssl/prefix/$ENV{ARCH}/include -I./wolfmqtt")
endif()

SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_MQTT)
aux_source_directory(./cJSON LIB_MQTT)

if ($ENV{WITH_MOSQUITTO} STREQUAL "ON")
aux_source_directory(../../third_party/mosquitto-1.5/lib/ MOSQUITTO_SRC)
aux_source_directory(./mosquitto MOSQUITTO_SRC)
else()
aux_source_directory(./wolfmqtt MOSQUITTO_SRC)
endif()

add_library (mqtt-$ENV{LIBPREFIX} STATIC ${LIB_MQTT} ${MOSQUITTO_SRC})

add_library(mqtt_dy-$ENV{LIBPREFIX} SHARED ${LIB_MQTT} ${MOSQUITTO_SRC})
set_target_properties(mqtt_dy-$ENV{LIBPREFIX} PROPERTIES OUTPUT_NAME "mqtt")
set_target_properties(mqtt_dy-$ENV{LIBPREFIX} PROPERTIES VERSION 1.0  SOVERSION 1)

add_subdirectory(test)
