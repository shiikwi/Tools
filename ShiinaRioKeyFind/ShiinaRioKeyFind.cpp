#include <iostream>
#include <vector>
#include <fstream>
#include <algorithm>
#include <cstdint>
#include <iomanip>

void GenKeyFile(char* exefile, char* pngfile, char* outfile) {
	std::ifstream png(pngfile, std::ios::binary);

	if (!png) {
		std::cout << "Failed to open PNG file: " << pngfile << std::endl;
		return;
	}

	std::vector<uint8_t> pngdata((std::istreambuf_iterator<char>(png)), {});
	png.close();


	std::ifstream exe(exefile, std::ios::binary);
	if (!exe) {
		std::cout << "Drag exe to this program" << std::endl;
		return;
	}
	std::vector<uint8_t> exedata((std::istreambuf_iterator<char>(exe)), {});
	exe.close();


	auto it = std::search(exedata.begin(), exedata.end(), pngdata.begin(), pngdata.end());
	if (it == exedata.end()) {
		std::cout << "Error: Cannot find key" << std::endl;
		return;
	}

	size_t pnghead = std::distance(exedata.begin(), it);
	size_t pos = pnghead + pngdata.size() + 4;

	std::vector<uint32_t> keys;
	for (int i = 0; i < 5; i++) {
		uint32_t val = static_cast<uint32_t>(exedata[pos]) |
			static_cast<uint32_t>(exedata[pos + 1] << 8) |
			static_cast<uint32_t>(exedata[pos + 2] << 16) |
			static_cast<uint32_t>(exedata[pos + 3]) << 24;
		keys.push_back(val);
		pos += 4;
	}

	std::ofstream out(outfile);

	for (size_t i = 0; i < keys.size(); i++) {
		out << "key" << (i + 1) << " = 0x"
			<< std::uppercase << std::hex << std::setw(8) << std::setfill('0') << keys[i]
			<< "\n";
	}

	out << '\n';
	for (size_t i = 0; i < keys.size(); i++) {
		out << "0x"
			<< std::uppercase << std::hex << std::setw(8) << std::setfill('0') << keys[i]
			<< ", ";
	}
	out.close();

	std::cout << "Key file has generated" << std::endl;
}


/*
00475A10 | B8 62BF0000       | mov eax,BF62                                |
00475A15 | 25 FFFF0000       | and eax,FFFF                                |
00475A1A | C3                | ret                                         |
*/
int findpattern(const std::vector<uint8_t>& exeData, uint32_t& outSize) {
	for (size_t i = 0; i + 11 < exeData.size(); i++) {
		if (exeData[i] == 0xB8 &&
			exeData[i + 5] == 0x25 && exeData[i + 6] == 0xFF && exeData[i + 7] == 0xFF && exeData[i + 8] == 0x00 && exeData[i + 9] == 0x00 &&
			exeData[i + 10] == 0xC3) {

			outSize = exeData[i + 1] | (exeData[i + 2] << 8) | (exeData[i + 3] << 16) | (exeData[i + 4] << 24);
			return static_cast<int>(i);
		}
	}
	return -1;
}


void GenExtra(char* exefile, char* jpgfile, char* outfile) {
	std::ifstream jpg(jpgfile, std::ios::binary);

	if (!jpg) {
		std::cout << "Failed to open JPG file: " << jpgfile << "\n";
		return;
	}

	std::vector<uint8_t> jpgdata((std::istreambuf_iterator<char>(jpg)), {});
	jpg.close();


	std::ifstream exe(exefile, std::ios::binary);
	if (!exe) {
		std::cout << "Drag exe to this program" << std::endl;
		return;
	}
	std::vector<uint8_t> exedata((std::istreambuf_iterator<char>(exe)), {});
	exe.close();

	auto it = std::search(exedata.begin(), exedata.end(), jpgdata.begin(), jpgdata.end());
	if (it == exedata.end()) {
		std::cout << "Error: Cannot find jpgdata in exe" << std::endl;
		return;
	}

	size_t jpghead = std::distance(exedata.begin(), it);
	size_t pos = jpghead + jpgdata.size();

	uint32_t imgsize = 0;
	int movpos = findpattern(exedata, imgsize);
	if (movpos < 0) {
		std::cout << "Error: Cannot find datasize" << std::endl;
		return;
	}

	std::ofstream out(outfile, std::ios::binary);

	out.write(reinterpret_cast<const char*>(&exedata[pos]), imgsize - jpgdata.size());
	out.close();

	std::cout << "Extra file has generated" << std::endl;
}


int main(int argc, char* args[])
{
	char* exefile = args[1];
	char* pngfile = (char*)"RioShiina2Key.png";
	char* jpgfile = (char*)"RioShiinaExtra.jpg";

	GenKeyFile(exefile, pngfile, (char*)"Key.txt");
	GenExtra(exefile, jpgfile, (char*)"Extra.bin");

	return 0;
}
