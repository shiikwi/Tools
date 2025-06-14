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

HFONT WINAPI HookedCreateFontIndirectA(CONST LOGFONTA* lplf) {

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


HFONT WINAPI HookedCreateFontIndirectW(CONST LOGFONTW* lplf) {

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


HFONT WINAPI HookedCreateFontIndirectExA(CONST ENUMLOGFONTEXDVA* lplf) {

	WrapIsFirstCall("CreateFontIndirectExA Hooked\n");

	ENUMLOGFONTEXDVA lfModified = *lplf;

	lfModified.elfEnumLogfontEx.elfLogFont.lfCharSet= g_iCharSet;

	if (g_pszFaceName && *g_pszFaceName) {
		WideCharToMultiByte(
			CP_ACP, 0,
			g_pszFaceName, -1,
			lfModified.elfEnumLogfontEx.elfLogFont.lfFaceName, LF_FACESIZE,
			nullptr, nullptr
		);
	}


	printf("CharSet: %u -> %u\n", lplf->elfEnumLogfontEx.elfLogFont.lfCharSet, lfModified.elfEnumLogfontEx.elfLogFont.lfCharSet);
	printf("FaceName: %s -> %s\n", lplf->elfEnumLogfontEx.elfLogFont.lfFaceName, lfModified.elfEnumLogfontEx.elfLogFont.lfFaceName);

	return TrueCreateFontIndirectExA(&lfModified);
}

HFONT WINAPI HookedCreateFontIndirectExW(CONST ENUMLOGFONTEXDVW* lplf) {
	WrapIsFirstCall("CreateFontIndirectExW Hooked\n");

	ENUMLOGFONTEXDVW lfModified = *lplf;
	lfModified.elfEnumLogfontEx.elfLogFont.lfCharSet = g_iCharSet;

	if (g_pszFaceName && *g_pszFaceName) {
		wcscpy_s(lfModified.elfEnumLogfontEx.elfLogFont.lfFaceName, LF_FACESIZE, g_pszFaceName);
	}


	printf("CharSet: %u -> %u\n", lplf->elfEnumLogfontEx.elfLogFont.lfCharSet, lfModified.elfEnumLogfontEx.elfLogFont.lfCharSet);
	wprintf(L"FaceName: %s -> %s\n", lplf->elfEnumLogfontEx.elfLogFont.lfFaceName, lfModified.elfEnumLogfontEx.elfLogFont.lfFaceName);

	return TrueCreateFontIndirectExW(&lfModified);
}

int WINAPI HookedEnumFontFamiliesExA(
	HDC hdc,
	LPLOGFONTA lpLogfont,
	FONTENUMPROCA lpProc,
	LPARAM lParam,
	DWORD dwFlags) {
	WrapIsFirstCall("EnumFontFamiliesExA Hooked\n");

	LPLOGFONTA lfModified = lpLogfont;
	lfModified->lfCharSet = g_iCharSet;
	if (g_pszFaceName && *g_pszFaceName) {
		WideCharToMultiByte(
			CP_ACP, 0,
			g_pszFaceName, -1,
			lfModified->lfFaceName, LF_FACESIZE,
			nullptr, nullptr
		);
	}

	printf("CharSet: %u -> %u\n", lpLogfont->lfCharSet, lfModified->lfCharSet);
	printf("FaceName: %s -> %s\n", lpLogfont->lfFaceName, lfModified->lfFaceName);

	return TrueEnumFontFamiliesExA(hdc, lfModified, lpProc, lParam, dwFlags);
}

int WINAPI HookedEnumFontFamiliesExW(
	HDC hdc,
	LPLOGFONTW lpLogfont,
	FONTENUMPROCW lpProc,
	LPARAM lParam,
	DWORD dwFlags) {
	WrapIsFirstCall("EnumFontFamiliesExW Hooked\n");
	LPLOGFONTW lfModified = lpLogfont;
	lfModified->lfCharSet = g_iCharSet;
	if (g_pszFaceName && *g_pszFaceName) {
		wcscpy_s(lfModified->lfFaceName, LF_FACESIZE, g_pszFaceName);
	}
	printf("CharSet: %u -> %u\n", lpLogfont->lfCharSet, lfModified->lfCharSet);
	wprintf(L"FaceName: %s -> %s\n", lpLogfont->lfFaceName, lfModified->lfFaceName);
	return TrueEnumFontFamiliesExW(hdc, lfModified, lpProc, lParam, dwFlags);
}

int WINAPI HookedMultiByteToWideChar(
	UINT CodePage,
	DWORD dwFlags,
	LPCSTR lpMultiByteStr,
	int cbMultiByte,
	LPWSTR lpWideCharStr,
	int cchWideChar) {

	auto FirstPrint = [](const char* msg, int codepage) {
		static int flag = 0;
		if (flag < 2) {
			printf(msg, codepage);
			flag ++;
			}
		};

	FirstPrint("MultiByteToWideChar Hooked\nOriCodePage: %u\n", CodePage);

	if (CodePage == CP_ACP) {
		CodePage = 932;
	}
	
	FirstPrint("NewCodePage: %u\n", CodePage);
	return TrueMultiByteToWideChar(CodePage, dwFlags, lpMultiByteStr, cbMultiByte, lpWideCharStr, cchWideChar);
}

int WINAPI HookedWideCharToMultiByte(
	UINT CodePage,
	DWORD dwFlags,
	LPCWSTR lpWideCharStr,
	int cchWideChar,
	LPSTR lpMultiByteStr,
	int cbMultiByte,
	LPCSTR lpDefaultChar,
	LPBOOL lpUsedDefaultChar) {

	auto FirstPrint = [](const char* msg, int codepage) {
		static int flag = 0;
		if (flag < 2) {
			printf(msg, codepage);
			flag++;
			}
		};

	FirstPrint("WideCharToMultiByte Hooked\nOriCodePage: %u\n", CodePage);

	if (CodePage == CP_ACP) {
		CodePage = 936;
	}

	FirstPrint("NewCodePage: %u\n", CodePage);
	return TrueWideCharToMultiByte(CodePage, dwFlags, lpWideCharStr, cchWideChar, lpMultiByteStr, cbMultiByte, lpDefaultChar, lpUsedDefaultChar);
}

void LoadMyFont::Load() {
	std::string AnsiPath = [&]() {
		int len = WideCharToMultiByte(CP_ACP, 0, FontPath.c_str(), -1, NULL, 0, NULL, NULL);
		std::string buf(len, 0);
		WideCharToMultiByte(CP_ACP, 0, FontPath.c_str(), -1, buf.data(), len, nullptr, nullptr);
		return buf;
		}();

	auto AFontRet = AddFontResourceExA(AnsiPath.c_str(), FR_NOT_ENUM, 0);
	auto WFontRet = AddFontResourceExW(FontPath.c_str(), FR_NOT_ENUM, 0);

	IfLoad = (AFontRet && WFontRet);

	if (IfLoad) {
		printf("%s Loaded Successfully\n", AnsiPath.c_str());
	}
	else {
		printf("%s Loaded Unsuccessfully\n", AnsiPath.c_str());
	}
}

void LoadMyFont::UnLoad(){
		std::string AnsiPath = [&]() {
		int len = WideCharToMultiByte(CP_ACP, 0, FontPath.c_str(), -1, NULL, 0, NULL, NULL);
		std::string buf(len, 0);
		WideCharToMultiByte(CP_ACP, 0, FontPath.c_str(), -1, buf.data(), len, nullptr, nullptr);
		return buf;
		}();

		RemoveFontResourceExA(AnsiPath.c_str(), FR_NOT_ENUM, 0);
		RemoveFontResourceExW(FontPath.c_str(), FR_NOT_ENUM, 0);
}

/*
void LoadFont(const wchar_t* FontPath) {

	WrapIsFirstCall("AddFontResourceExAW Called\n");


	std::string AnsiPath = [&]() {
		int len = WideCharToMultiByte(CP_ACP, 0, FontPath, -1, NULL, 0, NULL, NULL);
		std::string buf(len, 0);
		WideCharToMultiByte(CP_ACP, 0, FontPath, -1, buf.data(), len, nullptr, nullptr);
		return buf;
		}();

	int AFontRet = AddFontResourceExA(AnsiPath.c_str(), FR_NOT_ENUM, 0);
	int WFontRet = AddFontResourceExW(FontPath, FR_NOT_ENUM, 0);

	if (AFontRet && WFontRet) {
		wprintf(L"%s Loaded Successfully\n", FontPath);
	}
	else {
		wprintf(L"%s Loaded Unsuccessfully\n", FontPath);
	}
}

void UnloadFont(const wchar_t* FontPath) {

	std::string AnsiPath = [&]() {
		int len = WideCharToMultiByte(CP_ACP, 0, FontPath, -1, NULL, 0, NULL, NULL);
		std::string buf(len, 0);
		WideCharToMultiByte(CP_ACP, 0, FontPath, -1, buf.data(), len, nullptr, nullptr);
		return buf;
		}();

	RemoveFontResourceExA(AnsiPath.c_str(), FR_NOT_ENUM, 0);
	RemoveFontResourceExW(FontPath, FR_NOT_ENUM, 0);
}
*/



void PatchRangeCheck() {
	HMODULE hModule = GetModuleHandle(NULL);
	if (!hModule) {
		printf("GetModuleHandle Failed\n");
		return;
	}

	MODULEINFO modInfo;
	if (!GetModuleInformation(GetCurrentProcess(), hModule, &modInfo, sizeof(modInfo))) {
		printf("GetModuleInformation Failed\n");
		return;
	}

	uint8_t* baseAddress = static_cast<uint8_t*>(modInfo.lpBaseOfDll);
	size_t ImageBaseSize = modInfo.SizeOfImage;
	int count = 0;

	for (size_t i = 0; i < ImageBaseSize - 3; i++) {
		if (baseAddress[i] == 0x3C) {
			if (baseAddress[i + 2] == 0x72 || baseAddress[i + 2] == 0x76 || baseAddress[i + 2] == 0x77) {
				auto cmpvalue = baseAddress[i + 1];
				if (cmpvalue == 0x9F || cmpvalue == 0xFC || cmpvalue == 0xA0) {
					DWORD oldprotect;
					VirtualProtect(&baseAddress[i + 1], 1, PAGE_EXECUTE_READWRITE, &oldprotect);
					baseAddress[i + 1] = 0xFE;
					VirtualProtect(&baseAddress[i + 1], 1, oldprotect, &oldprotect);
					printf("Find RangeCheck At: %p\n", &baseAddress[i]);
					count++;
				}
			}
		}
	}

	printf("Total Change to 0xFE: %d\n", count);
}