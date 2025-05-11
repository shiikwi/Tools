#include "WindowHook.h"
#include "FontHook.h"
#pragma comment(lib, "detours.lib")
using namespace WindowHookConfig;
using namespace FontHookConfig;

extern "C" __declspec(dllexport) void DummyExport(void) {}


void AttachDebugConsole() {
    AllocConsole();
    FILE* fDummy;
    freopen_s(&fDummy, "CONOUT$", "w", stdout);
    freopen_s(&fDummy, "CONOUT$", "w", stderr);
    SetConsoleTitleW(L"Hook Debug Console");
}


bool InitializeHooks()
{
    DetourRestoreAfterWith();
    DetourTransactionBegin();
    DetourUpdateThread(GetCurrentThread());
	DetourAttach(&(PVOID&)TrueCreateWindowExA, HookedCreateWindowExA);
	DetourAttach(&(PVOID&)TrueCreateWindowExW, HookedCreateWindowExW);
	DetourAttach(&(PVOID&)TrueDefWindowProcA, HookedDefWindowProcA);
	DetourAttach(&(PVOID&)TrueDefWindowProcW, HookedDefWindowProcW);
    DetourAttach(&(PVOID&)TrueCreateFontA, HookedCreateFontA);
    DetourAttach(&(PVOID&)TrueCreateFontW, HookedCreateFontW);

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
	DetourDetach(&(PVOID&)TrueCreateWindowExA, HookedCreateWindowExA);
	DetourDetach(&(PVOID&)TrueCreateWindowExW, HookedCreateWindowExW);
    DetourDetach(&(PVOID&)TrueDefWindowProcA, HookedDefWindowProcA);
	DetourDetach(&(PVOID&)TrueDefWindowProcW, HookedDefWindowProcW);
	DetourDetach(&(PVOID&)TrueCreateFontA, HookedCreateFontA);
	DetourDetach(&(PVOID&)TrueCreateFontW, HookedCreateFontW);
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

        AttachDebugConsole();
        InitializeHooks();

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