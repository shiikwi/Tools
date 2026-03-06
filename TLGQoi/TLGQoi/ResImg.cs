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
        public void ImgConvert(string inFile, string outFile)
        {
            using (var fs = new FileStream(inFile, FileMode.Open))
            using (var br = new BinaryReader(fs))
            {
                var sig = br.ReadBytes(0xB);
                if (Encoding.ASCII.GetString(sig) == "TLGqoi\0raw\x1a")
                    ReadTLGQoi(br, outFile);
            }
        }

        private void ReadTLGQoi(BinaryReader br, string outPath)
        {
            var bpp = br.ReadByte();
            var width = br.ReadUInt32();
            var height = br.ReadUInt32();
            var qhdr = ReadQHDR(br);
            var qoi = new QoiDecoder();
            var data = br.ReadBytes((int)(br.BaseStream.Length - br.BaseStream.Position));
            byte[] output;
            if (qhdr.FingerPrint != 0)
            {
                output = qoi.DecodeQoiStream3(data, (int)width, (int)height, qhdr);
            }
            else
            {
                output = qoi.DecodeQoiStream4(data, (int)width, (int)height);
            }

            SavePng(output, (int)width, (int)height, outPath);
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
