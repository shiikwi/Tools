#include <windows.h>
#include <iostream>
#include <fstream>
#include <vector>
#include <io.h>
#include <fcntl.h>

bool IsSJISFile(const std::vector<char>& data) {
    int wideLen = MultiByteToWideChar(932, MB_ERR_INVALID_CHARS, 
                                    data.data(), (int)data.size(), 
                                    nullptr, 0);
    return wideLen > 0;
}


bool SJIS2UTF8(const std::vector<char>& sjisData, std::vector<char>& utf8Data) {
    int wideLen = MultiByteToWideChar(932, 0, sjisData.data(), sjisData.size(), NULL, 0);
    std::vector<wchar_t> utf16(wideLen);
    MultiByteToWideChar(932, 0, sjisData.data(), sjisData.size(), utf16.data(), wideLen);
    
    int utf8Len = WideCharToMultiByte(CP_UTF8, 0, utf16.data(), wideLen, NULL, 0, NULL, NULL);
    utf8Data.resize(utf8Len);
    WideCharToMultiByte(CP_UTF8, 0, utf16.data(), wideLen, utf8Data.data(), utf8Len, NULL, NULL);
    
    return true;
}


bool ConvertFileToUTF8(const std::string& filename) {
    std::ifstream inFile(filename, std::ios::binary);
    std::vector<char> content((std::istreambuf_iterator<char>(inFile)), 
                             std::istreambuf_iterator<char>());
    inFile.close();

    if (!IsSJISFile(content)) {
        std::cout << "跳过: " << filename << std::endl;
        return false;
    }


    std::vector<char> output;
    if (!SJIS2UTF8(content, output)) return false;


    std::ofstream outFile(filename, std::ios::binary | std::ios::trunc);
    outFile.write(output.data(), output.size());
    return outFile.good();
}


int main() {

	SetConsoleOutputCP(CP_UTF8);

    struct _finddata_t fileinfo;
    intptr_t handle = _findfirst("*.txt", &fileinfo);
    if (handle == -1) return 1;
    
    int count = 0;
    do {
        if (!(fileinfo.attrib & _A_SUBDIR)) {
            std::cout << "正在处理: " << fileinfo.name << std::endl;

            if (ConvertFileToUTF8(fileinfo.name)) {
                std::cout << "转换成功: " << fileinfo.name << std::endl;
                count++;
            }
            
        }

    } while (_findnext(handle, &fileinfo) == 0);
    
    _findclose(handle);

    std::cout << "转换完成, 共处理文件数 " << count << std::endl;
    system("pause");

    return 0;
}
