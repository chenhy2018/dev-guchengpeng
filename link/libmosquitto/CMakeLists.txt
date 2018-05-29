option(WITH_TLS
        "Include SSL/TLS support?" ON)
option(WITH_TLS_PSK
        "Include TLS-PSK support (requires WITH_TLS)?" ON)
option(WITH_EC
        "Include Elliptic Curve support (requires WITH_TLS)?" ON)
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

option(WITH_SOCKS "Include SOCKS5 support?" ON)
if (${WITH_SOCKS} STREQUAL ON)
        add_definitions("-DWITH_SOCKS")
endif (${WITH_SOCKS} STREQUAL ON)

option(WITH_SRV "Include SRV lookup support?" ON)

SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")
SET(GCC_INCLUDE_PATH "-I../../third_party/mosquitto-1.4.8/lib/ -I../../third_party/mosquitto-1.4.8/")
SET(CMAKE_C_FLAGS "${GCC_INCLUDE_PATH} ${GCC_COVERAGE_COMPILE_FLAGS}")

aux_source_directory(. LIB_MOSQUITTO_SRC)
aux_source_directory(../../third_party/mosquitto-1.4.8/lib/ MOSQUITTO_SRC)

add_library (emqtt ${LIB_MOSQUITTO_SRC} ${MOSQUITTO_SRC})
target_link_libraries(emqtt)

add_subdirectory(test)
