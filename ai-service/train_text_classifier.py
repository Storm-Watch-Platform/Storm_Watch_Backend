# ai-service/train_text_classifier.py
from pathlib import Path
import pandas as pd
from collections import Counter
from sklearn.model_selection import train_test_split
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline
import joblib

ROOT = Path(__file__).resolve().parent
DATA_PATH = ROOT / "data" / "hazard_reports_train.csv"
MODELS_DIR = ROOT / "models"
MODELS_DIR.mkdir(parents=True, exist_ok=True)
MODEL_URGENCY_PATH = MODELS_DIR / "urgency_clf.pkl"
MODEL_TYPE_PATH    = MODELS_DIR / "type_clf.pkl"

def load_data() -> pd.DataFrame:
    if not DATA_PATH.exists():
        DATA_PATH.parent.mkdir(parents=True, exist_ok=True)
        DATA_PATH.write_text(
            "text,urgency_label,incident_type_label\n"
            "\"Nhà tôi ngập đến đầu gối, 3 người mắc kẹt\",HIGH,FLOOD\n"
            "\"Cây đổ chắn ngang đường, xe khó đi\",MEDIUM,TREE_DOWN\n"
            "\"Mưa nhỏ, vẫn đi lại được\",LOW,RAIN\n",
            encoding="utf-8"
        )
    df = pd.read_csv(DATA_PATH)
    return df.dropna(subset=["text", "urgency_label", "incident_type_label"])

def train_one_label(texts: pd.Series, labels: pd.Series, out_path: Path, test_size=0.2, random_state=42):
    y = labels.astype(str).values
    counts = Counter(y)
    min_cnt = min(counts.values())

    def can_strat(ts):  # mỗi lớp có >=2 và test mỗi lớp >=1
        from math import floor
        return min_cnt >= 2 and floor(min_cnt * ts) >= 1

    stratify_arg = y if can_strat(test_size) else None
    if stratify_arg is None and min_cnt >= 2:
        for ts in (0.15, 0.10):
            if can_strat(ts):
                test_size = ts; stratify_arg = y; break

    pipe = Pipeline([
        ("tfidf", TfidfVectorizer(max_features=5000)),
        ("clf", LogisticRegression(max_iter=1000))
    ])

    if stratify_arg is None:
        print("[WARN] Not enough samples per class; training on FULL data (no holdout).")
        pipe.fit(texts, y)
        joblib.dump(pipe, out_path)
        print(f"Saved (no holdout): {out_path}")
        return

    Xtr, Xte, ytr, yte = train_test_split(
        texts, y, test_size=test_size, random_state=random_state, stratify=stratify_arg
    )
    pipe.fit(Xtr, ytr)
    acc = pipe.score(Xte, yte)
    print(f"{out_path.name} accuracy: {acc:.3f}")
    joblib.dump(pipe, out_path)
    print(f"Saved: {out_path}")

def main():
    df = load_data()
    train_one_label(df["text"], df["urgency_label"], MODEL_URGENCY_PATH)
    train_one_label(df["text"], df["incident_type_label"], MODEL_TYPE_PATH)

if __name__ == "__main__":
    main()
