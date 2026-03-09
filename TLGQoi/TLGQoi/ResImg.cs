using System;
using System.Collections.Generic;
using System.Text;
using System.Drawing;
using System.Drawing.Imaging;
using System.Runtime.InteropServices;
#pragma warning disable

namespace TLGQoi
{
    public class ResImg
    {
        private sealed class TlgImageData
        {
            public int Width;
            public int Height;
            public byte[] Pixels = Array.Empty<byte>();
        }

        private Dictionary<string, string> MainTLGImg = new Dictionary<string, string>();
        private HashSet<string> UsedMainTLG = new HashSet<string>();
        private List<string> RefFile = new List<string>();
        private List<string> MuxFile = new List<string>();

        public void InitDiffMap(string inFile)
        {
            using (var fs = new FileStream(inFile, FileMode.Open))
            using (var br = new BinaryReader(fs))
            {
                var sig = br.ReadBytes(0xB);
                string fileName = Path.GetFileName(inFile);
                string signature = Encoding.ASCII.GetString(sig);
                if (signature == "TLGqoi\0raw\x1a")
                {
                    MainTLGImg[fileName] = inFile;
                }
                else if (signature == "TLGref\0raw\x1a")
                {
                    RefFile.Add(inFile);
                }
                else if (signature == "TLGmux\0idx\x1a")
                {
                    MuxFile.Add(inFile);
                }
            }
        }

        public void ImgConvert()
        {
            foreach (var reff in RefFile)
            {
                var qref = ReadQREF(reff);
                string outPath = Path.Combine(Path.GetDirectoryName(reff), Path.GetFileNameWithoutExtension(reff) + ".png");
                if (MainTLGImg.TryGetValue(qref.MainImgName, out var mainImgPath))
                {
                    UsedMainTLG.Add(qref.MainImgName);
                    ReadTLGQoi(mainImgPath, outPath, (int)qref.SamplePhase);
                }
                else
                {
                    Console.WriteLine($"Cannot read MainImage of {Path.GetFileName(reff)}");
                }
            }
            Console.WriteLine("Process TLGqoi diff images finish");

            foreach (var kvp in MainTLGImg)
            {
                if (!UsedMainTLG.Contains(kvp.Key))
                {
                    string outPath = Path.ChangeExtension(kvp.Value, ".png");
                    ReadTLGQoi(kvp.Value, outPath, 0);
                    Console.WriteLine($"Cannot read diff images of {kvp.Key} phase use 0");
                }
            }
            Console.WriteLine("Process TLGqoi finish");

            foreach (var mux in MuxFile)
            {
                string outPath = Path.ChangeExtension(mux, ".png");
                ReadTLGMux(mux, outPath);
            }
            Console.WriteLine("Process TLGmux finish");
        }

        private void ReadTLGQoi(string inPath, string outPath, int samplePhase)
        {
            using (var fs = new FileStream(inPath, FileMode.Open))
            using (var br = new BinaryReader(fs))
            {
                var image = ReadTLGQoiImage(br, samplePhase);
                SavePng(image.Pixels, image.Width, image.Height, outPath);
            }
        }

        private QoiDecoder.TLGQoiQREF ReadQREF(string inPath)
        {
            using (var fs = new FileStream(inPath, FileMode.Open))
            using (var br = new BinaryReader(fs))
            {
                br.BaseStream.Seek(0xB, SeekOrigin.Current);
                var bpp = br.ReadByte();
                var width = br.ReadUInt32();
                var height = br.ReadUInt32();
                var qref = new QoiDecoder.TLGQoiQREF();
                uint tag = br.ReadUInt32();
                uint size = br.ReadUInt32();
                if (tag == 0x46455251)  //QREF
                {
                    qref.FingerPrint = br.ReadUInt32();
                    qref.SamplePhase = br.ReadUInt32();
                    qref.BlockWidth = br.ReadUInt32();
                    uint nameLen = br.ReadUInt32();
                    qref.MainImgName = Encoding.Unicode.GetString(br.ReadBytes((int)nameLen));
                }
                return qref;
            }
        }

        private QoiDecoder.TLGQoiQHDR ReadQHDR(BinaryReader br)
        {
            var qhdr = new QoiDecoder.TLGQoiQHDR();
            while (br.BaseStream.Position + 8 <= br.BaseStream.Length)
            {
                uint tag = br.ReadUInt32();
                uint size = br.ReadUInt32();
                if (tag == 0 && size == 0)
                    break;
                if (tag == 0x52444851)   //QHDR
                {
                    qhdr.FingerPrint = br.ReadUInt32();
                    qhdr.BlockWidth = br.ReadUInt32();
                    qhdr.ImageHeight = br.ReadUInt32();
                    qhdr.BandCount = br.ReadUInt32();
                    qhdr.Edge0 = br.ReadInt64();
                    qhdr.Edge1 = br.ReadInt64();
                    qhdr.Edge2 = br.ReadInt64();
                    qhdr.Edge3 = br.ReadInt64();
                }
                else
                {
                    br.BaseStream.Seek(size, SeekOrigin.Current);
                }
            }

            return qhdr;
        }

