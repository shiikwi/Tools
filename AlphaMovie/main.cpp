#include "miniz.h"
#include <algorithm>
#include <array>
#include <cctype>
#include <cerrno>
#include <cstdint>
#include <cstring>
#include <filesystem>
#include <fstream>
#include <iomanip>
#include <iostream>
#include <process.h>
#include <sstream>
#include <stdexcept>
#include <string>
#include <vector>

namespace fs = std::filesystem;

namespace AmvDec {
constexpr uint32_t AmvMagic = 0x4d504a41;
constexpr uint32_t FramMagic = 0x4d415246;

constexpr int Zigzag[64] = {0,  1,  8,  16, 9,  2,  3,  10, 17, 24, 32, 25, 18,
                            11, 4,  5,  12, 19, 26, 33, 40, 48, 41, 34, 27, 20,
                            13, 6,  7,  14, 21, 28, 35, 42, 49, 56, 57, 50, 43,
                            36, 29, 22, 15, 23, 30, 37, 44, 51, 58, 59, 52, 45,
                            38, 31, 39, 46, 53, 60, 61, 54, 47, 55, 62, 63};

constexpr uint8_t StdDcLumaBits[16] = {0, 1, 5, 1, 1, 1, 1, 1,
                                       1, 0, 0, 0, 0, 0, 0, 0};
constexpr uint8_t StdDcLumaVal[12] = {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11};
constexpr uint8_t StdDcChromaBits[16] = {0, 3, 1, 1, 1, 1, 1, 1,
                                         1, 1, 1, 0, 0, 0, 0, 0};
constexpr uint8_t StdDcChromaVal[12] = {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11};
constexpr uint8_t StdAcLumaBits[16] = {0, 2, 1, 3, 3, 2, 4, 3,
                                       5, 5, 4, 4, 0, 0, 1, 0x7d};
constexpr uint8_t StdAcLumaVal[162] = {
    0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
    0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xa1, 0x08,
    0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62, 0x72,
    0x82, 0x09, 0x0a, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x25, 0x26, 0x27, 0x28,
    0x29, 0x2a, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45,
    0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
    0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74, 0x75,
    0x76, 0x77, 0x78, 0x79, 0x7a, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
    0x8a, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3,
    0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6,
    0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9,
    0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe1, 0xe2,
    0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf1, 0xf2, 0xf3, 0xf4,
    0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa};
constexpr uint8_t StdAcChromaBits[16] = {0, 2, 1, 2, 4, 4, 3, 4,
                                         7, 5, 4, 4, 0, 1, 2, 0x77};
constexpr uint8_t StdAcChromaVal[162] = {
    0x00, 0x01, 0x02, 0x03, 0x11, 0x04, 0x05, 0x21, 0x31, 0x06, 0x12, 0x41,
    0x51, 0x07, 0x61, 0x71, 0x13, 0x22, 0x32, 0x81, 0x08, 0x14, 0x42, 0x91,
    0xa1, 0xb1, 0xc1, 0x09, 0x23, 0x33, 0x52, 0xf0, 0x15, 0x62, 0x72, 0xd1,
    0x0a, 0x16, 0x24, 0x34, 0xe1, 0x25, 0xf1, 0x17, 0x18, 0x19, 0x1a, 0x26,
    0x27, 0x28, 0x29, 0x2a, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44,
    0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58,
    0x59, 0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74,
    0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87,
    0x88, 0x89, 0x8a, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a,
    0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xb2, 0xb3, 0xb4,
    0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7,
    0xc8, 0xc9, 0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda,
    0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf2, 0xf3, 0xf4,
    0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa};

inline uint8_t Clip8(int v) {
  return (uint8_t)(v < 0 ? 0 : (v > 255 ? 255 : v));
}

inline uint16_t ReadU16(const std::vector<uint8_t> &data, size_t pos) {
  if (pos + 2 > data.size())
    throw std::runtime_error("unexpected end of file");
  return (uint16_t)(data[pos] | (data[pos + 1] << 8));
}

inline int16_t ReadS16(const std::vector<uint8_t> &data, size_t pos) {
  return (int16_t)ReadU16(data, pos);
}

inline uint32_t ReadU32(const std::vector<uint8_t> &data, size_t pos) {
  if (pos + 4 > data.size())
    throw std::runtime_error("unexpected end of file");
  return (uint32_t)(data[pos] | (data[pos + 1] << 8) | (data[pos + 2] << 16) |
                    ((uint32_t)data[pos + 3] << 24));
}

inline void RequireRange(const std::vector<uint8_t> &data, size_t pos,
                         size_t size, const char *what) {
  if (pos > data.size() || size > data.size() - pos)
    throw std::runtime_error(std::string("truncated ") + what);
}

struct HuffTable {
  int mincode[17] = {};
  int maxcode[18] = {};
  int valptr[17] = {};
  uint8_t huffval[256] = {};

