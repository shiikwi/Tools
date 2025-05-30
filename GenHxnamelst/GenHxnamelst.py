import os

"""
The tool is used to format .log file extracted from https://github.com/YeLikesss/KrkrExtractForCxdecV2
Put .log file in the directory then run py, it will generate HxName.lst for GARbro
"""

def format_file(input_files, output_file):
    results = []

    if isinstance(input_files, str):
        input_files = [input_files]

    for files in input_files:
        try:
            with open(files,'r', encoding="utf-16") as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue

                    parts = line.split("##YSig##")

                    if len(parts) == 2:
                        results.append(f"{parts[1]}:{parts[0]}")

        except Exception as e:
            print("Process file failed")

    if os.path.exists(output_file) and os.path.getsize(output_file) > 0:
        with open(output_file, 'a', encoding="utf8") as f:
            f.write("\n")
            f.write("\n".join(results))
            print("Append HxNames.lst Success")
    else:
        with open(output_file, 'w', encoding="utf8") as f:
            f.write("\n".join(results))
            print("Generate HxNames.lst Success")


if __name__ == '__main__':
    file_to_process = []
    output_file = "HxNames.lst"

    if os.path.exists("DirectoryHash.log"):
        file_to_process.append("DirectoryHash.log")

    if os.path.exists("FileNameHash.log"):
        file_to_process.append("FileNameHash.log")

    format_file(file_to_process, output_file)