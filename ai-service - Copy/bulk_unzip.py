import gzip
import shutil
from pathlib import Path

DATA_DIR = Path("data/raw/humaid")

for gz_file in DATA_DIR.glob("*.gz"):
    out_path = gz_file.with_suffix("")  # bỏ .gz
    print(f"Extracting {gz_file} -> {out_path}")
    with gzip.open(gz_file, "rb") as f_in, open(out_path, "wb") as f_out:
        shutil.copyfileobj(f_in, f_out)
