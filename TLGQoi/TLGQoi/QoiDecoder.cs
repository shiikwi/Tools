using System;
using System.Collections.Generic;

namespace TLGQoi
{
    internal class QoiDecoder
    {
        static readonly int[] KindTable = { 2, 4, 6, 1 };
        static readonly byte[] SizeTable = { 0, 1, 1, 2, 1, 2, 2, 3, 4, 5, 5, 6, 0, 0, 0, 0 };

        private sealed class Bpp3Tables
        {
            public int worker_count;
            public int run_base;
            public int data_end;
            public long[] op_lengths;
            public long[] run_lengths;
            public ulong[] control_values;
        }

        public class TLGQoiQHDR
        {
            public uint FingerPrint;
            public uint BlockWidth;
            public uint ImageHeight;
            public uint BandCount;
            public long Edge0;
            public long Edge1;
            public long Edge2;
            public long Edge3;
        }

        public class TLGQoiQREF
        {
            public uint FingerPrint;
            public uint SamplePhase;
            public uint BlockWidth;
            public string MainImgName;
        }

        public byte[] DecodeQoiStream4(byte[] data, int width, int height)
        {
            int pixel_count = width * height;
            var output = new byte[pixel_count * 4];
            var table = new uint[64];
            var op_stream = new OpStream(data, 0);
            uint current_pixel = 0xFF000000;
            int run_remaining = 0;
            int dst = 0;

            for (int i = 0; i < pixel_count; ++i)
            {
                if (run_remaining > 0)
                {
                    --run_remaining;
                }
                else
                {
                    int count;
                    int kind;
                    if (!op_stream.Next(table, current_pixel, out current_pixel, out count, out kind))
                        break;
                    run_remaining = count - 1;
                }
                WritePixel(output, ref dst, current_pixel);
            }
            return output;
        }

        public byte[] DecodeQoiStream3(byte[] input, int width, int height, TLGQoiQHDR qhdr, int SamplePhase = 0)
        {
            Bpp3Tables tables;
            if (!ReadBpp3Tables(input, qhdr, out tables))
                return null;

            int worker_count = tables.worker_count;
            if (qhdr.BandCount > 0)
                worker_count = Math.Min(worker_count, (int)qhdr.BandCount);
            if (worker_count <= 0)
                return null;

            int band_rows = (int)qhdr.ImageHeight;
            int sample_step = (int)qhdr.BlockWidth;

            var output = new byte[width * height * 4];
            var control_stream = new ulong[worker_count + 2];
            Array.Copy(tables.control_values, control_stream, worker_count);
            control_stream[worker_count] = (ulong)qhdr.Edge0;
            long op_pos = 0;
            long run_pos = 0;
            int band_y = 0;
            int written_total = 0;
            for (int worker_index = 0; worker_index < worker_count && band_y < height; ++worker_index)
            {
                int row_count = Math.Min(band_rows, height - band_y);
                if (row_count <= 0)
                    break;

                int op_offset = Math.Max(0, Math.Min((int)op_pos, input.Length));
                int run_offset = tables.run_base + Math.Max(0, Math.Min((int)run_pos, tables.data_end - tables.run_base));
                int dst = band_y * width * 4;
                int written = DecodeBpp3Band(
                    output,
                    dst,
                    width,
                    row_count,
                    sample_step,
                    SamplePhase,
                    new OpStream(input, op_offset),
                    new RunStream(input, run_offset),
                    control_stream,
                    worker_index);
                if (written <= 0)
                    return null;

                written_total += written;
                band_y += row_count;
                op_pos += tables.op_lengths[worker_index];
                run_pos += tables.run_lengths[worker_index];
            }
            return written_total > 0 ? output : null;
        }