  void Build(const uint8_t *bits16, const uint8_t *vals) {
    int huffsize[257];
    int huffcode[257];
    int k = 0;
    for (int l = 1; l <= 16; l++)
      for (int i = 0; i < bits16[l - 1]; i++)
        huffsize[k++] = l;
    huffsize[k] = 0;
    const int symbolCount = k;

    int code = 0;
    int codeLen = huffsize[0];
    k = 0;
    while (huffsize[k]) {
      while (huffsize[k] == codeLen) {
        huffcode[k] = code;
        code++;
        k++;
      }
      code <<= 1;
      codeLen++;
    }
    int p = 0;
    for (int l = 1; l <= 16; l++) {
      if (bits16[l - 1]) {
        valptr[l] = p;
        mincode[l] = huffcode[p];
        p += bits16[l - 1];
        maxcode[l] = huffcode[p - 1];
      } else {
        maxcode[l] = -1;
      }
    }
    maxcode[17] = 0x7fffffff;
    std::memcpy(huffval, vals, symbolCount);
  }
};

class BitReader {
public:
  BitReader(const uint8_t *data, size_t size)
      : m_cur(data), m_end(data + size) {
    if (size >= 4) {
      m_reservoir = ((uint32_t)data[0] << 24) | ((uint32_t)data[1] << 16) |
                    ((uint32_t)data[2] << 8) | (uint32_t)data[3];
      m_cur = data + 4;
    } else {
      for (size_t i = 0; i < size; ++i)
        m_reservoir |= (uint32_t)data[i] << (24 - 8 * i);
      m_cur = m_end;
    }
  }

  uint32_t Bits(int n) {
    if (n <= 0)
      return 0;
    if (n > 24)
      throw std::runtime_error("AlphaMovie bit request is too large");
    m_consumed += (uint32_t)n;
    const uint32_t mask = 0xffffffffu << (32 - n);
    uint32_t result;
    if (m_consumed >= 32) {
      const uint32_t left = m_consumed - 32;
      uint32_t next = 0;
      if (m_cur < m_end) {
        const size_t avail = (size_t)(m_end - m_cur);
        if (avail >= 4)
          next = ((uint32_t)m_cur[0] << 24) | ((uint32_t)m_cur[1] << 16) |
                 ((uint32_t)m_cur[2] << 8) | (uint32_t)m_cur[3];
        else
          for (size_t i = 0; i < avail; ++i)
            next |= (uint32_t)m_cur[i] << (24 - 8 * i);
        m_cur += std::min<size_t>(4, avail);
      }
      if (left)
        m_reservoir |= next >> (n - left);
      result = (m_reservoir & mask) >> (32 - n);
      m_reservoir = next << left;
      m_consumed = left;
    } else {
      result = (m_reservoir & mask) >> (32 - n);
      m_reservoir <<= n;
    }
    return result;
  }

  int Bit() { return (int)Bits(1); }

