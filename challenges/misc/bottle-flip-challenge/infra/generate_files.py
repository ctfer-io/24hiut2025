import os
import random
import argparse
from PIL import Image
import piexif
import zipfile

def main(flag):
    os.makedirs("challenge_images", exist_ok=True)

    # Pick a random index from 0 to 999
    flag_index = random.randint(0, 999)

    # Open the bottle.png image and convert it to RGB mode
    bottle_img = Image.open("bottle.png").convert("RGB")

    for i in range(1000):
        filename = f"24HIUT{{FreizhCola_Bottle_{i}}}.jpg"
        filepath = os.path.join("challenge_images", filename)

        if i == flag_index:
            exif_dict = {
                "0th": {piexif.ImageIFD.ImageDescription: flag.encode("utf-8")},
            }
            exif_bytes = piexif.dump(exif_dict)
            bottle_img.save(filepath, exif=exif_bytes)
        else:
            bottle_img.save(filepath)

    # Zip the challenge_images folder
    with zipfile.ZipFile("challenge_images.zip", "w", zipfile.ZIP_DEFLATED) as zipf:
        for root, dirs, files in os.walk("challenge_images"):
            for file in files:
                file_path = os.path.join(root, file)
                arcname = os.path.relpath(file_path, "challenge_images")
                zipf.write(file_path, arcname=os.path.join("challenge_images", arcname))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate images with a hidden flag.")
    parser.add_argument(
        "-s",
        "--secret",
        type=str,
        default="Upside_Down_Bottle_Found!",
        help="The secret flag to embed in one of the images.",
    )
    args = parser.parse_args()
    main(args.secret)
