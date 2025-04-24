cmake_minimum_required(VERSION 3.10)

project(SJIS2UTF8)

set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)

aux_source_directory(src SRC_LIST)

add_executable(SJIS2UTF8 ${SRC_LIST})

if(WIN32)
    target_link_libraries(SJIS2UTF8 
        kernel32.lib
        user32.lib
    )
    add_compile_definitions(_CRT_SECURE_NO_WARNINGS)
endif()