  int Decode(const HuffTable &table) {
    int length = 1;
    int code = Bit();
    while (length <= 16 && code > table.maxcode[length]) {
      code = (code << 1) | Bit();
      ++length;
    }
    if (length > 16)
      throw std::runtime_error("invalid AlphaMovie Huffman code");
    return table.huffval[table.valptr[length] + code - table.mincode[length]];
  }

private:
  const uint8_t *m_cur = nullptr;
  const uint8_t *m_end = nullptr;
  uint32_t m_reservoir = 0;
  uint32_t m_consumed = 0;
};

inline int ReceiveExtend(BitReader &reader, int size) {
  if (size == 0)
    return 0;
  int value = (int)reader.Bits(size);
  if (value < (1 << (size - 1)))
    value -= (1 << size) - 1;
  return value;
}

void Idct8x8Amv(const int16_t *coef, const uint8_t *quant, uint8_t *out) {
  int32_t t[64]{};
  auto q = [&](int i) -> int64_t { return (int64_t)coef[i] * quant[i]; };
  auto sar = [](int64_t v, int n) -> int32_t { return (int32_t)(v >> n); };
  for (int col = 0; col < 8; ++col) {
    bool nonzero = false;
    for (int row = 1; row < 8; ++row)
      nonzero |= coef[row * 8 + col] != 0;
    if (!nonzero) {
      const int32_t v = (int32_t)(4 * q(col));
      for (int row = 0; row < 8; ++row)
        t[row * 8 + col] = v;
      continue;
    }
    const int64_t c0 = q(col), c1 = q(8 + col), c2 = q(16 + col),
                  c3 = q(24 + col);
    const int64_t c4 = q(32 + col), c5 = q(40 + col), c6 = q(48 + col),
                  c7 = q(56 + col);
    const int64_t v9 = 4433 * (c6 + c2);
    const int64_t v10 = v9 + 6270 * c2;
    const int64_t v11 = v9 - 15137 * c6;
    const int64_t v14 = (c4 + c0) << 13;
    const int64_t v15 = (c0 - c4) << 13;
    const int64_t v69 = v14 + v10;
    const int64_t v67 = v15 + v11;
    const int64_t v65 = v15 - v11;
    const int64_t v61 = v14 - v10;
    const int64_t v18 = c7 + c3;
    const int64_t v19 = c1 + c5;
    const int64_t v20 = 9633 * (c1 + c5 + c7 + c3);
    const int64_t v21 = -7373 * (c7 + c1);
    const int64_t v22 = -20995 * (c3 + c5);
    const int64_t v23 = v20 - 16069 * v18;
    const int64_t v24 = v20 - 3196 * v19;
    const int64_t v25 = v22 + v23 + 25172 * c3;
    const int64_t v63 = v23 + v21 + 2446 * c7;
    const int64_t v26 = v21 + v24 + 12299 * c1;
    const int64_t v71 = v22 + v24 + 16819 * c5;
    t[col + 0] = sar(v26 + v69 + 1024, 11);
    t[col + 56] = sar(v69 - v26 + 1024, 11);
    t[col + 48] = sar(v67 - v25 + 1024, 11);
    t[col + 8] = sar(v25 + v67 + 1024, 11);
    t[col + 40] = sar(v65 - v71 + 1024, 11);
    t[col + 16] = sar(v65 + v71 + 1024, 11);
    t[col + 32] = sar(v61 - v63 + 1024, 11);
    t[col + 24] = sar(v61 + v63 + 1024, 11);
  }
  for (int row = 0; row < 8; ++row) {
    const int32_t *r = t + row * 8;
    if (r[1] || r[2] || r[3] || r[4] || r[5] || r[6] || r[7]) {
      const int64_t v31 = r[2];
      const int64_t v32 = r[6];
      const int64_t v33 = 4433 * (v32 + v31);
      const int64_t v34 = v33 - 15137 * v32;
      const int64_t v35 = v33 + 6270 * v31;
      const int64_t v37 = ((int64_t)r[0] + r[4]) << 13;
      const int64_t v38 = ((int64_t)r[0] - r[4]) << 13;
      const int64_t v39 = v37 + v35;
      const int64_t v40 = v37 - v35;
      const int64_t v64 = v38 + v34;
      const int64_t v41 = v38 - v34;
      const int64_t v42 = r[7];
      const int64_t v43 = r[3];
      const int64_t v44 = r[5];
      const int64_t v45 = v44 + r[1];
      const int64_t v81 = 9633 * (v44 + r[1] + v43);
      const int64_t v46 = -20995 * (v44 + v43);
      const int64_t v47 = v81 - 16069 * v43;
      const int64_t v82 = v81 - 3196 * v45;
      const int64_t v48 = -7373 * (v42 + r[1]);
      const int64_t v70 = v47 + v48 + 2446 * v42;
      const int64_t v73 = v46 + v82 + 16819 * v44;
      const int64_t v49 = v46 + v47 + 25172 * v43;
      const int64_t v50 = v48 + v82 + 12299 * r[1];
      auto pix = [&](int64_t v) -> uint8_t {
        return Clip8(sar(v + 0x20000, 18) + 128);
      };
      const int64_t p0 = v50 + v39;
      const int64_t p1 = v39 - v50;
      const int64_t p2 = v49 + v64;
      const int64_t p3 = v64 - v49;
      const int64_t p4 = v41 + v73;
      const int64_t p5 = v41 - v73;
      const int64_t p6 = v40 + v70;
      const int64_t p7 = v40 - v70;
      out[row * 8 + 0] = pix(p0);
      out[row * 8 + 1] = pix(p2);
      out[row * 8 + 2] = pix(p4);
      out[row * 8 + 3] = pix(p6);
      out[row * 8 + 4] = pix(p7);
      out[row * 8 + 5] = pix(p5);
      out[row * 8 + 6] = pix(p3);
      out[row * 8 + 7] = pix(p1);
    } else {
      const uint8_t v = Clip8(sar((int64_t)r[0] + 16, 5) + 128);
      for (int x = 0; x < 8; ++x)
        out[row * 8 + x] = v;
    }
  }
}
} // namespace AmvDec

