# ai-service/data_prep_seed.py
from datasets import load_dataset
import pandas as pd
import re
import os

# ---------- mapping rules ----------
def norm_incident(txt: str) -> str:
    t = (txt or "").lower()
    if re.search(r"\b(flood|flooding|inundat)", t): return "FLOOD"
    if re.search(r"\brain(s|ing)?\b|heavy rain|monsoon", t): return "RAIN"
    if re.search(r"landslide|mudslide", t): return "LANDSLIDE"
    if re.search(r"electr(ic|icity)|power line|downed line|short circuit|transformer", t): return "ELECTRIC"
    if re.search(r"tree (down|fell|fallen)|fallen tree|uprooted tree", t): return "TREE_DOWN"
    return "OTHER"

def infer_urgency(txt: str) -> str:
    t = (txt or "").lower()
    if re.search(r"trapped|cannot.*(leave|evacuate)|urgent|severe|critical|water.*(chest|waist|1m|meter)", t):
        return "HIGH"
    if re.search(r"blocked|damage|need help|help needed|stuck|power outage|road closed", t):
        return "MEDIUM"
    return "LOW"

def to_df(ds):
    rows=[]
    for r in ds:
        text = r.get("text") or r.get("tweet_text") or r.get("content")
        if not text: 
            continue
        rows.append({
            "text": text,
            "incident_type_label": norm_incident(text),
            "urgency_label": infer_urgency(text)
        })
    return pd.DataFrame(rows)

# ---------- load sources ----------
CFG = os.getenv("CRISISBENCH_CFG", "humanitarian")  # 'humanitarian' | 'informativeness'
df_list = []

# CrisisBench (yêu cầu config)
try:
    cb = load_dataset("QCRI/CrisisBench-all-lang", CFG, split="train")
except Exception as e1:
    print(f"[warn] CrisisBench({CFG}) failed: {e1}")
    alt = "informativeness" if CFG == "humanitarian" else "humanitarian"
    try:
        cb = load_dataset("QCRI/CrisisBench-all-lang", alt, split="train")
        print(f"[info] Fallback CrisisBench config -> {alt}")
    except Exception as e2:
        print(f"[warn] CrisisBench fallback failed: {e2}")
        # tùy chọn: dùng English-only
        cb = load_dataset("QCRI/CrisisBench-english", "humanitarian", split="train")
        print("[info] Using CrisisBench-english/humanitarian")
df_list.append(to_df(cb))

# HumAID (HF versions có text sẵn)
for repo in ["QCRI/HumAID-all", "QCRI/HumAID-events", "QCRI/HumAID-event-type"]:
    try:
        ds = load_dataset(repo, split="train")
        print(f"[info] Loaded {repo}")
        df_list.append(to_df(ds))
    except Exception as e:
        print(f"[warn] Skip {repo}: {e}")

# ---------- merge + balance ----------
df = pd.concat(df_list, ignore_index=True).dropna()
keep = {"FLOOD","RAIN","LANDSLIDE","ELECTRIC","TREE_DOWN","OTHER"}
df = df[df["incident_type_label"].isin(keep)].reset_index(drop=True)

# cân bằng nhẹ: 150–300 mẫu/lớp để train nhanh (tăng nếu máy khỏe)
per_class_max = int(os.getenv("PER_CLASS_MAX", "300"))
df_bal = (df.groupby("incident_type_label", group_keys=False)
            .apply(lambda g: g.sample(min(len(g), max(150, per_class_max)), random_state=42))
            .reset_index(drop=True))

# ---------- save ----------
out = "data/hazard_reports_train.csv"
df_bal.to_csv(out, index=False, encoding="utf-8")
print("Saved ->", out)
print(df_bal["incident_type_label"].value_counts())
print(df_bal["urgency_label"].value_counts())
