#pragma once
#include <Windows.h>
#include "detours.h"
#include <iostream>
#include <vector>
#include <string>

inline const wchar_t* NEW_TITLE = L"HOOK ´°¿Ú²âÊÔ";
inline const DWORD g_iCharSet = 134; //SJIS 128
inline const LPCWSTR g_pszFaceName = L"simhei";  //nullptr
inline const std::wstring MyFont = L"WenQuanYi.ttf";


static void WrapIsFirstCall(const char* message) {
    static bool IsFirstCall = true;
    if (IsFirstCall) {
        printf(message);
        IsFirstCall = false;
    }
}

static void AttachDebugConsole();