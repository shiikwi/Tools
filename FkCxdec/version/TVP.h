#pragma once
#include "Utils.h"
#include <fstream>
#include "../detours/detours.h"
#include "../tvp/tp_stub.h"

namespace TVPHook 
{
	inline decltype(&LoadLibraryExW) TrueLoadLibraryExW = (decltype(TrueLoadLibraryExW))GetProcAddress(GetModuleHandleW(L"kernelbase.dll"), "LoadLibraryExW");
	inline decltype(&GetProcAddress) TrueGetProcAddress = GetProcAddress;
	inline HRESULT(WINAPI* TrueV2Link)(iTVPFunctionExporter*) = nullptr;

	class TVPMod
	{
	public:
		static TVPMod& Instance()
		{
			if (!s_Instance) {
				s_Instance = new TVPMod();
			}
			return *s_Instance;
		}

		void Initialize();
		void Free();
	private:
		static TVPMod* s_Instance;
		TVPMod(const TVPMod&) = delete;
		TVPMod& operator=(const TVPMod&) = delete;
		TVPMod() = default;
		~TVPMod() = default;

		static void CheckByPass(const std::wstring& path);
		static HMODULE WINAPI HookedLoadLibraryExW(LPCWSTR lpLibFileName, HANDLE hFile, DWORD dwFlags);
		static FARPROC WINAPI HookedGetProcAddress(HMODULE hModule, LPCSTR lpProcName);
		static HRESULT WINAPI ProxyV2Link(iTVPFunctionExporter* exporter);
	};
}