struct DecodedFrame {
  int left = 0;
  int top = 0;
  int width = 0;
  int height = 0;
  std::vector<uint8_t> rgba;
};

class AmvMovie {
public:
  void Open(const fs::path &path) {
    std::ifstream file(path, std::ios::binary);
    if (!file)
      throw std::runtime_error("cannot open: " + path.string());
    file.seekg(0, std::ios::end);
    const auto size = file.tellg();
    if (size < 0)
      throw std::runtime_error("cannot determine file size");
    m_data.resize((size_t)size);
    file.seekg(0, std::ios::beg);
    if (!m_data.empty())
      file.read((char *)m_data.data(), (std::streamsize)m_data.size());
    if (!file && !m_data.empty())
      throw std::runtime_error("read failed: " + path.string());
    ParseHeader();
    ScanFrames();
  }

  uint16_t Width() const { return m_width; }
  uint16_t Height() const { return m_height; }
  uint32_t FrameCount() const { return m_frameCount; }
  double Fps() const { return (double)m_fpsRate / (double)m_fpsScale; }

  void DecodeFrame(size_t index, DecodedFrame &out) const {
    if (index >= m_frameOffsets.size())
      throw std::runtime_error("frame index out of range");
    DecodeZlibAlpha(m_frameOffsets[index], out);
  }

private:
  void ParseHeader() {
    AmvDec::RequireRange(m_data, 0, 0x28, "AMV header");
    if (AmvDec::ReadU32(m_data, 0) != AmvDec::AmvMagic)
      throw std::runtime_error("not an AlphaMovie file (missing AJPM)");
    if (AmvDec::ReadU32(m_data, 8) != 0)
      throw std::runtime_error("unsupported AlphaMovie revision");
    const uint32_t headerSize = AmvDec::ReadU32(m_data, 0x0c);
    if (headerSize < 0x28 || headerSize > m_data.size())
      throw std::runtime_error("invalid header size");
    m_quantSize = headerSize - 0x28;
    if (m_quantSize != 0x80)
      throw std::runtime_error("invalid quantization table size");
    m_frameCount = AmvDec::ReadU32(m_data, 0x14);
    m_fpsScale = AmvDec::ReadU32(m_data, 0x18);
    m_fpsRate = AmvDec::ReadU32(m_data, 0x1c);
    m_width = AmvDec::ReadU16(m_data, 0x20);
    m_height = AmvDec::ReadU16(m_data, 0x22);
    if (!m_frameCount || !m_fpsScale || !m_fpsRate || !m_width || !m_height)
      throw std::runtime_error(
          "invalid frame count, frame rate, or dimensions");
    const uint32_t flags = AmvDec::ReadU32(m_data, 0x24);
    if (flags != 2)
      throw std::runtime_error("unsupported AlphaMovie attributes");
    AmvDec::RequireRange(m_data, 0x28, m_quantSize, "quantization tables");
    std::memcpy(m_quant[0], m_data.data() + 0x28, 64);
    std::memcpy(m_quant[1], m_data.data() + 0x28 + 64, 64);
    m_dcLuma.Build(AmvDec::StdDcLumaBits, AmvDec::StdDcLumaVal);
    m_dcChroma.Build(AmvDec::StdDcChromaBits, AmvDec::StdDcChromaVal);
    m_acLuma.Build(AmvDec::StdAcLumaBits, AmvDec::StdAcLumaVal);
    m_acChroma.Build(AmvDec::StdAcChromaBits, AmvDec::StdAcChromaVal);
  }

