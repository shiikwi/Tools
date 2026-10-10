set_project("alphamovie")
set_version("0.1.0")
set_languages("cxx17")

add_rules("mode.debug", "mode.release")

target("alphamovie")
    set_kind("binary")
    add_files("main.cpp", "miniz.c", "miniz_tdef.c", "miniz_tinfl.c", "miniz_zip.c")
    add_includedirs(".")
    set_warnings("all")
    if is_plat("windows") then
        add_defines("_CRT_SECURE_NO_WARNINGS")
    end