        private static bool ReadBpp3Tables(byte[] input, TLGQoiQHDR qhdr, out Bpp3Tables tables)
        {
            tables = null;
            int edge_1 = (int)qhdr.Edge1;
            int edge_2 = (int)qhdr.Edge2;
            int edge_3 = (int)qhdr.Edge3;

            byte[] dtbl_payload;
            int dtbl_end;
            if (!ReadTaggedPayload(input, edge_1, 0x4C425444u, out dtbl_payload, out dtbl_end))  //"DTBL"
                return false;

            byte[] rtbl_payload;
            int rtbl_end;
            if (!ReadTaggedPayload(input, edge_2, 0x4C425452u, out rtbl_payload, out rtbl_end))  //"RTBL"
                return false;

            long[] op_lengths;
            ulong[] control_values;
            int worker_count;
            if (!ParseDtbl(dtbl_payload, out op_lengths, out control_values, out worker_count))
                return false;

            long[] run_lengths;
            if (!ParseRtbl(rtbl_payload, worker_count, out run_lengths))
                return false;

            worker_count = Math.Min(worker_count, Math.Min(op_lengths.Length, run_lengths.Length));
            if (worker_count <= 0 || rtbl_end < 0 || rtbl_end > edge_3)
                return false;

            tables = new Bpp3Tables
            {
                worker_count = worker_count,
                run_base = rtbl_end,
                data_end = edge_3,
                op_lengths = op_lengths,
                run_lengths = run_lengths,
                control_values = control_values,
            };
            return true;
        }

        private int DecodeBpp3Band(byte[] output, int dst, int width, int rows, int sample_step, int sample_phase,
                                    OpStream op_stream, RunStream run_stream, ulong[] control_stream, int control_index)
        {
            var qoi_table = new uint[64];
            uint current_pixel = 0xFF000000;
            ulong chunk_budget = 0;
            int run_remaining = 0;
            int written = 0;
            int row_stride = width * 4;

            for (int row = 0; row < rows; ++row)
            {
                int row_dst = dst + row * row_stride;
                int countdown = sample_phase;
                int total_samples = width * sample_step;
                for (int sample = 0; sample < total_samples; ++sample)
                {
                    while (run_remaining <= 0)
                    {
                        if (0 == chunk_budget)
                        {
                            if (!RefillChunkState(op_stream, run_stream, control_stream, ref control_index,
                                                      qoi_table, out current_pixel, out chunk_budget, out run_remaining))
                                return written;
                            if (run_remaining > 0)
                                break;
                        }
                        if (0 == chunk_budget)
                            return written;

                        --chunk_budget;

                        uint pixel;
                        int count;
                        int kind;
                        if (!op_stream.Next(qoi_table, current_pixel, out pixel, out count, out kind))
                            return written;

                        int extra = run_stream.ReadVarInt();
                        if (extra < 0)
                            return written;

                        current_pixel = pixel;
                        run_remaining = count + extra;
                        if (run_remaining <= 0)
                            return written;
                    }

                    --run_remaining;
                    if (0 == countdown)
                    {
                        WritePixel(output, ref row_dst, current_pixel | 0xFF000000);
                        ++written;
                        countdown = sample_step - 1;
                    }
                    else
                    {
                        --countdown;
                    }
                }
            }
            return written;
        }

        private bool RefillChunkState(OpStream op_stream, RunStream run_stream, ulong[] control_stream, ref int control_index,
                                          uint[] qoi_table, out uint current_pixel, out ulong chunk_budget, out int run_remaining)
        {
            Array.Clear(qoi_table, 0, qoi_table.Length);
            current_pixel = 0xFF000000;
            chunk_budget = 0;
            run_remaining = 0;

            uint marker_0;
            int count_0;
            int kind_0;
            if (!op_stream.Next(qoi_table, current_pixel, out marker_0, out count_0, out kind_0))
                return false;

            uint marker_1;
            int count_1;
            int kind_1;
            if (!op_stream.Next(qoi_table, marker_0, out marker_1, out count_1, out kind_1))
                return false;

            if (control_index >= 0 && control_index < control_stream.Length)
                chunk_budget = control_stream[control_index++];

            run_remaining = run_stream.ReadVarInt();
            if (run_remaining < 0)
                return false;

            return 1 == count_0 && 1 == count_1 && 0 == marker_0 && 0xFF000000 == marker_1 && chunk_budget > 0;
        }

        private static bool ReadTaggedPayload(byte[] data, int offset, uint tag, out byte[] payload, out int next)
        {
            payload = null;
            next = offset;
            if (offset < 0 || offset + 8 > data.Length)
                return false;
            if (ReadUInt32(data, offset) != tag)
                return false;

            uint size = ReadUInt32(data, offset + 4);
            if (size > data.Length - offset - 8)
                return false;

            payload = new byte[size];
            Buffer.BlockCopy(data, offset + 8, payload, 0, (int)size);
            next = offset + 8 + (int)size;
            return true;
        }