  void ScanFrames() {
    m_frameOffsets.clear();
    size_t ofs = 0x28 + m_quantSize;
    for (uint32_t i = 0; i < m_frameCount; ++i) {
      AmvDec::RequireRange(m_data, ofs, 24, "frame header");
      if (AmvDec::ReadU32(m_data, ofs) != AmvDec::FramMagic)
        throw std::runtime_error("invalid FRAM marker at frame " +
                                 std::to_string(i));
      const uint32_t size = AmvDec::ReadU32(m_data, ofs + 4);
      if (size < 16)
        throw std::runtime_error("invalid frame size at frame " +
                                 std::to_string(i));
      AmvDec::RequireRange(m_data, ofs + 8, size, "frame payload");
      m_frameOffsets.push_back(ofs);
      ofs += (size_t)size + 8;
    }
  }

  void DecodeZlibAlpha(size_t ofs, DecodedFrame &out) const {
    AmvDec::RequireRange(m_data, ofs, 24, "zlib-alpha frame header");
    const uint32_t size = AmvDec::ReadU32(m_data, ofs + 4);
    out.left = AmvDec::ReadS16(m_data, ofs + 12);
    out.top = AmvDec::ReadS16(m_data, ofs + 14);
    out.width = AmvDec::ReadU16(m_data, ofs + 16);
    out.height = AmvDec::ReadU16(m_data, ofs + 18);
    const uint32_t alphaSize = AmvDec::ReadU32(m_data, ofs + 20);
    if ((out.width & 15) || (out.height & 15) || !out.width || !out.height ||
        size < 16 || alphaSize > size - 16)
      throw std::runtime_error("invalid zlib-alpha frame dimensions or sizes");
    const size_t colorSize = (size_t)size - alphaSize - 16;
    AmvDec::RequireRange(m_data, ofs + 24, (size_t)alphaSize + colorSize,
                         "zlib-alpha frame payload");

    std::vector<uint8_t> alpha((size_t)out.width * out.height, 255);
    mz_ulong alphaOut = (mz_ulong)alpha.size();
    if (mz_uncompress(alpha.data(), &alphaOut, m_data.data() + ofs + 24,
                      alphaSize) != MZ_OK ||
        alphaOut != alpha.size())
      throw std::runtime_error("invalid zlib alpha stream");

    AmvDec::BitReader reader(m_data.data() + ofs + 24 + alphaSize, colorSize);
    int dcPred = 0;
    int lumaPred = 0;
    constexpr size_t kCb = 0;
    constexpr size_t kCr = 1;
    constexpr size_t kY0 = 2;
    constexpr size_t kY1 = 3;
    constexpr size_t kY2 = 4;
    constexpr size_t kY3 = 5;
    std::array<std::array<uint8_t, 64>, 6> blocks{};
    std::array<int16_t, 64> coeff{};
    auto DecodeBlock = [&](const AmvDec::HuffTable &dc,
                           const AmvDec::HuffTable &ac, const uint8_t *qt,
                           int &pred, std::array<uint8_t, 64> &dst) {
      coeff.fill(0);
      const int category = reader.Decode(dc);
      pred += AmvDec::ReceiveExtend(reader, category);
      coeff[0] = (int16_t)pred;
      int k = 1;
      while (k < 64) {
        const int rs = reader.Decode(ac);
        const int run = rs >> 4;
        const int bits = rs & 15;
        if (bits == 0) {
          if (run == 15) {
            k += 16;
            continue;
          }
          break;
        }
        k += run;
        if (k >= 64)
          throw std::runtime_error("invalid AlphaMovie AC run");
        coeff[AmvDec::Zigzag[k]] = (int16_t)AmvDec::ReceiveExtend(reader, bits);
        ++k;
      }
      AmvDec::Idct8x8Amv(coeff.data(), qt, dst.data());
    };

    out.rgba.assign((size_t)out.width * out.height * 4, 0);
    const int mbw = out.width / 16;
    const int mbh = out.height / 16;
    for (int my = 0; my < mbh; ++my) {
      for (int mx = 0; mx < mbw; ++mx) {
        DecodeBlock(m_dcChroma, m_acChroma, m_quant[1], dcPred, blocks[kCb]);
        DecodeBlock(m_dcChroma, m_acChroma, m_quant[1], dcPred, blocks[kCr]);
        DecodeBlock(m_dcLuma, m_acLuma, m_quant[0], lumaPred, blocks[kY0]);
        DecodeBlock(m_dcLuma, m_acLuma, m_quant[0], lumaPred, blocks[kY1]);
        DecodeBlock(m_dcLuma, m_acLuma, m_quant[0], lumaPred, blocks[kY2]);
        DecodeBlock(m_dcLuma, m_acLuma, m_quant[0], lumaPred, blocks[kY3]);
        for (int y = 0; y < 16; ++y) {
          for (int x = 0; x < 16; ++x) {
            const size_t yBlock = kY0 + (y < 8 ? 0 : 2) + (x < 8 ? 0 : 1);
            const uint8_t yy = blocks[yBlock][(y & 7) * 8 + (x & 7)];
            const int cb = blocks[kCb][(y >> 1) * 8 + (x >> 1)] - 128;
            const int cr = blocks[kCr][(y >> 1) * 8 + (x >> 1)] - 128;
            const size_t pos =
                ((size_t)(my * 16 + y) * out.width + mx * 16 + x) * 4;
            out.rgba[pos + 0] = AmvDec::Clip8((int)yy + ((91881 * cr) >> 16));
            out.rgba[pos + 1] =
                AmvDec::Clip8((int)yy - ((22554 * cb + 46802 * cr) >> 16));
            out.rgba[pos + 2] = AmvDec::Clip8((int)yy + ((116130 * cb) >> 16));
            out.rgba[pos + 3] =
                alpha[(size_t)(my * 16 + y) * out.width + mx * 16 + x];
          }
        }
      }
    }
  }

