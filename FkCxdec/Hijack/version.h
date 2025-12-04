#include <Windows.h>

#pragma comment(linker, "/EXPORT:GetFileVersionInfoA=_Redirect_GetFileVersionInfoA,@1")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoByHandle=_Redirect_GetFileVersionInfoByHandle,@2")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoExA=_Redirect_GetFileVersionInfoExA,@3")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoExW=_Redirect_GetFileVersionInfoExW,@4")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoSizeA=_Redirect_GetFileVersionInfoSizeA,@5")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoSizeExA=_Redirect_GetFileVersionInfoSizeExA,@6")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoSizeExW=_Redirect_GetFileVersionInfoSizeExW,@7")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoSizeW=_Redirect_GetFileVersionInfoSizeW,@8")
#pragma comment(linker, "/EXPORT:GetFileVersionInfoW=_Redirect_GetFileVersionInfoW,@9")
#pragma comment(linker, "/EXPORT:VerFindFileA=_Redirect_VerFindFileA,@10")
#pragma comment(linker, "/EXPORT:VerFindFileW=_Redirect_VerFindFileW,@11")
#pragma comment(linker, "/EXPORT:VerInstallFileA=_Redirect_VerInstallFileA,@12")
#pragma comment(linker, "/EXPORT:VerInstallFileW=_Redirect_VerInstallFileW,@13")
#pragma comment(linker, "/EXPORT:VerLanguageNameA=_Redirect_VerLanguageNameA,@14")
#pragma comment(linker, "/EXPORT:VerLanguageNameW=_Redirect_VerLanguageNameW,@15")
#pragma comment(linker, "/EXPORT:VerQueryValueA=_Redirect_VerQueryValueA,@16")
#pragma comment(linker, "/EXPORT:VerQueryValueW=_Redirect_VerQueryValueW,@17")

PVOID Real_GetFileVersionInfoA = NULL;
PVOID Real_GetFileVersionInfoByHandle = NULL;
PVOID Real_GetFileVersionInfoExA = NULL;
PVOID Real_GetFileVersionInfoExW = NULL;
PVOID Real_GetFileVersionInfoSizeA = NULL;
PVOID Real_GetFileVersionInfoSizeExA = NULL;
PVOID Real_GetFileVersionInfoSizeExW = NULL;
PVOID Real_GetFileVersionInfoSizeW = NULL;
PVOID Real_GetFileVersionInfoW = NULL;
PVOID Real_VerFindFileA = NULL;
PVOID Real_VerFindFileW = NULL;
PVOID Real_VerInstallFileA = NULL;
PVOID Real_VerInstallFileW = NULL;
PVOID Real_VerLanguageNameA = NULL;
PVOID Real_VerLanguageNameW = NULL;
PVOID Real_VerQueryValueA = NULL;
PVOID Real_VerQueryValueW = NULL;

HMODULE Real_Module = NULL;

FARPROC WINAPI GetAddress(LPCSTR lpProcName)
{
    FARPROC pAddress = GetProcAddress(Real_Module, lpProcName);
    if (pAddress == NULL)
    {
        MessageBoxW(NULL, L"Get address failed", L"version", MB_OK);
        ExitProcess(1);
    }
    return pAddress;
}

VOID WINAPI FreeHijack()
{
    if (Real_Module != NULL)
        FreeLibrary(Real_Module);
    return;
}

VOID WINAPI InitHijack()
{
    WCHAR real_dll_path[MAX_PATH];
    GetSystemDirectoryW(real_dll_path, MAX_PATH);
    lstrcatW(real_dll_path, L"\\version.dll");

    Real_Module = LoadLibraryW(real_dll_path);
    if (Real_Module == NULL)
    {
        MessageBoxW(NULL, L"Load original dll failed", L"version", MB_OK);
        ExitProcess(1);
    }

    Real_GetFileVersionInfoA = GetAddress("GetFileVersionInfoA");
    Real_GetFileVersionInfoByHandle = GetAddress("GetFileVersionInfoByHandle");
    Real_GetFileVersionInfoExA = GetAddress("GetFileVersionInfoExA");
    Real_GetFileVersionInfoExW = GetAddress("GetFileVersionInfoExW");
    Real_GetFileVersionInfoSizeA = GetAddress("GetFileVersionInfoSizeA");
    Real_GetFileVersionInfoSizeExA = GetAddress("GetFileVersionInfoSizeExA");
    Real_GetFileVersionInfoSizeExW = GetAddress("GetFileVersionInfoSizeExW");
    Real_GetFileVersionInfoSizeW = GetAddress("GetFileVersionInfoSizeW");
    Real_GetFileVersionInfoW = GetAddress("GetFileVersionInfoW");
    Real_VerFindFileA = GetAddress("VerFindFileA");
    Real_VerFindFileW = GetAddress("VerFindFileW");
    Real_VerInstallFileA = GetAddress("VerInstallFileA");
    Real_VerInstallFileW = GetAddress("VerInstallFileW");
    Real_VerLanguageNameA = GetAddress("VerLanguageNameA");
    Real_VerLanguageNameW = GetAddress("VerLanguageNameW");
    Real_VerQueryValueA = GetAddress("VerQueryValueA");
    Real_VerQueryValueW = GetAddress("VerQueryValueW");
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoA(void)
{
    __asm jmp Real_GetFileVersionInfoA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoByHandle(void)
{
    __asm jmp Real_GetFileVersionInfoByHandle;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoExA(void)
{
    __asm jmp Real_GetFileVersionInfoExA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoExW(void)
{
    __asm jmp Real_GetFileVersionInfoExW;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoSizeA(void)
{
    __asm jmp Real_GetFileVersionInfoSizeA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoSizeExA(void)
{
    __asm jmp Real_GetFileVersionInfoSizeExA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoSizeExW(void)
{
    __asm jmp Real_GetFileVersionInfoSizeExW;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoSizeW(void)
{
    __asm jmp Real_GetFileVersionInfoSizeW;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_GetFileVersionInfoW(void)
{
    __asm jmp Real_GetFileVersionInfoW;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerFindFileA(void)
{
    __asm jmp Real_VerFindFileA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerFindFileW(void)
{
    __asm jmp Real_VerFindFileW;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerInstallFileA(void)
{
    __asm jmp Real_VerInstallFileA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerInstallFileW(void)
{
    __asm jmp Real_VerInstallFileW;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerLanguageNameA(void)
{
    __asm jmp Real_VerLanguageNameA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerLanguageNameW(void)
{
    __asm jmp Real_VerLanguageNameW;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerQueryValueA(void)
{
    __asm jmp Real_VerQueryValueA;
}

EXTERN_C __declspec(naked) void __cdecl Redirect_VerQueryValueW(void)
{
    __asm jmp Real_VerQueryValueW;
}