        private static bool ParseDtbl(byte[] payload, out long[] op_lengths, out ulong[] control_values, out int worker_count)
        {
            op_lengths = null;
            control_values = null;
            worker_count = 0;

            List<ulong> values;
            if (!ParseVarUIntList(payload, out values) || values.Count < 3)
                return false;

            worker_count = (values.Count - 1) / 2;
            if (worker_count <= 0)
                return false;

            op_lengths = new long[worker_count];
            control_values = new ulong[worker_count];
            for (int i = 0; i < worker_count; ++i)
            {
                op_lengths[i] = (long)values[1 + i * 2];
                control_values[i] = values[2 + i * 2];
            }
            return true;
        }

        private static bool ParseRtbl(byte[] payload, int worker_count, out long[] run_lengths)
        {
            run_lengths = null;
            if (worker_count <= 0)
                return false;

            List<ulong> values;
            if (!ParseVarUIntList(payload, out values) || values.Count - 1 < worker_count)
                return false;

            run_lengths = new long[worker_count];
            for (int i = 0; i < worker_count; ++i)
                run_lengths[i] = (long)values[1 + i];
            return true;
        }

        private static bool ParseVarUIntList(byte[] data, out List<ulong> values)
        {
            values = new List<ulong>();
            int pos = 0;
            while (pos < data.Length)
            {
                ulong value;
                if (!ReadVarUInt64(data, ref pos, out value))
                    return false;
                values.Add(value);
            }
            return values.Count > 0;
        }

        private static bool ReadVarUInt64(byte[] data, ref int pos, out ulong value)
        {
            value = 0;
            int shift = 0;
            while (pos < data.Length && shift < 64)
            {
                byte current = data[pos++];
                value |= (ulong)(current & 0x7F) << shift;
                if (0 == (current & 0x80))
                    return true;
                shift += 7;
            }
            return false;
        }

        private static uint InternalPixelDecode(ulong packet, int kind, uint prev_pixel, uint[] table, out int count)
        {
            uint pixel = prev_pixel;
            count = 1;
            uint lo = (uint)packet;
            uint hi = (uint)(packet >> 32);
            switch (kind)
            {
                case 1:
                    count = (int)(lo & 0x3F) + 1;
                    break;
                case 2:
                    pixel = table[lo & 0x3F];
                    break;
                case 4:
                    {
                        int db = (int)(lo & 3) - 2;
                        int dg = (int)((lo >> 2) & 3) - 2;
                        int dr = (int)((lo >> 4) & 3) - 2;
                        uint b = (uint)((byte)prev_pixel + db);
                        uint g = (uint)((byte)(prev_pixel >> 8) + dg);
                        uint r = (uint)((byte)(prev_pixel >> 16) + dr);
                        pixel = (prev_pixel & 0xFF000000) | b | (g << 8) | (r << 16);
                    }
                    break;
                case 6:
                    {
                        int dg = (int)(lo & 0x3F) - 0x20;
                        int db = (int)(((lo >> 8) | (hi << 0x18)) & 0x0F) - 8 + dg;
                        int dr = (int)(((lo >> 12) | (hi << 0x14)) & 0x0F) - 8 + dg;
                        uint b = (uint)((byte)prev_pixel + db);
                        uint g = (uint)((byte)(prev_pixel >> 8) + dg);
                        uint r = (uint)((byte)(prev_pixel >> 16) + dr);
                        pixel = (prev_pixel & 0xFF000000) | b | (g << 8) | (r << 16);
                    }
                    break;
                case 8:
                    {
                        uint b = (lo >> 24) | (hi << 8);
                        uint g = (lo >> 16) | (hi << 16);
                        uint r = (lo >> 8) | (hi << 24);
                        pixel = (prev_pixel & 0xFF000000) | (b & 0xFF) | ((g & 0xFF) << 8) | ((r & 0xFF) << 16);
                    }
                    break;
                case 0x0A:
                    {
                        uint b = (lo >> 24) | (hi << 8);
                        uint g = (lo >> 16) | (hi << 16);
                        uint r = (lo >> 8) | (hi << 24);
                        uint a = hi & 0xFF;
                        pixel = (b & 0xFF) | ((g & 0xFF) << 8) | ((r & 0xFF) << 16) | (a << 24);
                    }
                    break;
            }
            return pixel;
        }

