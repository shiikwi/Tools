#pragma once
#include "Management.h"

namespace WindowHookConfig {

	inline decltype(&CreateWindowExA) TrueCreateWindowExA = CreateWindowExA;
	inline decltype(&CreateWindowExW) TrueCreateWindowExW = CreateWindowExW;
	inline decltype(&DefWindowProcA) TrueDefWindowProcA = DefWindowProcA;
	inline decltype(&DefWindowProcW) TrueDefWindowProcW = DefWindowProcW;
}

namespace WindowHookUtils {

	static std::wstring ShiftJISToWide(const char* str) {
		if (!str) return L"NULL";
		int len = MultiByteToWideChar(932, 0, str, -1, nullptr, 0);
		std::wstring wstr(len, 0);
		MultiByteToWideChar(932, 0, str, -1, &wstr[0], len);
		wstr.pop_back();
		return wstr;
	}

	static BOOL CALLBACK EnumWindowsProc(HWND hWnd, LPARAM lParam) {
		auto* windows = reinterpret_cast<std::vector<HWND>*>(lParam);
		DWORD windowProcessId = 0;
		GetWindowThreadProcessId(hWnd, &windowProcessId);
		if (windowProcessId == GetCurrentProcessId()) {
			windows->push_back(hWnd);
		}
		return TRUE;
	}

}

HWND WINAPI HookedCreateWindowExA(DWORD dwExStyle,
	LPCSTR lpClassName,
	LPCSTR lpWindowName,
	DWORD dwStyle,
	int X, int Y,
	int nWidth,
	int nHeight,
	HWND hWndParent,
	HMENU hMenu,
	HINSTANCE hInstance,
	LPVOID lpParam);


HWND WINAPI HookedCreateWindowExW(DWORD dwExStyle, 
	LPCWSTR lpClassName, 
	LPCWSTR lpWindowName, 
	DWORD dwStyle, 
	int X, 
	int Y, 
	int nWidth, 
	int nHeight, 
	HWND hWndParent, 
	HMENU hMenu, 
	HINSTANCE hInstance, 
	LPVOID lpParam);


LRESULT WINAPI HookedDefWindowProcA(HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam);
LRESULT WINAPI HookedDefWindowProcW(HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam);
void ProcessExistingWindows();