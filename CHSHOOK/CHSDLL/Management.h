#pragma once
#include <Windows.h>
#include "detours.h"
#include <iostream>
#include <vector>
#include <string>


static void WrapIsFirstCall(const char* message) {
    static bool IsFirstCall = true;
    if (IsFirstCall) {
        printf(message);
        IsFirstCall = false;
    }
}


static void AttachDebugConsole();