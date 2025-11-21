# ai-service/train_calibrated_baseline.py
import pandas as pd
import numpy as np
import joblib
from collections import Counter
from typing import Tuple

from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.calibration import CalibratedClassifierCV

from imblearn.pipeline import Pipeline as ImbPipeline
from imblearn.over_sampling import RandomOverSampler

# --------- Paths ----------
DATA_PATH = "data/hazard_reports_train.csv"
MODEL_URGENCY = "models/urgency_clf.pkl"
MODEL_TYPE    = "models/type_clf.pkl"

# --------- Split/Validate helpers ----------
MIN_PER_CLASS = 2
TEST_SIZE     = 0.2
RANDOM_STATE  = 42

def print_counts(tag: str, y):
    c = Counter(y)
    print(f"[{tag}] class counts:", dict(c))
    return c

def safe_split(X, y, test_size=TEST_SIZE, stratify=True) -> Tuple:
    """Stratify chỉ khi mọi lớp có >= MIN_PER_CLASS; tự co test_size để có mẫu cho mỗi lớp."""
    counts = Counter(y)
    ok = all(v >= MIN_PER_CLASS for v in counts.values())
    if stratify and ok:
        min_c = min(counts.values())
        ts = test_size
        if min_c * ts < 1:
            ts = max(0.1, 1.0 / max(min_c, 2))
        return train_test_split(X, y, test_size=ts, random_state=RANDOM_STATE, stratify=y)
    return train_test_split(X, y, test_size=test_size, random_state=RANDOM_STATE, shuffle=True)

def min_class_count(y) -> int:
    return min(Counter(y).values())

# --------- Build components ----------
def build_base_pipeline() -> ImbPipeline:
    """
    TF-IDF (char_wb 3–5) -> RandomOverSampler('auto') -> LogisticRegression(saga)
    - ROS trong pipeline: chỉ resample khi fit (tránh leakage).
    """
    vect = TfidfVectorizer(
        analyzer="char_wb",      # n-gram ký tự trong biên từ (bền lỗi chính tả/dấu)
        ngram_range=(3, 5),
        max_features=30000
    )
    ros  = RandomOverSampler(sampling_strategy="auto", random_state=RANDOM_STATE)
    clf  = LogisticRegression(solver="saga", max_iter=1500, class_weight=None, n_jobs=None)
    return ImbPipeline(steps=[("tfidf", vect), ("ros", ros), ("clf", clf)])

def build_calibrated_model(pipe: ImbPipeline, y_train) -> CalibratedClassifierCV:
    """
    Hiệu chỉnh xác suất bằng Platt/sigmoid. Dùng CV nội bộ với cv = min(3, min_class_count).
    Nếu lớp quá ít, ép cv=2 (điều kiện tối thiểu cho CV).
    """
    mcc = min_class_count(y_train)
    if mcc < 2:
        raise ValueError("Train set has a class with <2 samples. Please add more labeled data.")

    cv = max(2, min(3, mcc))
    return CalibratedClassifierCV(estimator=pipe, method="sigmoid", cv=cv)

# --------- Train entry ----------
def train_and_save(target_col: str, out_path: str):
    # 1) Load
    df = pd.read_csv(DATA_PATH).dropna(subset=["text", target_col])
    X = df["text"].astype(str).values
    y = df[target_col].astype(str).values
    print_counts(f"{target_col}/ALL", y)

    if len(set(y)) < 2:
        raise ValueError(f"Dataset for {target_col} has only ONE class. Need at least 2 classes to train.")

    # 2) Split train/test (stratify an toàn)
    X_tr, X_te, y_tr, y_te = safe_split(X, y, test_size=TEST_SIZE, stratify=True)
    print_counts(f"{target_col}/TRAIN", y_tr)
    print_counts(f"{target_col}/TEST",  y_te)

    # 3) Build pipeline + calibration (CV nội bộ)
    base_pipe = build_base_pipeline()
    cal_model = build_calibrated_model(base_pipe, y_tr)

    # 4) Fit calibrated model (ROS chỉ chạy trong fit)
    cal_model.fit(X_tr, y_tr)

    # 5) Evaluate & save
    y_pred = cal_model.predict(X_te)
    print(f"\n[{target_col}] classification report (calibrated):\n",
          classification_report(y_te, y_pred, zero_division=0))
    joblib.dump(cal_model, out_path)
    print("Saved =>", out_path)

if __name__ == "__main__":
    train_and_save("urgency_label", MODEL_URGENCY)
    train_and_save("incident_type_label", MODEL_TYPE)
