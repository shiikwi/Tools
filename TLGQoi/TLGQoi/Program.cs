using System;
using System.Collections.Generic;
using System.Text;

namespace TLGQoi
{
    class Program
    {
        static void Main(string[] args)
        {
            var inFilePath = args[0];
            var allFiles = Directory.GetFiles(inFilePath, "*.tlg", SearchOption.AllDirectories);
            var tlg = new ResImg();
            foreach(var file in allFiles)
            {
                tlg.InitDiffMap(file);
            }

            tlg.ImgConvert();
        }
    }
}