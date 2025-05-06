import os
import json
from pathlib import Path
from collections import OrderedDict


def load_mapping(mapping_file):
    mapping = OrderedDict()
    with open(mapping_file, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if line and '\t' in line:
                hans, kanji = line.split('\t')[:2]
                if len(hans) == 1 and len(kanji) == 1:
                    mapping[hans] = kanji
    return mapping


def process_text(text, mapping):
    result = []
    used_mapping = OrderedDict()

    for char in text:
        if char in mapping:
            result.append(mapping[char])
            if char not in used_mapping:
                used_mapping[char] = mapping[char]
        else:
            result.append(char)

    return ''.join(result), used_mapping


def generate_report(output_encoding, total_files, used_mapping):
    return {
        "encoding_used": output_encoding,
        "total_files": total_files,
        "source_characters": ''.join(used_mapping.values()),
        "target_characters": ''.join(used_mapping.keys())
    }


def main(input_dir='input',
         mapping_file='hanzi2kanji_table.txt',
         input_encoding='utf-8',
         output_encoding='cp932'):

    output_dir = 'output'
    os.makedirs(output_dir, exist_ok=True)

    mapping = load_mapping(mapping_file)
    if not mapping:
        raise ValueError("映射表为空或格式错误")

    total_processed = 0
    global_used_mapping = OrderedDict()

    for input_file in Path(input_dir).glob('*.txt'):
        if '_converted' in input_file.stem:
            continue

        with open(input_file, 'r', encoding=input_encoding, errors='replace') as f:
            text = f.read()

        processed_text, file_used_mapping = process_text(text, mapping)

        for k, v in file_used_mapping.items():
            if k not in global_used_mapping:
                global_used_mapping[k] = v

        output_file = Path(output_dir) / f"{input_file.stem}_converted.txt"
        with open(output_file, 'w', encoding=output_encoding, errors='replace') as f:
            f.write(processed_text)

        total_processed += 1

    report = generate_report(
        output_encoding=output_encoding,
        total_files=total_processed,
        used_mapping=global_used_mapping
    )

    report_path = Path(output_dir) / 'Substitution.json'
    with open(report_path, 'w', encoding='utf-8') as f:
        json.dump(report, f, ensure_ascii=False, indent=2)

    print("\n" + "=" * 50)
    print(f"  SUCCESS！")
    print(f"• 处理文件数: {total_processed}")
    print(f"• 输出编码格式: {output_encoding}")
    print(f"• 映射字符对数: {len(global_used_mapping)}")
    print("=" * 50)


if __name__ == "__main__":
    main()
