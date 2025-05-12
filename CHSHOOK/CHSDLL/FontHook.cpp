#include "FontHook.h"
#pragma comment(lib, "detours.lib")
using namespace FontHookConfig;

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
	LPCSTR pszFaceName) {

	WrapIsFirstCall("CreateFontA Hooked\n");

	const DWORD NewCharSet = g_iCharSet ? g_iCharSet : iCharSet;

	static std::string forcedFontNameA;
	LPCSTR NewFontName = pszFaceName;

	if (g_pszFaceName) {
		int len = WideCharToMultiByte(CP_ACP, 0, g_pszFaceName, -1, NULL, 0, NULL, NULL);
		forcedFontNameA.resize(len);
		WideCharToMultiByte(CP_ACP, 0, g_pszFaceName, -1, &forcedFontNameA[0], len, NULL, NULL);
		NewFontName = forcedFontNameA.c_str();
	}

	printf("CharSet: %u -> %u\n", iCharSet, NewCharSet);
	printf("FaceName: %s -> %s\n", pszFaceName, NewFontName);
	return TrueCreateFontA(cHeight, cWidth, cEscapement, cOrientation, cWeight, bItalic, bUnderline, bStrikeOut, NewCharSet, iOutPrecision, iClipPrecision, iQuality, iPitchAndFamily, NewFontName);
}



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
	LPCWSTR pszFaceName) {

	WrapIsFirstCall("CreateFontW Hooked\n");


	const DWORD NewCharSet = g_iCharSet ? g_iCharSet : iCharSet;
	const LPCWSTR NewFontName = g_pszFaceName ? g_pszFaceName : pszFaceName;


	printf("CharSet: %u -> %u\n", iCharSet, NewCharSet);
	wprintf(L"FaceName: %s -> %s\n", pszFaceName, NewFontName);
	return TrueCreateFontW(cHeight, cWidth, cEscapement, cOrientation, cWeight, bItalic, bUnderline, bStrikeOut, NewCharSet, iOutPrecision, iClipPrecision, iQuality, iPitchAndFamily, NewFontName);
}

HFONT WINAPI HookedCreateFontIndirectA(CONST LOGFONTA* lplf)
{

	WrapIsFirstCall("CreateFontIndirectA Hooked\n");

	LOGFONTA lfModified = *lplf;
	
	lfModified.lfCharSet = g_iCharSet;

	if (g_pszFaceName && *g_pszFaceName) {
		WideCharToMultiByte(
			CP_ACP, 0,
			g_pszFaceName, -1,
			lfModified.lfFaceName, LF_FACESIZE,
			nullptr, nullptr
		);
	}


	printf("CharSet: %u -> %u\n", lplf->lfCharSet, lfModified.lfCharSet);
	printf("FaceName: %s -> %s\n", lplf->lfFaceName, lfModified.lfFaceName);
	return TrueCreateFontIndirectA(&lfModified);
}


HFONT WINAPI HookedCreateFontIndirectW(CONST LOGFONTW* lplf)
{

	WrapIsFirstCall("CreateFontIndirectW Hooked\n");

	LOGFONTW lfModified = *lplf;
	lfModified.lfCharSet = g_iCharSet;

	if (g_pszFaceName && *g_pszFaceName) {
		wcscpy_s(lfModified.lfFaceName, LF_FACESIZE, g_pszFaceName);
	}


	printf("CharSet: %u -> %u\n", lplf->lfCharSet, lfModified.lfCharSet);
	wprintf(L"FaceName: %s -> %s\n", lplf->lfFaceName, lfModified.lfFaceName);
	return TrueCreateFontIndirectW(&lfModified);
}