#include "WindowHook.h"
#pragma comment(lib, "detours.lib")
using namespace WindowHookConfig;
using namespace WindowHookUtils;



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
    LPVOID lpParam) {

    WrapIsFirstCall("CreateWindowExA Hooked\n");

    static std::string ansiTitle;

    int len = WideCharToMultiByte(CP_ACP, 0, NEW_TITLE, -1, NULL, 0, NULL, NULL);
    ansiTitle.resize(len);
    WideCharToMultiByte(CP_ACP, 0, NEW_TITLE, -1, &ansiTitle[0], len, NULL, NULL);

    LPCSTR newTitle = (lpWindowName && *lpWindowName) ? ansiTitle.c_str() : lpWindowName;


    return TrueCreateWindowExA(dwExStyle, lpClassName, newTitle, dwStyle, X, Y, nWidth, nHeight, hWndParent, hMenu, hInstance, lpParam);

}

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
    LPVOID lpParam) {

    WrapIsFirstCall("CreateWindowExW Hooked\n");

    LPCWSTR newTitle = (lpWindowName && *lpWindowName) ? NEW_TITLE : lpWindowName;

    return TrueCreateWindowExW(dwExStyle, lpClassName, newTitle, dwStyle, X, Y, nWidth, nHeight, hWndParent, hMenu, hInstance, lpParam);
}



LRESULT WINAPI HookedDefWindowProcA(HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam) {

    WrapIsFirstCall("DefWindowProcA Hooked\n");

    switch (Msg) {
    case WM_NCCREATE: {
        CREATESTRUCTA* csA = reinterpret_cast<CREATESTRUCTA*>(lParam);
        if (csA) {
            CREATESTRUCTW csW{};
            memcpy(&csW, csA, sizeof(CREATESTRUCTW));
            csW.lpszName = NEW_TITLE;
            csW.lpszClass = ShiftJISToWide(csA->lpszClass).c_str();
            return TrueDefWindowProcW(hWnd, Msg, wParam, reinterpret_cast<LPARAM>(&csW));
        }
        break;
    }
    case WM_SETTEXT: {
        return TrueDefWindowProcW(hWnd, Msg, wParam, reinterpret_cast<LPARAM>(NEW_TITLE));
    }

    }

    return TrueDefWindowProcA(hWnd, Msg, wParam, lParam);
}


LRESULT WINAPI HookedDefWindowProcW(HWND hWnd, UINT Msg, WPARAM wParam, LPARAM lParam) {

    WrapIsFirstCall("DefWindowProcW Hooked\n");

    switch (Msg) {
    case WM_NCCREATE: {
        CREATESTRUCTW* cs = reinterpret_cast<CREATESTRUCTW*>(lParam);
        if (cs) {
            cs->lpszName = NEW_TITLE;
            return TrueDefWindowProcW(hWnd, Msg, wParam, lParam);
        }
        break;
    }
    case WM_SETTEXT: {
        return TrueDefWindowProcW(hWnd, Msg, wParam, reinterpret_cast<LPARAM>(NEW_TITLE));
    }
    }
    
    return TrueDefWindowProcW(hWnd, Msg, wParam, lParam);
}


void ProcessExistingWindows() {
    std::vector<HWND> windows;
    EnumWindows(EnumWindowsProc, reinterpret_cast<LPARAM>(&windows));
    for (const auto& window : windows) {
        TrueDefWindowProcW(window, WM_SETTEXT, 0, reinterpret_cast<LPARAM>(NEW_TITLE));
    }
}