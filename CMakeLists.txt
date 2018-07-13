cmake_minimum_required (VERSION 2.8)

project (linking)

SET(GCC_COVERAGE_COMPILE_FLAGS "-Wall -Wno-unused -g")

if ($ENV{WITH_P2P} STREQUAL "ON")
        add_definitions("-DWITH_P2P")
endif()

add_subdirectory(third_party)
add_subdirectory(link)
