cmake_minimum_required(VERSION 3.11)

project(ScnParse)

set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)
set(CMAKE_CXX_STANDARD 17)

include(FetchContent)

FetchContent_Declare(
    json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.12.0
)

FetchContent_MakeAvailable(json)

aux_source_directory(src SRC_LIST)
add_executable(ScnParse ${SRC_LIST})

target_link_libraries(ScnParse PRIVATE nlohmann_json::nlohmann_json)

if (MSVC)
    target_compile_options(ScnParse PRIVATE "/utf-8")
endif()