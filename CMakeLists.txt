cmake_minimum_required (VERSION 2.8)

project (linking)

SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")

if ($ENV{WITH_P2P} STREQUAL "ON")
        add_definitions("-DWITH_P2P")
endif()

if ($ENV{ARCH} STREQUAL "x86")
    add_subdirectory(third_party)
elseif ($ENV{ARCH} STREQUAL "80386")
    add_subdirectory(third_party)
endif()
add_subdirectory(link)