  std::vector<uint8_t> m_data;
  std::vector<size_t> m_frameOffsets;
  uint32_t m_frameCount = 0;
  uint32_t m_fpsScale = 0;
  uint32_t m_fpsRate = 0;
  uint16_t m_width = 0;
  uint16_t m_height = 0;
  uint32_t m_quantSize = 0;
  uint8_t m_quant[2][64] = {};
  AmvDec::HuffTable m_dcLuma;
  AmvDec::HuffTable m_dcChroma;
  AmvDec::HuffTable m_acLuma;
  AmvDec::HuffTable m_acChroma;
};

namespace Png {
class Writer {
public:
  Writer(int width, int height) : m_width(width), m_height(height) {
    const size_t rawSize = (size_t)m_height * ((size_t)m_width * 4 + 1);
    m_raw.resize(rawSize);
    m_packed.resize(mz_compressBound((mz_ulong)rawSize));
    PutBE32(m_ihdr, (uint32_t)m_width);
    PutBE32(m_ihdr, (uint32_t)m_height);
    m_ihdr.insert(m_ihdr.end(), {8, 6, 0, 0, 0});
  }

  void Write(const fs::path &path, const std::vector<uint8_t> &rgba) {
    for (int y = 0; y < m_height; ++y) {
      uint8_t *row = m_raw.data() + (size_t)y * ((size_t)m_width * 4 + 1);
      row[0] = 0;
      std::memcpy(row + 1, rgba.data() + (size_t)y * m_width * 4,
                  (size_t)m_width * 4);
    }
    mz_ulong packedSize = (mz_ulong)m_packed.size();
    if (mz_compress2(m_packed.data(), &packedSize, m_raw.data(),
                     (mz_ulong)m_raw.size(), MZ_BEST_SPEED) != MZ_OK)
      throw std::runtime_error("PNG compression failed");
    std::ofstream file(path, std::ios::binary);
    if (!file)
      throw std::runtime_error("cannot create: " + path.string());
    static const uint8_t signature[8] = {137, 80, 78, 71, 13, 10, 26, 10};
    file.write((const char *)signature, 8);
    WriteChunk(file, "IHDR", m_ihdr.data(), m_ihdr.size());
    WriteChunk(file, "IDAT", m_packed.data(), (size_t)packedSize);
    WriteChunk(file, "IEND", nullptr, 0);
  }

private:
  static void PutBE32(std::vector<uint8_t> &out, uint32_t value) {
    out.push_back((uint8_t)(value >> 24));
    out.push_back((uint8_t)(value >> 16));
    out.push_back((uint8_t)(value >> 8));
    out.push_back((uint8_t)value);
  }

