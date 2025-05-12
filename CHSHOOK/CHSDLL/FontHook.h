#pragma once
#include "Management.h"

namespace FontHookConfig {
	inline const DWORD g_iCharSet = 134; //SJIS 128
	inline const LPCWSTR g_pszFaceName = L"simhei";  //nullptr

	inline decltype(&CreateFontA) TrueCreateFontA = CreateFontA;
	inline decltype(&CreateFontW) TrueCreateFontW = CreateFontW;
	inline decltype(&CreateFontIndirectA) TrueCreateFontIndirectA = CreateFontIndirectA;
	inline decltype(&CreateFontIndirectW) TrueCreateFontIndirectW = CreateFontIndirectW;
	inline decltype(&CreateFontIndirectExA) TrueCreateFontIndirectExA = CreateFontIndirectExA;
	inline decltype(&CreateFontIndirectExW) TrueCreateFontIndirectExW = CreateFontIndirectExW;

	/*EnumFontFamiliesExA
	EnumFontFamiliesExW
	*/


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