option(WITH_TLS
        "Include SSL/TLS support?" ON)
option(WITH_TLS_PSK
        "Include TLS-PSK support (requires WITH_TLS)?" OFF)
option(WITH_EC
        "Include Elliptic Curve support (requires WITH_TLS)?" OFF)

option(WITH_UUID
        "Build using libuuid for clientid generation (Linux only - please report if
supported on your platform)." OFF)


if (${WITH_TLS} STREQUAL ON)
        find_package(OpenSSL REQUIRED)
        add_definitions("-DWITH_TLS")

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

SET(GCC_INCLUDE_PATH "-I../../third_party/mosquitto-1.5/lib/ -I../../third_party/mosquitto-1.5/")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_MOSQUITTO_SRC)
aux_source_directory(../../third_party/mosquitto-1.5/lib/ MOSQUITTO_SRC)

add_library (mqtt STATIC ${LIB_MOSQUITTO_SRC} ${MOSQUITTO_SRC})

add_library(mqtt_dy SHARED ${LIB_MOSQUITTO_SRC} ${MOSQUITTO_SRC})
set_target_properties(mqtt_dy PROPERTIES OUTPUT_NAME "mqtt")
set_target_properties(mqtt_dy PROPERTIES VERSION 1.0  SOVERSION 1)


add_subdirectory(test)