  static void WriteChunk(std::ofstream &file, const char type[4],
                         const uint8_t *data, size_t size) {
    std::vector<uint8_t> crcInput(4 + size);
    std::memcpy(crcInput.data(), type, 4);
    if (size)
      std::memcpy(crcInput.data() + 4, data, size);
    std::vector<uint8_t> head;
    PutBE32(head, (uint32_t)size);
    file.write((const char *)head.data(), 4);
    file.write(type, 4);
    if (size)
      file.write((const char *)data, (std::streamsize)size);
    std::vector<uint8_t> crc;
    PutBE32(crc, (uint32_t)mz_crc32(0, crcInput.data(), crcInput.size()));
    file.write((const char *)crc.data(), 4);
  }

  int m_width = 0;
  int m_height = 0;
  std::vector<uint8_t> m_ihdr;
  std::vector<uint8_t> m_raw;
  std::vector<uint8_t> m_packed;
};
} // namespace Png

class FrameExporter {
public:
  FrameExporter(const AmvMovie &movie, const fs::path &framesDir)
      : m_movie(movie), m_framesDir(framesDir),
        m_writer(movie.Width(), movie.Height()),
        m_canvas((size_t)movie.Width() * movie.Height() * 4, 0) {}

  void ExportAll() {
    const size_t count = m_movie.FrameCount();
    DecodedFrame frame;
    for (size_t i = 0; i < count; ++i) {
      m_movie.DecodeFrame(i, frame);
      Blit(frame);
      std::ostringstream name;
      name << "frame_" << std::setfill('0') << std::setw(6) << i << ".png";
      m_writer.Write(m_framesDir / name.str(), m_canvas);
      if ((i + 1) % 10 == 0 || i + 1 == count)
        std::cout << "\rDecoded " << (i + 1) << "/" << count << std::flush;
    }
    std::cout << "\n"
              << m_movie.Width() << "x" << m_movie.Height() << " @ "
              << m_movie.Fps() << " fps\n";
  }

private:
  void Blit(const DecodedFrame &frame) {
    const int w = m_movie.Width();
    const int h = m_movie.Height();
    int sx = 0, sy = 0;
    int dx = frame.left, dy = frame.top;
    int cw = frame.width, ch = frame.height;
    if (dx < 0) {
      sx = -dx;
      cw += dx;
      dx = 0;
    }
    if (dy < 0) {
      sy = -dy;
      ch += dy;
      dy = 0;
    }
    cw = std::min(cw, w - dx);
    ch = std::min(ch, h - dy);
    if (cw > 0 && ch > 0) {
      for (int y = 0; y < ch; ++y) {
        std::memcpy(m_canvas.data() + ((size_t)(dy + y) * w + dx) * 4,
                    frame.rgba.data() +
                        ((size_t)(sy + y) * frame.width + sx) * 4,
                    (size_t)cw * 4);
      }
    }
  }

  const AmvMovie &m_movie;
  fs::path m_framesDir;
  Png::Writer m_writer;
  std::vector<uint8_t> m_canvas;
};