        private static int CalculateHash(uint pixel)
        {
            return (((int)(pixel >> 24) * 11)
                    + ((int)(pixel >> 8 & 0xFF) * 5)
                    + ((int)(pixel >> 16 & 0xFF) * 3)
                    + ((int)(pixel & 0xFF) * 7)) & 0x3F;
        }

        private static void WritePixel(byte[] buffer, ref int dst, uint pixel)
        {
            buffer[dst++] = (byte)pixel;
            buffer[dst++] = (byte)(pixel >> 8);
            buffer[dst++] = (byte)(pixel >> 16);
            buffer[dst++] = (byte)(pixel >> 24);
        }

        private static uint ReadUInt32(byte[] data, int offset)
        {
            return (uint)(data[offset]
                | (data[offset + 1] << 8)
                | (data[offset + 2] << 16)
                | (data[offset + 3] << 24));
        }

        private sealed class OpStream
        {
            readonly byte[] m_data;
            int m_pos;

            public OpStream(byte[] data, int offset)
            {
                m_data = data;
                m_pos = offset;
            }

            public bool Next(uint[] table, uint prev_pixel, out uint pixel, out int count, out int kind)
            {
                pixel = prev_pixel;
                count = 0;
                kind = 0;
                if (m_pos >= m_data.Length)
                    return false;

                byte first = m_data[m_pos];
                kind = (first == 0xFF) ? 0x0A : (first == 0xFE ? 8 : KindTable[first >> 6]);
                int size = SizeTable[kind & 0x0F];
                if (size <= 0 || m_pos + size > m_data.Length)
                    return false;

                ulong packet = 0;
                for (int i = 0; i < size; ++i)
                    packet |= (ulong)m_data[m_pos + i] << (i * 8);
                m_pos += size;
                pixel = InternalPixelDecode(packet, kind, prev_pixel, table, out count);
                if (count > 0 && kind != 1)
                    table[CalculateHash(pixel)] = pixel;
                return true;
            }
        }

        private sealed class RunStream
        {
            const int MaxChunk = 0x8000;

            readonly byte[] m_data;
            readonly byte[] m_current = new byte[MaxChunk];
            byte[] m_previous = new byte[MaxChunk];
            int m_src;
            int m_pos;
            int m_len;
            int m_prev_len;

            public RunStream(byte[] data, int offset)
            {
                m_data = data;
                m_src = offset;
            }

            public int ReadVarInt()
            {
                int value = 0;
                int shift = 0;
                while (true)
                {
                    if (m_pos >= m_len && !Refill())
                        return -1;
                    if (m_pos >= m_len)
                        return -1;

                    byte current = m_current[m_pos++];
                    value |= (current & 0x7F) << shift;
                    if (0 == (current & 0x80))
                        return value;

                    shift += 7;
                    if (shift > 28)
                        return -1;
                }
            }

            private bool Refill()
            {
                if (m_src + 4 > m_data.Length)
                    return false;

                uint header = ReadUInt32(m_data, m_src);
                m_src += 4;
                int compressed_length = (int)(header >> 16);
                int decoded_length = (int)(header & 0x7FFF);
                if (0 == decoded_length)
                    decoded_length = MaxChunk;
                bool has_dict = 0 != (header & 0x8000);
                if (compressed_length < 0 || decoded_length <= 0 || decoded_length > MaxChunk)
                    return false;
                if (m_src + compressed_length > m_data.Length)
                    return false;

                var compressed = new byte[compressed_length];
                Buffer.BlockCopy(m_data, m_src, compressed, 0, compressed_length);
                m_src += compressed_length;

                int unpacked;
                try
                {
                    if (has_dict)
                        unpacked = Lz4Compressor.DecompressLinkedLz4(compressed, compressed_length, m_current, decoded_length, m_previous, m_prev_len);
                    else
                        unpacked = Lz4Compressor.DecompressBlock(compressed, compressed_length, m_current, decoded_length);
                }
                catch
                {
                    return false;
                }
                if (unpacked <= 0 || unpacked > decoded_length)
                    return false;
                if (unpacked < decoded_length)
                    Array.Clear(m_current, unpacked, decoded_length - unpacked);

                if (m_previous.Length < decoded_length)
                    m_previous = new byte[decoded_length];
                Buffer.BlockCopy(m_current, 0, m_previous, 0, decoded_length);
                m_prev_len = decoded_length;
                m_pos = 0;
                m_len = decoded_length;
                return true;
            }
        }

    }

}
