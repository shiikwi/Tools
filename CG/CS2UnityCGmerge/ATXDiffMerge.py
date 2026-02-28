"""
The tool is used to merge diff imgs exported from CatSystem2ForUnity *.atx files, using https://github.com/rewjx/GameEngineTools

Usage: ATXDiffMerge.py -l/-b <ATXfileExportdirectory>
-l process one folder
-b process multiple folders

Notice: The edge Alpha fix is a band-aid solution
"""

import os
import json
import argparse
from itertools import product
from PIL import Image

def fixedge(image: Image.Image, threshold = 128) -> Image.Image:

    alpha = image.getchannel('A')
    
    hard_alpha = alpha.point(lambda p: 255 if p > threshold else 0)
    
    image.putalpha(hard_alpha)
    
    return image

def GetGroupInfo(infojson, mainimage):
    diffgroups = {}
    try:
        with open(infojson, 'r', encoding='utf-8-sig') as f:
            infodata = json.load(f)
            infos = {item['Key']: item['Value'] for item in infodata}
    except json.JSONDecodeError:
        print(f"Json Parse Failed: {infojson}")
        return
    
    for key, value in infos.items():
        if(value == mainimage):
            continue
        parts = key.split("_")
        if(len(parts) > 0):
            groupid = parts[0]
            if groupid not in diffgroups:
                diffgroups[groupid] = []
            diffgroups[groupid].append(value)
    
    return diffgroups


def merge_images(directory: str):
    offsetjson = os.path.join(directory, "offset.json")
    infojson = os.path.join(directory, "info.json")

    try:
        with open(offsetjson, 'r', encoding='utf-8-sig') as f:
            offsetdata = json.load(f)
            mainimage = offsetdata[0]['Key']   #mainimage is fixed the first element
            offsets = {item['Key']: tuple(item['Value']) for item in offsetdata}
    except json.JSONDecodeError:
        print(f"Json Parse Failed: {offsetjson}")
        return
    
    diffgroups = GetGroupInfo(infojson, mainimage)
    baseoffsetx, baseoffsety = offsets[mainimage]
    
    allcombinations = list(product(*diffgroups.values()))
    count = 0

    for combination in allcombinations:
        try:
            base_image = Image.open(os.path.join(directory, mainimage) + ".png").convert("RGBA")
            for key in combination:
                diffimagepath = os.path.join(directory, f"{key}.png")

                diffimage = Image.open(diffimagepath).convert("RGBA")
                diffimage = fixedge(diffimage)
                diffoffsetx, diffoffsety = offsets[key]

                offsetx = diffoffsetx - baseoffsetx
                offsety = diffoffsety - baseoffsety

                base_image.paste(diffimage, (offsetx, offsety), diffimage)
                diffimage.close()

            outputfilename = f"merge_{os.path.basename(directory)}_{count}.png"
            output = os.path.join(directory, outputfilename)
            base_image.save(output)
            base_image.close()
            count += 1
        except Exception as e:
            print(e)
    print(f"Merge {directory} Success")    

def main():
    parser = argparse.ArgumentParser(description="Useage: ATXDiffMerge.py -l <ATXfileExportdirectory>\n" \
                                                 "-l : single folder mode\n" \
                                                 "-b : multiple folders mode",
                                                 formatter_class=argparse.RawTextHelpFormatter)
    
    parser.add_argument("-l", "--single", action="store_true")
    parser.add_argument("-b", "--multiple", action="store_true")
    parser.add_argument("inputpath")

    args = parser.parse_args()

    if args.single:
        merge_images(args.inputpath)
        return
    
    if args.multiple:
        for dirpath, _, filenames in os.walk(args.inputpath):
            if 'info.json' in filenames and 'offset.json' in filenames:
                merge_images(dirpath)
        return
    
if __name__ == "__main__":
    main()
    
