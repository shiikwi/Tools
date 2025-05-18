#pragma once
#include "Management.h"

namespace FontHookConfig {

	inline decltype(&CreateFontA) TrueCreateFontA = CreateFontA;
	inline decltype(&CreateFontW) TrueCreateFontW = CreateFontW;
	inline decltype(&CreateFontIndirectA) TrueCreateFontIndirectA = CreateFontIndirectA;
	inline decltype(&CreateFontIndirectW) TrueCreateFontIndirectW = CreateFontIndirectW;
	inline decltype(&CreateFontIndirectExA) TrueCreateFontIndirectExA = CreateFontIndirectExA;
	inline decltype(&CreateFontIndirectExW) TrueCreateFontIndirectExW = CreateFontIndirectExW;
	inline decltype(&EnumFontFamiliesExA) TrueEnumFontFamiliesExA = EnumFontFamiliesExA;
	inline decltype(&EnumFontFamiliesExW) TrueEnumFontFamiliesExW = EnumFontFamiliesExW;

}



HFONT WINAPI HookedCreateFontA(
	int     cHeight,
	int     cWidth,
	int     cEscapement,
	int     cOrientation,
	int     cWeight,
	DWORD   bItalic,
	DWORD   bUnderline,
	DWORD   bStrikeOut,
	DWORD   iCharSet,
	DWORD   iOutPrecision,
	DWORD   iClipPrecision,
	DWORD   iQuality,
	DWORD   iPitchAndFamily,
	LPCSTR pszFaceName
);


HFONT WINAPI HookedCreateFontW(
	int     cHeight,
	int     cWidth,
	int     cEscapement,
	int     cOrientation,
	int     cWeight,
	DWORD   bItalic,
	DWORD   bUnderline,
	DWORD   bStrikeOut,
	DWORD   iCharSet,
	DWORD   iOutPrecision,
	DWORD   iClipPrecision,
	DWORD   iQuality,
	DWORD   iPitchAndFamily,
	LPCWSTR pszFaceName
);


HFONT WINAPI HookedCreateFontIndirectA(
	CONST LOGFONTA* lplf
);


HFONT WINAPI HookedCreateFontIndirectW(
	CONST LOGFONTW* lplf
);

HFONT WINAPI HookedCreateFontIndirectExA(
	CONST ENUMLOGFONTEXDVA* lplf
);

HFONT WINAPI HookedCreateFontIndirectExW(
	CONST ENUMLOGFONTEXDVW* lplf
);

int WINAPI HookedEnumFontFamiliesExA(
	HDC hdc,
	LPLOGFONTA lpLogfont,
	FONTENUMPROCA lpProc,
	LPARAM lParam,
	DWORD dwFlags
);

int WINAPI HookedEnumFontFamiliesExW(
	HDC hdc,
	LPLOGFONTW lpLogfont,
	FONTENUMPROCW lpProc,
	LPARAM lParam,
	DWORD dwFlags
);

void LoadFont(const wchar_t* FontPath);

void UnloadFont(const wchar_t* FontPath);