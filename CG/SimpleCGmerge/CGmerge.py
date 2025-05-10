import cv2
import os
import re
from pathlib import Path


def blend_images(base_img, diff_img):
    if base_img.shape[:2] != diff_img.shape[:2]:
        raise ValueError(f"图片分辨率不匹配！主图片: {base_img.shape[:2]}, 差分图片: {diff_img.shape[:2]}")

    if base_img.shape[2] == 3:
        base_img = cv2.cvtColor(base_img, cv2.COLOR_BGR2BGRA)

    if diff_img.shape[2] == 4:
        alpha = diff_img[:, :, 3] / 255.0
        for c in range(3):
            base_img[:, :, c] = (1 - alpha) * base_img[:, :, c] + alpha * diff_img[:, :, c]
    else:
        base_img[:, :, :3] = diff_img[:, :, :3]

    return base_img


def process_diff_folders(base_image_path, diff_root_dir, output_dir):
    base_img = cv2.imread(base_image_path, cv2.IMREAD_UNCHANGED)
    if base_img is None:
        raise ValueError(f"无法读取主图片: {base_image_path}")
    print(f"主图片初始 shape: {base_img.shape}")

    Path(output_dir).mkdir(parents=True, exist_ok=True)

    base_name = Path(base_image_path).stem

    diff_dirs = [d for d in os.listdir(diff_root_dir) if re.match(r'diff\d+', d)]
    diff_dirs.sort(key=lambda s: int(re.search(r'\d+', s).group()))

    counter = 0

    for diff_dir in diff_dirs:
        diff_dir_path = os.path.join(diff_root_dir, diff_dir)
        if not os.path.isdir(diff_dir_path):
            continue

        current_img = base_img.copy()

        diff_files = [f for f in os.listdir(diff_dir_path) if f.endswith(('.webp', '.png', '.jpg'))]
        for diff_file in diff_files:
            diff_path = os.path.join(diff_dir_path, diff_file)
            diff_img = cv2.imread(diff_path, cv2.IMREAD_UNCHANGED)

            current_img = blend_images(current_img, diff_img)

        output_name = f"{base_name}_{counter}.png"
        output_path = os.path.join(output_dir, output_name)
        cv2.imwrite(output_path, current_img)
        print(f"生成图片: {output_path}")
        counter += 1


if __name__ == "__main__":
    base_image_path = "./ae1cgcs06_a1.png"
    diff_root_dir = "./"  
    output_dir = "./"  

    process_diff_folders(base_image_path, diff_root_dir, output_dir)