cmake_minimum_required(VERSION 3.15)

project(ExHIBITHijack LANGUAGES CXX)

add_library(version_proxy SHARED
    src/dllmain.cpp
)

target_include_directories(version_proxy PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src
)

target_link_libraries(version_proxy PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/src/detours.lib"
    Version
)

target_compile_features(version_proxy PRIVATE cxx_std_20)

target_compile_definitions(version_proxy PRIVATE
    WIN32_LEAN_AND_MEAN
    NOMINMAX
    UNICODE
    _UNICODE
)

target_compile_options(version_proxy PRIVATE
    /source-charset:utf-8
)

set_target_properties(version_proxy PROPERTIES
    OUTPUT_NAME "version"
    PREFIX ""
)