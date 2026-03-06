// Copyright (C) 2011-2016, Yann Collet.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// C# implementation Copyright (C) 2017 by morkt
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.
//

using System;

namespace TLGQoi
{
    public class Lz4Compressor
    {
        const int MinMatch = 4;
        const int LastLiterals = 5;
        const int MFLimit = 12;
        const int MatchLengthBits = 4;
        const int MatchLengthMask = 0xF;
        const int RunMask = 0xF;

        public static int DecompressBlock(byte[] block, int block_size, byte[] output, int output_size)
        {
            int src = 0;
            int iend = block_size;

            int dst = 0;
            int oend = output_size;

            for (; ; )
            {
                /* get literal length */
                int token = block[src++];
                int length = token >> MatchLengthBits;
                if (RunMask == length)
                {
                    int n;
                    do
                    {
                        n = block[src++];
                        length += n;
                    }
                    while ((src < iend - RunMask) && (0xFF == n));
                    if (dst + length < dst || src + length < src) // overflow detection
                        throw new Exception();
                }

                /* copy literals */
                int copy_end = dst + length;
                if ((copy_end > oend - MFLimit) || (src + length > iend - (3 + LastLiterals)))
                {
                    if ((src + length != iend) || copy_end > oend)
                        throw new Exception();
                    Buffer.BlockCopy(block, src, output, dst, length);
                    src += length;
                    dst += length;
                    break;
                }
                Buffer.BlockCopy(block, src, output, dst, length);
                src += length;
                dst = copy_end;

                /* get offset */
                int offset = ReadUInt16(block, src);
                src += 2;
                int match = dst - offset;
                if (match < 0)
                    throw new Exception();

                /* get matchlength */
                length = token & MatchLengthMask;
                if (MatchLengthMask == length)
                {
                    int n;
                    do
                    {
                        n = block[src++];
                        if (src > iend - LastLiterals)
                            throw new Exception();
                        length += n;
                    }
                    while (0xFF == n);
                    if (dst + length < dst) // overflow detection
                        throw new Exception();
                }
                length += MinMatch;

                /* copy match within block */
                CopyOverlapped(output, match, dst, length);
                dst += length;
            }
            return dst; // number of output bytes decoded
        }


        private static ushort ReadUInt16(byte[] data, int offset)
        {
            return (ushort)(data[offset] | (data[offset + 1] << 8));
        }

        private static void CopyOverlapped(byte[] data, int src, int dst, int count)
        {
            if (src >= dst || dst - src >= count)
            {
                Buffer.BlockCopy(data, src, data, dst, count);
                return;
            }
            while (count-- > 0)
                data[dst++] = data[src++];
        }

        public static int DecompressLinkedLz4(byte[] block, int block_size, byte[] output, int output_size, byte[] dict, int dict_size)
        {
            int src = 0;
            int dst = 0;
            while (src < block_size)
            {
                int token = block[src++];
                int literal_length = token >> 4;
                if (15 == literal_length)
                {
                    int extra;
                    do
                    {
                        if (src >= block_size)
                            return -1;
                        extra = block[src++];
                        literal_length += extra;
                    }
                    while (0xFF == extra);
                }
                if (src + literal_length > block_size || dst + literal_length > output_size)
                    return -1;

                Buffer.BlockCopy(block, src, output, dst, literal_length);
                src += literal_length;
                dst += literal_length;
                if (src >= block_size)
                    break;

                if (src + 2 > block_size)
                    return -1;
                int offset = ReadUInt16(block, src);
                src += 2;
                if (offset <= 0)
                    return -1;

                int match_length = token & 0x0F;
                if (15 == match_length)
                {
                    int extra;
                    do
                    {
                        if (src >= block_size)
                            return -1;
                        extra = block[src++];
                        match_length += extra;
                    }
                    while (0xFF == extra);
                }
                match_length += 4;
                if (dst + match_length > output_size)
                    return -1;

                int match = dst - offset;
                if (match < 0)
                {
                    int need = -match;
                    if (need > dict_size)
                        return -1;
                    int copy = Math.Min(need, match_length);
                    Buffer.BlockCopy(dict, dict_size - need, output, dst, copy);
                    dst += copy;
                    match_length -= copy;
                    match = 0;
                }
                if (match_length > 0)
                {
                    CopyOverlapped(output, match, dst, match_length);
                    dst += match_length;
                }
            }
            return dst;
        }
    }
}
