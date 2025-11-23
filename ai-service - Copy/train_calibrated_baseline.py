# ai-service/train_calibrated_baseline.py

import pandas as pd
from collections import Counter
from typing import Tuple

from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline

# --------- Paths ----------
DATA_PATH    = "data/hazard_reports_train.csv"
MODEL_URGENCY = "models/urgency_clf.pkl"
MODEL_TYPE    = "models/type_clf.pkl"

# --------- Split/Validate helpers ----------
TEST_SIZE     = 0.2
RANDOM_STATE  = 42

# Số mẫu tối đa sẽ dùng cho mỗi task (để train nhanh hơn)
MAX_TOTAL_URGENCY = 60000      # 3 lớp -> khoảng 20k mỗi lớp
MAX_TOTAL_TYPE    = 90000      # nhiều lớp hơn một chút

def print_counts(tag: str, y):
    c = Counter(y)
    print(f"[{tag}] class counts:", dict(c))
    return c

def safe_split(X, y, test_size=TEST_SIZE) -> Tuple:
    """Split train/test có stratify theo y."""
    return train_test_split(
        X, y,
        test_size=test_size,
        random_state=RANDOM_STATE,
        stratify=y
    )

def downsample_balanced(df: pd.DataFrame, target_col: str, max_total: int) -> pd.DataFrame:
    """
    Downsample balanced theo nhãn:
    - Tính số lớp K
    - Lấy tối đa floor(max_total / K) mẫu cho mỗi lớp
    """
    y = df[target_col]
    counts = y.value_counts()
    n_classes = len(counts)
    per_class = max_total // max(1, n_classes)

    if per_class == 0:
        # quá ít, thôi không downsample
        print(f"[{target_col}] max_total quá nhỏ, bỏ qua downsample.")
        return df

    orig_n = len(df)

    def sample_group(g):
        if len(g) <= per_class:
            return g
        return g.sample(n=per_class, random_state=RANDOM_STATE)

    df_small = df.groupby(target_col, group_keys=False).apply(sample_group)
    df_small = df_small.sample(frac=1.0, random_state=RANDOM_STATE).reset_index(drop=True)

    print(
        f"Downsample {target_col}: from {orig_n} to {len(df_small)} "
        f"(~{per_class} per class x {n_classes} classes)"
    )
    return df_small

# --------- Build pipeline ----------
def build_base_pipeline() -> Pipeline:
    """
    TF-IDF (word 1–2 gram) -> LogisticRegression
    Version nhẹ, không dùng char n-gram, không RandomOverSampler.
    """
    vect = TfidfVectorizer(
        analyzer="word",
        ngram_range=(1, 2),
        max_features=20000,
        min_df=3
    )
    clf  = LogisticRegression(
        solver="lbfgs",
        max_iter=300,
        n_jobs=-1,
        multi_class="auto"
    )
    return Pipeline(steps=[
        ("tfidf", vect),
        ("clf", clf),
    ])

# --------- Train entry ----------
def train_and_save(target_col: str, out_path: str, max_total: int):
    # 1) Load
    df = pd.read_csv(DATA_PATH).dropna(subset=["text", target_col])
    print_counts(f"{target_col}/ALL", df[target_col].astype(str))

    if len(df) > max_total:
        df = downsample_balanced(df, target_col, max_total)
    else:
        print(f"[{target_col}] Dataset nhỏ hơn max_total, giữ nguyên {len(df)} mẫu.")

    X = df["text"].astype(str).values
    y = df[target_col].astype(str).values

    if len(set(y)) < 2:
        raise ValueError(f"Dataset for {target_col} has only ONE class. Need at least 2 classes to train.")

    # 2) Split
    X_tr, X_te, y_tr, y_te = safe_split(X, y, test_size=TEST_SIZE)
    print_counts(f"{target_col}/TRAIN", y_tr)
    print_counts(f"{target_col}/TEST",  y_te)

    # 3) Build + fit (KHÔNG calibration, để train nhanh)
    pipe = build_base_pipeline()
    print(f"[{target_col}] Fitting model...")
    pipe.fit(X_tr, y_tr)

    # 4) Evaluate & save
    y_pred = pipe.predict(X_te)
    print(f"\n[{target_col}] classification report:\n",
          classification_report(y_te, y_pred, zero_division=0))

    import joblib
    joblib.dump(pipe, out_path)
    print("Saved =>", out_path)

if __name__ == "__main__":
    # Train urgency với ~60k mẫu
    train_and_save("urgency_label", MODEL_URGENCY, max_total=MAX_TOTAL_URGENCY)
    # Train incident_type với ~90k mẫu
    train_and_save("incident_type_label", MODEL_TYPE, max_total=MAX_TOTAL_TYPE)
