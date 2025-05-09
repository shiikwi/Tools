import os
import shutil

def copy_images_to_folder(a_folder, c_folder):

    if not os.path.exists(c_folder):
        os.makedirs(c_folder)


    for root, dirs, files in os.walk(a_folder):

        first_image_copied = False

        for file in files:

            if file.lower().endswith(('.png', '.jpg', '.jpeg', '.gif', '.bmp', '.tiff', '.dpak')):

                if not first_image_copied:
                    src_file_path = os.path.join(root, file)
                    dst_file_path = os.path.join(c_folder, file)


                    if os.path.exists(dst_file_path):
                        base, extension = os.path.splitext(file)
                        i = 1
                        while os.path.exists(dst_file_path):
                            dst_file_path = os.path.join(c_folder, f"{base}_{i}{extension}")
                            i += 1


                    shutil.copy2(src_file_path, dst_file_path)
                    print(f"Copied: {src_file_path} -> {dst_file_path}")


                    #first_image_copied = True

a_folder = ""
c_folder = ""
copy_images_to_folder(a_folder, c_folder)