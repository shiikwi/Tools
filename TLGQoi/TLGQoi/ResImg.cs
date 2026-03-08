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
        private Dictionary<string, string> MainTLGImg = new Dictionary<string, string>();
        private HashSet<string> UsedMainTLG = new HashSet<string>();
        private List<string> RefFile = new List<string>();

        public void InitDiffMap(string inFile)
        {
            using (var fs = new FileStream(inFile, FileMode.Open))
            using (var br = new BinaryReader(fs))
            {
                var sig = br.ReadBytes(0xB);
                string fileName = Path.GetFileName(inFile);
                if (Encoding.ASCII.GetString(sig) == "TLGqoi\0raw\x1a")
                {
                    MainTLGImg[fileName] = inFile;
                }
                else if (Encoding.ASCII.GetString(sig) == "TLGref\0raw\x1a")
                {
                    RefFile.Add(inFile);
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

            foreach(var kvp in MainTLGImg)
            {
                if(!UsedMainTLG.Contains(kvp.Key))
                {
                    string outPath = Path.ChangeExtension(kvp.Value, ".png");
                    ReadTLGQoi(kvp.Value, outPath, 0);
                    Console.WriteLine($"Cannot read diff images of {kvp.Key} phase use 0");
                }
            }
        }

        private void ReadTLGQoi(string inPath, string outPath, int samplePhase)
        {
            using (var fs = new FileStream(inPath, FileMode.Open))
            using (var br = new BinaryReader(fs))
            {
                br.BaseStream.Seek(0xB, SeekOrigin.Current);
                var bpp = br.ReadByte();
                var width = br.ReadUInt32();
                var height = br.ReadUInt32();
                var qhdr = ReadQHDR(br);
                var qoi = new QoiDecoder();
                var data = br.ReadBytes((int)(br.BaseStream.Length - br.BaseStream.Position));
                byte[] output = qhdr.FingerPrint != 0 ? qoi.DecodeQoiStream3(data, (int)width, (int)height, qhdr, samplePhase)
                                                      : qoi.DecodeQoiStream4(data, (int)width, (int)height);

                SavePng(output, (int)width, (int)height, outPath);
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
