#pragma once
#include <Windows.h>
#include <iostream>
#include <print>
#include <string>
#include <format>
#include <optional>
#include <span>
#include <algorithm>
#include <vector>

namespace Utils
{
	class Logger
	{
	public:
		template<typename... Args>
		static void Log(std::format_string<Args...> fmt, Args&&... args)
		{
			std::cout << std::format("[*FkCxdec] {}\n", std::format(fmt, std::forward<Args>(args)...));
		}
		template<typename... Args>
		static void Error(std::format_string<Args...> fmt, Args&&... args)
		{
			std::cerr << std::format("[FkCxdec Error] {}\n", std::format(fmt, std::forward<Args>(args)...));
		}
		static std::string WStringToUTF8(const std::wstring& wstr) {
			if (wstr.empty()) return {};
			int size = WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, nullptr, 0, nullptr, nullptr);
			std::string result(size - 1, '\0');
			WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, result.data(), size, nullptr, nullptr);
			return result;
		}
	};

	class PatternScanner {
	public:
		struct PatternByte
		{
			uint8_t data;
			bool wildcard;
		};

		static std::vector<PatternByte> Parse(std::span<const char> bytes, std::span<const char> mask)
		{
			std::vector<PatternByte> result;
			result.reserve(bytes.size());
			for (size_t i = 0; i < bytes.size(); i++)
			{
				bool is_wildcard = (i < bytes.size()) ? (mask[i] == 0) : false;
				result.push_back({ static_cast<uint8_t>(bytes[i]), is_wildcard });
			}
			return result;
		}

		static std::optional<size_t> Find(std::span<const char> buffer, std::vector<PatternByte>& pattern)
		{
			auto it = std::search(buffer.begin(), buffer.end(),
				pattern.begin(), pattern.end(),
				[](char rawByte, const PatternByte& pat)
				{return pat.wildcard || (static_cast<uint8_t>(rawByte) == pat.data);}
			);

			if (it != buffer.end())
			{
				return std::distance(buffer.begin(), it);
			}
			return std::nullopt;
		}
	};

	class Console
	{
	public:
		static void AttachConsole()
		{
			::AllocConsole();
			FILE* f;
			freopen_s(&f, "CONOUT$", "w", stdout);
			freopen_s(&f, "CONOUT$", "w", stderr);
			SetConsoleTitleW(L"FkCxdec Debug Console");
		}
	};
}
