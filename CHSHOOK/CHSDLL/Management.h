#pragma once
#include <Windows.h>
#include <psapi.h>
#include <iostream>
#include <vector>
#include <string>
#include <memory>
#include "detours.h"

inline const wchar_t* NEW_TITLE = L"HOOK ´°¿Ú²âÊÔ";
inline const DWORD g_iCharSet = 134; //SJIS 128
inline const LPCWSTR g_pszFaceName = L"simhei";  //nullptr
inline const std::wstring MyFont = L"WenQuanYi.ttf";

static const bool IfTrueCreateWindowExA          =    0;
static const bool IfTrueCreateWindowExW          =    0;

static const bool IfTrueDefWindowProcA           =    0;      
static const bool IfTrueDefWindowProcW           =    0;

static const bool IfTrueCreateFontA              =    1;
static const bool IfTrueCreateFontW              =    1;

static const bool IfTrueCreateFontIndirectA      =    0;
static const bool IfTrueCreateFontIndirectW      =    0;

static const bool IfTrueCreateFontIndirectExA    =    0;
static const bool IfTrueCreateFontIndirectExW    =    0;

static const bool IfTrueEnumFontFamiliesExA      =    0;
static const bool IfTrueEnumFontFamiliesExW      =    0;

static const bool IfTrueMultiByteToWideChar      =    0;
static const bool IfTrueWideCharToMultiByte      =    0;

#define    IfLoadMyFont                               1

#define    IfRangeCheck                               0

#define    IfAttachDebugConsole                       1


static void WrapIsFirstCall(const char* message) {
    static bool IsFirstCall = true;
    if (IsFirstCall) {
        printf(message);
        IsFirstCall = false;
    }
}

static void AttachDebugConsole() {
    AllocConsole();
    FILE* fDummy;
    freopen_s(&fDummy, "CONOUT$", "w", stdout);
    freopen_s(&fDummy, "CONOUT$", "w", stderr);
    SetConsoleTitleW(L"Hook Debug Console");
}

class LoadMyFont {
private:
	std::wstring FontPath;
	bool IfLoad = false;
	void Load();
	void UnLoad();

public:
	LoadMyFont(const wchar_t* fPath) :FontPath(fPath) {
		printf("AddFontResourceExAW Called\n");
		Load();
	}
	~LoadMyFont() {
		UnLoad();
	}

	LoadMyFont(const LoadMyFont&) = delete;
	LoadMyFont& operator=(const LoadMyFont&) = delete;

	bool getIfLoad() { return IfLoad; }

};

template<bool ENABLE>
inline void DetourAttach_IF_Enable(PVOID* realfunc, PVOID hookfunc) {
    if (ENABLE) {
        DetourAttach(realfunc, hookfunc);
    }
}

template<bool ENABLE>
inline void DetourDetach_IF_Enable(PVOID* realfunc, PVOID hookfunc) {
    if (ENABLE) {
        DetourDetach(realfunc, hookfunc);
    }
}