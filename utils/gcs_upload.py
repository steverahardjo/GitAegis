"""
Download, Resize, and Upload Image to Google Cloud Storage
This script downloads an image from a given google drive link, resizes it to specified dimensions, and uploads it to a specified Google Cloud Storage bucket.
Currently, the Image size is capped in 1 MB
Date: 12/11/2025
file link: https://colab.research.google.com/drive/1dOAjhGPzEyXihuCXMjaauiONM5ylWUpn?usp=sharing
requirements:
- gcs.json credentials file for GCS access
- ppi_photo.csv file with image links and metadata

Changes needed:
- Update gcs if needed
- May need replacement of Google Cloud Storage into other services
- Adjust image size to enable faster load.
"""
import pandas as pd
import gdown
from google.cloud import storage
from google.oauth2 import service_account
import io
import requests
from PIL import Image
import os

creds = service_account.Credentials.from_service_account_file("/content/gcs.json")
client = storage.Client(credentials=creds)

df = pd.read_csv("/content/ppi_photo.csv")
df.columns = df.iloc[0]  # set first row as header
df = df[1:]               # remove the header row from data
df.columns.values[1] = "Name"

# Add Rank column
df['Rank'] = range(1, len(df) + 1)

# --- GDRIVE FILE ID EXTRACT ---
def extract_gdrive_id(url: str):
    match = re.search(r'/d/([^/]+)', str(url))
    return match.group(1) if match else None

TARGET_WIDTH = 450
TARGET_HEIGHT = 600
QUALITY = 80

def download_gdrive_file(file_id, save_name, output_dir="output_images",
                         width=TARGET_WIDTH, height=TARGET_HEIGHT, quality=QUALITY):
    """
    Download a Google Drive image, resize it, convert to WebP, and save locally.

    Args:
        file_id (str): Google Drive file ID
        save_name (str): Name of the output file (without extension)
        output_dir (str): Directory to save WebP images
        width (int): Target width
        height (int): Target height
        quality (int): WebP quality (0-100)

    Returns:
        str: Path to saved WebP file, or None if failed
    """
    os.makedirs(output_dir, exist_ok=True)
    url = f"https://drive.google.com/uc?export=download&id={file_id}"

    r = requests.get(url, stream=True)
    if r.status_code != 200:
        print(f"[ERROR] Failed to download {file_id}, status {r.status_code}")
        return None

    # Load image into memory
    file_bytes = io.BytesIO()
    for chunk in r.iter_content(2048):
        file_bytes.write(chunk)
    file_bytes.seek(0)

    try:
        with Image.open(file_bytes) as img:
            # Resize and convert
            img = img.resize((width, height), Image.LANCZOS)
            if img.mode != "RGB":
                img = img.convert("RGB")

            webp_path = os.path.join(output_dir, f"{save_name}.webp")
            img.save(webp_path, "WEBP", quality=quality)
            print(f"[INFO] Saved WebP: {webp_path}")
            return webp_path
    except Exception as e:
        print(f"[ERROR] Failed to process image {file_id}: {e}")
        return None

def extract_gdrive_id(url: str) -> str | None:
    """Extract Google Drive file ID from a URL."""
    if not isinstance(url, str):
        return None
    if "id=" in url:
        return url.split("id=")[1].split("&")[0]
    elif "drive.google.com/file/d/" in url:
        return url.split("/file/d/")[1].split("/")[0]
    return None


def download_gdrive_file(file_id, save_name, output_dir="output_images", quality=60, max_size=1280):
    """
    Download a Google Drive image, compress, and convert to WebP.
    Perfect for Next.js optimization.
    """
    os.makedirs(output_dir, exist_ok=True)
    url = f"https://drive.google.com/uc?export=download&id={file_id}"
    local_path = os.path.join(output_dir, f"{save_name}.webp")

    try:
        response = requests.get(url, stream=True, timeout=20)
        response.raise_for_status()

        # Load image from memory
        img = Image.open(BytesIO(response.content)).convert("RGB")

        # Resize if image is too large
        w, h = img.size
        if max(w, h) > max_size:
            scale = max_size / max(w, h)
            new_size = (int(w * scale), int(h * scale))
            img = img.resize(new_size, Image.LANCZOS)

        # Save optimized WebP
        img.save(local_path, "WEBP", quality=quality, method=6, optimize=True)
        print(f"[OK] Saved {save_name}.webp ({os.path.getsize(local_path)//1024} KB)")
        return local_path

    except Exception as e:
        print(f"[ERR] Failed to download/convert {save_name}: {e}")
        return None


def upload_to_gcs(client, bucket_name, gcs_path, local_path, metadata=None):
    """Upload local file to GCS with optional metadata."""
    bucket = client.bucket(bucket_name)
    blob = bucket.blob(gcs_path)

    if metadata:
        blob.metadata = {k: str(v) for k, v in metadata.items() if v is not None}

    blob.upload_from_filename(local_path)
    print(f"[UPLOADED] gs://{bucket_name}/{gcs_path}")
    return f"gs://{bucket_name}/{gcs_path}"


# ==============================
# Main Script
# ==============================

def main():
    # 1. Load CSV
    df = pd.read_csv("/content/ppi_photo.csv")
    df.columns = df.iloc[0]
    df = df[1:]
    df.columns.values[1] = "Name"
    df["Rank"] = range(1, len(df) + 1)

    # 2. Initialize GCS client
    creds = service_account.Credentials.from_service_account_file("/content/gcs.json")
    gcs_client = storage.Client(credentials=creds)
    bucket_name = "bucket-image-about-us-ppi"

    # 3. Directory for temporary images
    output_dir = "output_images"

    # 4. Process each row
    for _, row in df.iterrows():
        name = row.get("Name")
        gdrive_url = row.get("Photo Link")

        if not gdrive_url:
            print(f"[WARN] No image URL for {name}, skipping.")
            continue

        file_id = extract_gdrive_id(gdrive_url)
        if not file_id:
            print(f"[WARN] Could not extract file ID for {name}, skipping.")
            continue

        # Download + compress to WebP
        local_path = download_gdrive_file(file_id, f"{row['Rank']}_{name}", output_dir)
        if not local_path:
            continue

        # Metadata for GCS
        metadata = {
            "Name": row.get("Name", ""),
            "Division": row.get("Division", ""),
            "School": row.get("School", ""),
            "Role": row.get("Executive", ""),
            "Rank": str(row.get("Rank", ""))
        }

        # Upload to GCS
        upload_to_gcs(
            gcs_client,
            bucket_name,
            f"photos/{os.path.basename(local_path)}",
            local_path,
            metadata
        )


if __name__ == "__main__":
    main()
