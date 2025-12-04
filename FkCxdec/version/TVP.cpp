#include "TVP.h"
namespace TVPHook
{
	TVPMod* TVPMod::s_Instance = nullptr;
	void TVPMod::Initialize()
	{
		Utils::Logger::Log("Init Hook");
		DetourTransactionBegin();
		DetourUpdateThread(GetCurrentThread());
		DetourAttach(&(PVOID&)TrueLoadLibraryExW, HookedLoadLibraryExW);
		DetourAttach(&(PVOID&)TrueGetProcAddress, HookedGetProcAddress);
		auto err = DetourTransactionCommit();
		if (err == NO_ERROR)
		{
			Utils::Logger::Log("Hooks installed success");
		}
		else
		{
			Utils::Logger::Error("Hook failed with error: {}", err);
		}
	}

	void TVPMod::Free()
	{
		DetourTransactionBegin();
		DetourUpdateThread(GetCurrentThread());
		DetourDetach(&(LPVOID&)TrueLoadLibraryExW, HookedLoadLibraryExW);
		DetourDetach(&(PVOID&)TrueGetProcAddress, HookedGetProcAddress);
		DetourTransactionCommit();
	}

	void TVPMod::CheckByPass(const std::wstring& path)
	{
		std::fstream file(path, std::ios::in | std::ios::out | std::ios::binary);
		if (!file.is_open()) return;

		file.seekg(0, std::ios::end);
		size_t fileSize = file.tellg();
		file.seekg(0, std::ios::beg);
		std::vector<char> buffer(fileSize);
		file.read(buffer.data(), fileSize);

		const char searchByte[] = {
		0x55,                         // push ebp
		0x8B, 0xEC,                   // mov ebp, esp
		0x8B, 0x4D, 0x08,             // mov ecx, dword ptr ss:[ebp+0x08]
		0x85, 0xC9,                   // test ecx, ecx
		0x74, 0x13,                   // je short 0x13
		0xFF, 0x75, 0x10,             // push dword ptr ss:[ebp+0x10]
		0xFF, 0x75, 0x0C,             // push dword ptr ss:[ebp+0x0C]
		0xE8, 0xFB, 0x08, 0x00, 0x00, // call 0x0000C0F0
		0x84, 0xC0,                   // test al, al
		0x74, 0x04,                   // je short 0x04
		0xB0, 0x01,                   // mov al, 0x01
		0x5D,                         // pop ebp
		0xC3,                         // ret
		0x32, 0xC0,                   // xor al, al
		0x5D,                         // pop ebp
		0xC3                          // ret
		};

		const char Mask[] = {
		1,
		1, 1,
		1, 1, 1,
		1, 1,
		1, 1,
		1, 1, 1,
		1, 1, 1,
		1, 0, 0, 0, 0,
		1, 1,
		1, 1,
		1, 1,
		1,
		1,
		1, 1,
		1,
		1
		};

		const char patchByte[] = {
		0x55,                          // push ebp
		0x8B, 0xEC,                    // mov ebp, esp
		0x8B, 0x4D, 0x08,              // mov ecx, dword ptr ss:[ebp+0x08]
		0x85, 0xC9,                    // test ecx, ecx
		0xEB, 0x13,                    // jmp short 0x13 <===== Force Jump
		0xFF, 0x75, 0x10,              // push dword ptr ss:[ebp+0x10]
		0xFF, 0x75, 0x0C,              // push dword ptr ss:[ebp+0x0C]
		0xE8, 0x3B, 0x0F, 0x00, 0x00,  // call 0x0000C0F0
		0x84, 0xC0,                    // test al, al
		0x74, 0x04,                    // je short 0x04
		0xB0, 0x01,                    // mov al, 0x01
		0x5D,                          // pop ebp
		0xC3,                          // ret
		0xB0, 0x01,                    // mov al, 0x01   <===== Force bypass
		0x5D,                          // pop ebp
		0xC3                           // ret
		};

		auto pattern = Utils::PatternScanner::Parse(searchByte, Mask);
		auto offsetOpt = Utils::PatternScanner::Find(buffer, pattern);

		if (offsetOpt)
		{
			size_t offset = *offsetOpt;
			Utils::Logger::Log("Verification Pattern Match at offset {:#x}", offset);

			if (offset + sizeof(patchByte) <= buffer.size())
			{
				std::memcpy(buffer.data() + offset, patchByte, sizeof(patchByte));
				file.seekg(0, std::ios::beg);
				file.write(buffer.data(), buffer.size());
				Utils::Logger::Log("Cxdec Check ByPass Success");
			}
		}
		else
		{
			Utils::Logger::Error("Cxdec Check ByPass Failed: Not Find Verification Pattern");
		}
	}


	HMODULE WINAPI TVPMod::HookedLoadLibraryExW(LPCWSTR lpLibFileName, HANDLE hFile, DWORD dwFlags)
	{
		if (lpLibFileName)
		{
			try
			{
				std::wstring path(lpLibFileName);
				std::transform(path.begin(), path.end(), path.begin(), ::tolower);
				if (path.find(L"appdata\\local\\temp") != std::wstring::npos)
				{
					Utils::Logger::Log("Intercepted temp DLL: {}", Utils::Logger::WStringToUTF8(path));
					CheckByPass(path);
				}
			}
			catch(...){}
		}
		return TrueLoadLibraryExW(lpLibFileName, hFile, dwFlags);
	}

	FARPROC WINAPI TVPMod::HookedGetProcAddress(HMODULE hModule, LPCSTR lpProcName)
	{
		auto result = TrueGetProcAddress(hModule, lpProcName);
		if (result && HIWORD(lpProcName) != 0)
		{
			if (strcmp(lpProcName, "V2Link") == 0)
			{
				Utils::Logger::Log("V2Link export access detected");
				TrueV2Link = decltype(TrueV2Link)(result);
				return (FARPROC)ProxyV2Link;
			}
		}
		return result;
	}

	HRESULT WINAPI TVPMod::ProxyV2Link(iTVPFunctionExporter* exporter)
	{
		TVPInitImportStub(exporter);
		TVPSetCommandLine(L"-debugwin", L"yes");
		Utils::Logger::Log("Init TVP SDK Success. Enabled debugwin");

		HRESULT hr = S_OK;
		if (TrueV2Link)
		{
			hr = TrueV2Link(exporter);
		}

		return hr;
	}
}
