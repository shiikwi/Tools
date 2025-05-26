#include "WindowHook.h"
#include "FontHook.h"
#include "Management.h"
#pragma comment(lib, "detours.lib")
using namespace WindowHookConfig;
using namespace FontHookConfig;

extern "C" __declspec(dllexport) void ShiKwi(void) {}


bool InitializeHooks()
{
    DetourRestoreAfterWith();
    DetourTransactionBegin();
    DetourUpdateThread(GetCurrentThread());
    DetourAttach_IF_Enable<IfTrueCreateWindowExA>(&(PVOID&)TrueCreateWindowExA, HookedCreateWindowExA);
    DetourAttach_IF_Enable<IfTrueCreateWindowExW>(&(PVOID&)TrueCreateWindowExW, HookedCreateWindowExW);
    DetourAttach_IF_Enable<IfTrueDefWindowProcA>(&(PVOID&)TrueDefWindowProcA, HookedDefWindowProcA);
    DetourAttach_IF_Enable<IfTrueDefWindowProcW>(&(PVOID&)TrueDefWindowProcW, HookedDefWindowProcW);
    DetourAttach_IF_Enable<IfTrueCreateFontA>(&(PVOID&)TrueCreateFontA, HookedCreateFontA);
    DetourAttach_IF_Enable<IfTrueCreateFontW>(&(PVOID&)TrueCreateFontW, HookedCreateFontW);
    DetourAttach_IF_Enable<IfTrueCreateFontIndirectA>(&(PVOID&)TrueCreateFontIndirectA, HookedCreateFontIndirectA);
    DetourAttach_IF_Enable<IfTrueCreateFontIndirectW>(&(PVOID&)TrueCreateFontIndirectW, HookedCreateFontIndirectW);
    DetourAttach_IF_Enable<IfTrueCreateFontIndirectExW>(&(PVOID&)TrueCreateFontIndirectExA, HookedCreateFontIndirectExA);
    DetourAttach_IF_Enable<IfTrueCreateFontIndirectExW>(&(PVOID&)TrueCreateFontIndirectExW, HookedCreateFontIndirectExW);
    DetourAttach_IF_Enable<IfTrueEnumFontFamiliesExA>(&(PVOID&)TrueEnumFontFamiliesExA, HookedEnumFontFamiliesExA);
    DetourAttach_IF_Enable<IfTrueEnumFontFamiliesExW>(&(PVOID&)TrueEnumFontFamiliesExW, HookedEnumFontFamiliesExW);
    DetourAttach_IF_Enable<IfTrueMultiByteToWideChar>(&(PVOID&)TrueMultiByteToWideChar, HookedMultiByteToWideChar);
    DetourAttach_IF_Enable<IfTrueWideCharToMultiByte>(&(PVOID&)TrueWideCharToMultiByte, HookedWideCharToMultiByte);
	LoadFont(MyFont.c_str());

    LONG status = DetourTransactionCommit();
    if (status != NO_ERROR) {
    printf("[CHSHOOK] hook初始化失败\n");
        return false;
    }
    printf("[CHSHOOK] hook初始化成功\n");
    return true;
}


bool FreeHooks()
{
    DetourTransactionBegin();
    DetourUpdateThread(GetCurrentThread());
    DetourDetach_IF_Enable<IfTrueCreateWindowExA>(&(PVOID&)TrueCreateWindowExA, HookedCreateWindowExA);
    DetourDetach_IF_Enable<IfTrueCreateWindowExW>(&(PVOID&)TrueCreateWindowExW, HookedCreateWindowExW);
    DetourDetach_IF_Enable<IfTrueDefWindowProcA>(&(PVOID&)TrueDefWindowProcA, HookedDefWindowProcA);
    DetourDetach_IF_Enable<IfTrueDefWindowProcW>(&(PVOID&)TrueDefWindowProcW, HookedDefWindowProcW);
    DetourDetach_IF_Enable<IfTrueCreateFontA>(&(PVOID&)TrueCreateFontA, HookedCreateFontA);
    DetourDetach_IF_Enable<IfTrueCreateFontW>(&(PVOID&)TrueCreateFontW, HookedCreateFontW);
    DetourDetach_IF_Enable<IfTrueCreateFontIndirectA>(&(PVOID&)TrueCreateFontIndirectA, HookedCreateFontIndirectA);
    DetourDetach_IF_Enable<IfTrueCreateFontIndirectW>(&(PVOID&)TrueCreateFontIndirectW, HookedCreateFontIndirectW);
    DetourDetach_IF_Enable<IfTrueCreateFontIndirectExW>(&(PVOID&)TrueCreateFontIndirectExA, HookedCreateFontIndirectExA);
    DetourDetach_IF_Enable<IfTrueCreateFontIndirectExW>(&(PVOID&)TrueCreateFontIndirectExW, HookedCreateFontIndirectExW);
    DetourDetach_IF_Enable<IfTrueEnumFontFamiliesExA>(&(PVOID&)TrueEnumFontFamiliesExA, HookedEnumFontFamiliesExA);
    DetourDetach_IF_Enable<IfTrueEnumFontFamiliesExW>(&(PVOID&)TrueEnumFontFamiliesExW, HookedEnumFontFamiliesExW);
    DetourDetach_IF_Enable<IfTrueMultiByteToWideChar>(&(PVOID&)TrueMultiByteToWideChar, HookedMultiByteToWideChar);
    DetourDetach_IF_Enable<IfTrueWideCharToMultiByte>(&(PVOID&)TrueWideCharToMultiByte, HookedWideCharToMultiByte);
	UnloadFont(MyFont.c_str());
    return true;
}

BOOL APIENTRY DllMain(HMODULE hModule,
    DWORD  ul_reason_for_call,
    LPVOID lpReserved
)
{
    switch (ul_reason_for_call)
    {
    case DLL_PROCESS_ATTACH:

#if IfAttachDebugConsole 1
        AttachDebugConsole();
#endif

        InitializeHooks();

#if IfRangeCheck 1
        PatchRangeCheck();
#endif
        break;
    case DLL_THREAD_ATTACH:
        break;
    case DLL_THREAD_DETACH:
        break;
    case DLL_PROCESS_DETACH:

        FreeHooks();

        break;
    }
    return TRUE;
}