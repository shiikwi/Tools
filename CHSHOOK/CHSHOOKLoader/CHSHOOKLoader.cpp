#include <Windows.h>
#include <tchar.h>
#include "detours.h"
#pragma comment(lib, "detours.lib")
#pragma comment(linker, "/SUBSYSTEM:WINDOWS /ENTRY:mainCRTStartup")

STARTUPINFO si = { sizeof(si) };
PROCESS_INFORMATION pi;

int main()
{
    LPCSTR dllName[] = { "CHSHOOKDLL.dll" };
   bool IfHookSuccess = DetourCreateProcessWithDllsW(L"HookTest.exe", NULL, NULL, NULL, FALSE, CREATE_SUSPENDED, NULL, NULL, &si, &pi, 1,dllName, NULL);
    if (IfHookSuccess)
    {
        ResumeThread(pi.hThread);
        WaitForSingleObject(pi.hProcess, INFINITE);
        CloseHandle(pi.hProcess);
        CloseHandle(pi.hThread);
    }
    else
    {
        MessageBoxW(NULL,L"HOOK ERROR", L"shikwi", MB_ICONERROR);
    }

    return 0;
}