        private TlgImageData ReadTLGQoiImage(BinaryReader br, int samplePhase)
        {
            string signature = Encoding.ASCII.GetString(br.ReadBytes(0xB));
            if (signature != "TLGqoi\0raw\x1a")
                throw new InvalidDataException("Invalid TLGqoi header");

            var bpp = br.ReadByte();
            int width = (int)br.ReadUInt32();
            int height = (int)br.ReadUInt32();
            var qhdr = ReadQHDR(br);
            var qoi = new QoiDecoder();
            var data = br.ReadBytes((int)(br.BaseStream.Length - br.BaseStream.Position));
            byte[] output = qhdr.FingerPrint != 0 ? qoi.DecodeQoiStream3(data, width, height, qhdr, samplePhase)
                                                  : qoi.DecodeQoiStream4(data, width, height);

            return new TlgImageData
            {
                Width = width,
                Height = height,
                Pixels = output,
            };
        }

        private void ReadTLGMux(string inPath, string outPath)
        {
            using (var fs = new FileStream(inPath, FileMode.Open))
            using (var br = new BinaryReader(fs))
            {
                string signature = Encoding.ASCII.GetString(br.ReadBytes(0xB));
                if (signature != "TLGmux\0idx\x1a")
                    throw new InvalidDataException($"Invalid TLGmux header: {Path.GetFileName(inPath)}");

                var bpp = br.ReadByte();
                int canvasWidth = (int)br.ReadUInt32();
                int canvasHeight = (int)br.ReadUInt32();

                uint cmuxTag = br.ReadUInt32();
                uint cmuxSize = br.ReadUInt32();
                if (cmuxTag != 0x58554D43) // CMUX
                    throw new InvalidDataException($"Missing CMUX chunk: {Path.GetFileName(inPath)}");

                long cmuxStart = br.BaseStream.Position;
                int entryCount = (int)br.ReadUInt32();
                var entries = new List<QoiDecoder.TLGMuxEntry>(entryCount);
                for (int i = 0; i < entryCount; ++i)
                {
                    entries.Add(new QoiDecoder.TLGMuxEntry
                    {
                        X = br.ReadUInt32(),
                        Y = br.ReadUInt32(),
                        Width = br.ReadUInt32(),
                        Height = br.ReadUInt32(),
                        Offset = br.ReadUInt32(),
                        Reserved = br.ReadUInt32(),
                    });

                }

                br.BaseStream.Position = cmuxStart + cmuxSize;
                br.ReadBytes(8);  //Skip padding
                long tileDataBase = br.BaseStream.Position;
                long tileDataSize = br.BaseStream.Length - tileDataBase;
                var canvas = new byte[canvasWidth * canvasHeight * 4];

                for (int i = 0; i < entries.Count; ++i)
                {
                    var entry = entries[i];
                    int tileLength = GetTileLength(entries, i, tileDataSize);
                    br.BaseStream.Position = tileDataBase + entry.Offset;
                    byte[] tileBytes = br.ReadBytes(tileLength);
                    using (var ts = new MemoryStream(tileBytes, false))
                    using (var tr = new BinaryReader(ts))
                    {
                        var tile = ReadTLGQoiImage(tr, 0);
                        BlitTile(canvas, canvasWidth, canvasHeight, tile.Pixels, entry);
                    }
                }

                SavePng(canvas, canvasWidth, canvasHeight, outPath);
            }
        }


        private static int GetTileLength(List<QoiDecoder.TLGMuxEntry> entries, int index, long tileDataSize)
        {
            int currentOffset = (int)entries[index].Offset;
            int nextOffset = index + 1 < entries.Count ? (int)entries[index + 1].Offset : (int)tileDataSize;
            return nextOffset - currentOffset;
        }

        private static void BlitTile(byte[] canvas, int canvasWidth, int canvasHeight, byte[] tilePixels, QoiDecoder.TLGMuxEntry entry)
        {
            if (entry.X + entry.Width > canvasWidth || entry.Y + entry.Height > canvasHeight)
                throw new InvalidDataException("TLGmux tile placement exceeds canvas bounds");

            int tileStride = (int)entry.Width * 4;
            for (int row = 0; row < entry.Height; ++row)
            {
                int src = row * tileStride;
                int dst = (int)(((entry.Y + row) * canvasWidth + entry.X) * 4);
                Buffer.BlockCopy(tilePixels, src, canvas, dst, tileStride);
            }
        }

        private void SavePng(byte[] data, int w, int h, string outPath)
        {
            using (var bitmap = new Bitmap(w, h, PixelFormat.Format32bppArgb))
            {
                var rect = new Rectangle(0, 0, w, h);
                var pngData = bitmap.LockBits(rect, ImageLockMode.WriteOnly, bitmap.PixelFormat);
                Marshal.Copy(data, 0, pngData.Scan0, data.Length);
                bitmap.UnlockBits(pngData);
                bitmap.Save(outPath, ImageFormat.Png);
            }
        }

    }
}
