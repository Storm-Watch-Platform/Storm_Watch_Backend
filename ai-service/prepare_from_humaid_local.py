import pandas as pd, glob, os, re, json
from collections import Counter

BASE_DIR = "data/raw/humaid"                 # gốc HumAID
RAW_DIRS = [BASE_DIR, os.path.join(BASE_DIR, "events_set1")]
OUT_CSV  = "data/hazard_reports_train.csv"

TEXT_COL_CANDIDATES = [
    "tweet_text","text","content","full_text","tweet","message","body"
]

def norm_incident(t: str) -> str:
    s = (t or "").lower()
    if re.search(r"\b(flood|flooding|inundat|ngập|lũ)\b", s): return "FLOOD"
    if re.search(r"\brain(s|ing)?\b|heavy rain|monsoon|mưa to|mưa lớn|mưa\b", s): return "RAIN"
    if re.search(r"landslide|mudslide|sạt lở|lở đất", s): return "LANDSLIDE"
    if re.search(r"electr(ic|icity)|power line|downed line|short circuit|transformer|điện|cột điện|cháy nổ điện", s): 
        return "ELECTRIC"
    if re.search(r"tree (down|fell|fallen)|fallen tree|uprooted tree|cây đổ|cây ngã", s): 
        return "TREE_DOWN"
    return "OTHER"

def infer_urgency(t: str) -> str:
    s = (t or "").lower()
    if re.search(r"trapped|cannot.*(leave|evacuate)|urgent|severe|critical|"
                 r"nước.*(ngực|eo|1m|mét)|mắc kẹt|khẩn cấp|nguy hiểm", s):
        return "HIGH"
    if re.search(r"blocked|damage|need help|help needed|stuck|power outage|"
                 r"road closed|tắc đường|hư hại|cần cứu trợ|mất điện", s):
        return "MEDIUM"
    return "LOW"

def pick_text_col(df: pd.DataFrame):
    for c in TEXT_COL_CANDIDATES:
        if c in df.columns:
            return c
    # thử tìm cột có từ "text" trong tên
    for c in df.columns:
        if "text" in c.lower():
            return c
    return None

def read_any(fp: str) -> list[str]:
    ext = os.path.splitext(fp)[1].lower()
    texts = []

    if ext == ".csv":
        try:
            df = pd.read_csv(fp, encoding="utf-8", on_bad_lines="skip")
        except UnicodeDecodeError:
            df = pd.read_csv(fp, encoding="ISO-8859-1", on_bad_lines="skip")
        col = pick_text_col(df)
        if col: texts = df[col].astype(str).tolist()

    elif ext == ".tsv":
        df = pd.read_csv(fp, sep="\t", encoding="utf-8", on_bad_lines="skip")
        col = pick_text_col(df)
        if col: texts = df[col].astype(str).tolist()

    elif ext in (".jsonl", ".json"):
        # jsonl: mỗi dòng 1 json; json: list hoặc object lớn
        with open(fp, "r", encoding="utf-8", errors="ignore") as f:
            first = f.read(2); f.seek(0)
            if ext == ".jsonl" or first.strip().startswith("{") and "\n" in first:
                for line in f:
                    try:
                        obj = json.loads(line)
                        for c in TEXT_COL_CANDIDATES:
                            if c in obj:
                                texts.append(str(obj[c]))
                                break
                    except Exception:
                        continue
            else:
                try:
                    obj = json.load(f)
                    if isinstance(obj, list):
                        for o in obj:
                            if isinstance(o, dict):
                                for c in TEXT_COL_CANDIDATES:
                                    if c in o:
                                        texts.append(str(o[c])); break
                except Exception:
                    pass

    elif ext == ".txt":
        # một số bộ dump dạng txt: 1 tweet / dòng
        with open(fp, "r", encoding="utf-8", errors="ignore") as f:
            texts = [line.strip() for line in f if line.strip()]

    return texts

def main():
    # gom danh sách file các định dạng quan tâm
    patterns = ["**/*.csv","**/*.tsv","**/*.jsonl","**/*.json","**/*.txt"]
    files = []
    for rd in RAW_DIRS:
        for pat in patterns:
            files.extend(glob.glob(os.path.join(rd, pat), recursive=True))

    if not files:
        print(f"[warn] Không tìm thấy file dữ liệu trong: {RAW_DIRS}")
        return

    rows = []
    for fp in files:
        for tx in read_any(fp):
            it = norm_incident(tx)
            urg = infer_urgency(tx)
            rows.append({"text": tx, "incident_type_label": it, "urgency_label": urg})

    out = pd.DataFrame(rows).dropna()
    if out.empty:
        print("[warn] Không trích được text nào — kiểm tra lại định dạng/tên cột.")
        return

    # cân nhanh: tối đa 300 mẫu/lớp để train gọn
    out = (out.groupby("incident_type_label", group_keys=False)
              .apply(lambda g: g.sample(min(len(g), 300), random_state=42))
              .reset_index(drop=True))

    print("incident counts:", dict(Counter(out["incident_type_label"])))
    print("urgency counts:", dict(Counter(out["urgency_label"])))
    os.makedirs(os.path.dirname(OUT_CSV), exist_ok=True)
    out.to_csv(OUT_CSV, index=False, encoding="utf-8")
    print("Wrote ->", OUT_CSV, " | shape:", out.shape)

if __name__ == "__main__":
    main()
