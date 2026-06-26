#include "version.h"
#include "detours.h"

#include <windows.h>

#include <cstdint>
#include <cstdio>
#include <cstring>
#include <fstream>

#pragma comment(lib, "detours.lib")

struct UxMemoryCryptor
{
    void* pVtable;
    uint32_t uiKey;
    uint8_t aTable[1024];
};

using FnDecryptRLD = void(__thiscall*)(UxMemoryCryptor*, uint32_t, uint32_t, uint32_t, uint32_t);
using FnLoadLibraryExA = HMODULE(WINAPI*)(LPCSTR, HANDLE, DWORD);

static FnDecryptRLD g_DecryptRLD = nullptr;
static FnLoadLibraryExA g_LoadLibraryExA = ::LoadLibraryExA;

static const char* BaseNameA(const char* path)
{
    if (!path) return "";
    const char* a = strrchr(path, '\\');
    const char* b = strrchr(path, '/');
    const char* p = a > b ? a : b;
    return p ? p + 1 : path;
}

static DWORD GetModuleSize(HMODULE module)
{
    auto* dos = reinterpret_cast<IMAGE_DOS_HEADER*>(module);
    if (!dos || dos->e_magic != IMAGE_DOS_SIGNATURE) return 0;

    auto* nt = reinterpret_cast<IMAGE_NT_HEADERS32*>(
        reinterpret_cast<uint8_t*>(module) + dos->e_lfanew
    );
    if (nt->Signature != IMAGE_NT_SIGNATURE) return 0;

    return nt->OptionalHeader.SizeOfImage;
}

static uint8_t* FindBytesForward(uint8_t* begin, uint8_t* end, const uint8_t* pat, size_t patSize)
{
    if (!begin || !end || !pat || patSize == 0 || begin >= end) return nullptr;

    for (uint8_t* p = begin; p + patSize <= end; ++p)
    {
        if (memcmp(p, pat, patSize) == 0) return p;
    }
    return nullptr;
}

static uint8_t* FindBytesBackward(uint8_t* start, uint8_t* begin, const uint8_t* pat, size_t patSize)
{
    if (!start || !begin || !pat || patSize == 0 || start < begin) return nullptr;

    for (uint8_t* p = start; ; --p)
    {
        if (memcmp(p, pat, patSize) == 0) return p;
        if (p == begin) break;
    }
    return nullptr;
}

static void SaveKeyPair(const char* binName, const char* txtName, const UxMemoryCryptor* cryptor)
{
    {
        std::ofstream ofs(binName, std::ios::binary);
        ofs.write(reinterpret_cast<const char*>(cryptor->aTable), sizeof(cryptor->aTable));
    }

    {
        char buf[16] = {};
        std::snprintf(buf, sizeof(buf), "0x%08X", cryptor->uiKey);
        std::ofstream ofs(txtName, std::ios::binary);
        ofs << buf;
    }
}

static void __fastcall DecryptRLD_Hook(
    UxMemoryCryptor* cryptor,
    void*,
    uint32_t isDecode,
    uint32_t pScript,
    uint32_t szScript,
    uint32_t uiKey)
{
    static uint32_t callTimes = 0;

    if (callTimes == 0)
        SaveKeyPair("key_def.bin", "key_def.txt", cryptor);
    else if (callTimes == 1)
        SaveKeyPair("key.bin", "key.txt", cryptor);

    ++callTimes;
    g_DecryptRLD(cryptor, isDecode, pScript, szScript, uiKey);
}

static uint8_t* FindFnDecodeScript(HMODULE imageBase)
{
    auto* base = reinterpret_cast<uint8_t*>(imageBase);
    const DWORD size = GetModuleSize(imageBase);
    if (!base || size == 0) return nullptr;

    uint8_t* textBegin = base + 0x1000;
    uint8_t* textEnd = base + size;

    const uint8_t loadRldStr[] = { 0x6C, 0x6F, 0x61, 0x64, 0x52, 0x6C, 0x64, 0x00 };
    uint8_t* strLoadRld = FindBytesForward(textBegin, textEnd, loadRldStr, sizeof(loadRldStr));
    if (!strLoadRld) return nullptr;

    uint8_t pushStr[] = { 0x68, 0, 0, 0, 0 };
    memcpy(pushStr + 1, &strLoadRld, 4);

    uint8_t* pushLoadRld = FindBytesBackward(strLoadRld, textBegin, pushStr, sizeof(pushStr));
    if (!pushLoadRld) return nullptr;

    const uint8_t callOp[] = { 0xE8 };

    uint8_t* firstCall = FindBytesBackward(pushLoadRld, textBegin, callOp, sizeof(callOp));
    if (!firstCall || firstCall == textBegin) return nullptr;

    uint8_t* secondCall = FindBytesBackward(firstCall - 1, textBegin, callOp, sizeof(callOp));
    if (!secondCall) return nullptr;

    int32_t rel = *reinterpret_cast<int32_t*>(secondCall + 1);
    return secondCall + 5 + rel;
}

static HMODULE WINAPI LoadLibraryExA_Hook(LPCSTR lpLibFileName, HANDLE hFile, DWORD dwFlags)
{
    HMODULE mod = g_LoadLibraryExA(lpLibFileName, hFile, dwFlags);

    if (mod && _stricmp(BaseNameA(lpLibFileName), "resident.dll") == 0)
    {
        auto* fn = FindFnDecodeScript(mod);
        if (fn)
        {
            g_DecryptRLD = reinterpret_cast<FnDecryptRLD>(fn);

            DetourTransactionBegin();
            DetourUpdateThread(GetCurrentThread());
            DetourAttach(reinterpret_cast<PVOID*>(&g_DecryptRLD), DecryptRLD_Hook);
            DetourDetach(reinterpret_cast<PVOID*>(&g_LoadLibraryExA), LoadLibraryExA_Hook);
            DetourTransactionCommit();
        }
    }

    return mod;
}

static void InstallHooks()
{
    DetourTransactionBegin();
    DetourUpdateThread(GetCurrentThread());
    DetourAttach(reinterpret_cast<PVOID*>(&g_LoadLibraryExA), LoadLibraryExA_Hook);
    DetourTransactionCommit();
}

BOOL APIENTRY DllMain(HMODULE hModule, DWORD reason, LPVOID)
{
    if (reason == DLL_PROCESS_ATTACH)
    {
        DisableThreadLibraryCalls(hModule);
        InstallHooks();
        InitHijack();
    }
    else if(reason == DLL_PROCESS_DETACH)
    {
        FreeHijack();
    }
    return TRUE;
}