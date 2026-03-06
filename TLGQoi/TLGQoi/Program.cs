using System;
using System.Collections.Generic;
using System.Text;

namespace TLGQoi
{
    class Program
    {
        static void Main(string[] args)
        {
            var inFile = args[0];
            var outFile = Path.ChangeExtension(inFile, ".png");

            var cov = new ResImg();
            cov.ImgConvert(inFile, outFile);
        }
    }
}