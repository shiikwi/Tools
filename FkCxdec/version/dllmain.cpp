#include "TVP.h"
#include "../Hijack/version.h"
#pragma comment(lib, "../detours/detours.lib")
extern "C" __declspec(dllexport) void TVP() {}

BOOL APIENTRY DllMain(HMODULE hModule,
	DWORD  ul_reason_for_call,
	LPVOID lpReserved
)
{
	switch (ul_reason_for_call)
	{
	case DLL_PROCESS_ATTACH:
	{
		//Utils::Console::AttachConsole();
		TVPHook::TVPMod::Instance().Initialize();
		InitHijack();
		break;
	}
	case DLL_THREAD_ATTACH:
	case DLL_THREAD_DETACH:
		break;
	case DLL_PROCESS_DETACH:
	{
		TVPHook::TVPMod::Instance().Free();
		FreeHijack();
		break;
	}
	}
	return TRUE;
}