class Ffmpeg {
public:
  explicit Ffmpeg(fs::path exe) : m_exe(std::move(exe)) {}

  void Encode(double fps, const fs::path &framesDir, size_t frameCount,
              const fs::path &output) const {
    std::vector<std::wstring> args;
    args.push_back(m_exe.wstring());
    args.push_back(L"-y");
    args.push_back(L"-framerate");
    args.push_back(std::to_wstring(fps));
    args.push_back(L"-start_number");
    args.push_back(L"0");
    args.push_back(L"-i");
    args.push_back((framesDir / L"frame_%06d.png").wstring());
    args.push_back(L"-frames:v");
    args.push_back(std::to_wstring(frameCount));
    args.push_back(L"-c:v");
    args.push_back(L"libx264");
    args.push_back(L"-pix_fmt");
    args.push_back(L"yuv420p");
    args.push_back(L"-movflags");
    args.push_back(L"+faststart");
    args.push_back(output.wstring());

    std::vector<const wchar_t *> argv;
    argv.reserve(args.size() + 1);
    for (const auto &arg : args)
      argv.push_back(arg.c_str());
    argv.push_back(nullptr);

    const intptr_t rc = _wspawnv(_P_WAIT, args[0].c_str(), argv.data());
    const int code = rc < 0 ? errno : (int)rc;
    if (code != 0)
      throw std::runtime_error("ffmpeg failed with exit code " +
                               std::to_string(code));
  }

private:
  fs::path m_exe;
};

struct Options {
  fs::path input;
  fs::path ffmpeg;

  static void PrintUsage() {
    std::cout
        << "AlphaMovie CLI\n\n"
        << "Usage: alphamovie -i <path to .amv> -f <path to ffmpeg>\n\n"
        << "Options:\n"
        << "  -i <path to .amv>    input AlphaMovie file\n"
        << "  -f <path to ffmpeg>  ffmpeg executable used to encode the mp4\n"
        << "  -h, --help           show this help\n";
  }

  static Options Parse(int argc, char *argv[]) {
    Options opts;
    for (int i = 1; i < argc; ++i) {
      const std::string arg = argv[i];
      if (arg == "-i") {
        if (i + 1 >= argc)
          throw std::runtime_error("missing input file");
        opts.input = argv[++i];
      } else if (arg == "-f") {
        if (i + 1 >= argc)
          throw std::runtime_error("missing ffmpeg path");
        opts.ffmpeg = argv[++i];
      } else {
        throw std::runtime_error("unknown option: " + arg);
      }
    }
    if (opts.input.empty())
      throw std::runtime_error("input .amv is required (-i)");
    if (opts.ffmpeg.empty())
      throw std::runtime_error("ffmpeg path is required (-f)");
    std::string ext = opts.input.extension().string();
    std::transform(ext.begin(), ext.end(), ext.begin(),
                   [](unsigned char c) { return (char)std::tolower(c); });
    if (ext != ".amv")
      throw std::runtime_error("input must be a .amv file");
    return opts;
  }
};

int main(int argc, char *argv[]) {
  try {
    if (argc < 2) {
      Options::PrintUsage();
      return 2;
    }
    for (int i = 1; i < argc; ++i) {
      const std::string arg = argv[i];
      if (arg == "-h" || arg == "--help") {
        Options::PrintUsage();
        return 0;
      }
    }
    const Options opts = Options::Parse(argc, argv);

    AmvMovie movie;
    movie.Open(opts.input);

    fs::path framesDir = opts.input.parent_path() / opts.input.stem();
    fs::path video = opts.input;
    video.replace_extension(".mp4");
    fs::create_directories(framesDir);

    FrameExporter(movie, framesDir).ExportAll();

    std::cout << "Running FFmpeg...\n";
    Ffmpeg(opts.ffmpeg)
        .Encode(movie.Fps(), framesDir, movie.FrameCount(), video);
    std::cout << "Output: " << video.string() << "\n";
    return 0;
  } catch (const std::exception &e) {
    std::cerr << "Error: " << e.what() << "\n";
    return 1;
  }
}
