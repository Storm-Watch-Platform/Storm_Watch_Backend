# ai-service/prepare_noaa_dataset.py
import pandas as pd
import numpy as np
from pathlib import Path

RAW_DIR   = Path("data/raw")                     # Thư mục chứa tất cả file raw
OUT_PATH  = Path("data/hazard_reports_train.csv")

def parse_damage(val: str) -> float:
    """
    '15K' -> 15000, '2.5M' -> 2_500_000, '0' or '' -> 0
    """
    if pd.isna(val):
        return 0.0
    s = str(val).strip()
    if s == "" or s == "0":
        return 0.0
    try:
        if s[-1] in ["K", "M", "B"]:
            num = float(s[:-1])
            factor = {"K": 1e3, "M": 1e6, "B": 1e9}[s[-1]]
            return num * factor
        return float(s)
    except Exception:
        return 0.0

def make_urgency_label(row) -> str:
    """
    Quy tắc đơn giản (bạn có thể chỉnh sau):

    - HIGH: có người chết hoặc injured hoặc damage lớn
    - MEDIUM: không chết nhưng có thiệt hại vừa
    - LOW: còn lại
    """
    deaths = row["DEATHS_DIRECT"] + row["DEATHS_INDIRECT"]
    inj    = row["INJURIES_DIRECT"] + row["INJURIES_INDIRECT"]
    dmg    = row["DAMAGE_TOTAL"]

    if deaths > 0 or inj > 10 or dmg >= 1_000_000:
        return "HIGH"
    if inj > 0 or dmg >= 50_000:
        return "MEDIUM"
    return "LOW"

def load_all_noaa_raw() -> pd.DataFrame:
    """
    Đọc tất cả các file .csv / .csv.gz trong data/raw
    và chỉ giữ các file có cột EVENT_TYPE (kiểu NOAA StormEvents).
    """
    if not RAW_DIR.exists():
        raise FileNotFoundError(f"RAW_DIR không tồn tại: {RAW_DIR.resolve()}")

    all_paths = list(RAW_DIR.glob("*.csv")) + list(RAW_DIR.glob("*.csv.gz"))

    if not all_paths:
        raise FileNotFoundError(f"Không tìm thấy file .csv hoặc .csv.gz nào trong {RAW_DIR.resolve()}")

    dfs = []
    print("Tìm thấy các file raw:")
    for p in all_paths:
        print(" -", p)
    print("\nĐang load từng file...")

    for path in all_paths:
        try:
            df = pd.read_csv(path, dtype=str, low_memory=False)
            if "EVENT_TYPE" not in df.columns:
                print(f"[SKIP] {path} vì không có cột EVENT_TYPE (không phải file NOAA details).")
                continue
            df["__SOURCE_FILE"] = str(path.name)  # lưu xem record từ file nào
            dfs.append(df)
            print(f"[OK] Loaded {path}, shape = {df.shape}")
        except Exception as e:
            print(f"[ERROR] Không đọc được {path}: {e}")

    if not dfs:
        raise RuntimeError("Không có file NOAA hợp lệ (có EVENT_TYPE) trong data/raw.")

    big_df = pd.concat(dfs, ignore_index=True)
    print("\nTổng hợp tất cả NOAA raw, shape =", big_df.shape)
    return big_df

def main():
    print("Loading ALL NOAA raw CSVs in data/raw ...")
    df = load_all_noaa_raw()

    # ---- Chuyển các cột numeric ----
    for col in ["DEATHS_DIRECT", "DEATHS_INDIRECT",
                "INJURIES_DIRECT", "INJURIES_INDIRECT"]:
        if col in df.columns:
            df[col] = pd.to_numeric(df[col], errors="coerce").fillna(0).astype(int)
        else:
            df[col] = 0

    # Damage property/crops -> tiền
    for col in ["DAMAGE_PROPERTY", "DAMAGE_CROPS"]:
        if col in df.columns:
            df[col] = df[col].apply(parse_damage)
        else:
            df[col] = 0.0

    df["DAMAGE_TOTAL"] = df["DAMAGE_PROPERTY"] + df["DAMAGE_CROPS"]

    # ---- Tạo text: ưu tiên narrative; nếu thiếu thì fallback EVENT_TYPE ----
    text_cols = []
    for col in ["EVENT_NARRATIVE", "EPISODE_NARRATIVE"]:
        if col in df.columns:
            text_cols.append(col)

    if text_cols:
        # Gộp theo từng hàng: "EPISODE_NARRATIVE . EVENT_NARRATIVE"
        df["text"] = df[text_cols].fillna("").agg(" . ".join, axis=1)
    else:
        # fallback: nếu file thiếu narrative, dùng EVENT_TYPE
        df["text"] = df["EVENT_TYPE"].astype(str)

    # ---- Incident type label = EVENT_TYPE ----
    df["incident_type_label"] = df["EVENT_TYPE"].astype(str)

    # ---- Urgency label ----
    df["urgency_label"] = df.apply(make_urgency_label, axis=1)

    # ---- Chỉ giữ 3 cột chính + lọc text quá ngắn ----
    out = df[["text", "urgency_label", "incident_type_label"]].dropna()
    out = out[out["text"].str.len() > 20]

    print("\nSample counts (urgency):")
    print(out["urgency_label"].value_counts())

    print("\nSample top 10 incident types:")
    print(out["incident_type_label"].value_counts().head(10))

    OUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    out.to_csv(OUT_PATH, index=False)
    print("\nSaved =>", OUT_PATH)

if __name__ == "__main__":
    main()
