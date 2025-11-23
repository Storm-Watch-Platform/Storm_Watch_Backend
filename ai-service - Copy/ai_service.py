from __future__ import annotations

import os
from contextlib import asynccontextmanager
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Literal, Optional

import joblib
import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, field_validator

from translator import translate_vi_to_en  # dùng translator tiếng Việt -> tiếng Anh

# ---------------- Config / ENV ----------------
ROOT = Path(__file__).resolve().parent
MODEL_DIR = Path(os.getenv("MODEL_DIR", ROOT / "models"))
URG_PATH = MODEL_DIR / "urgency_clf.pkl"
TYP_PATH = MODEL_DIR / "type_clf.pkl"

# Ngưỡng "an toàn" CHO INCIDENT TYPE thôi (urgency luôn giữ LOW/MEDIUM/HIGH)
CONFIDENCE_MIN_TYPE = float(os.getenv("CONFIDENCE_MIN_TYPE", "0.65"))

_ALLOWED_INCIDENT = {
    "FLOOD", "RAIN", "LANDSLIDE", "ELECTRIC", "TREE_DOWN",
    "WIND", "STORM", "OTHER", "UNKNOWN"
}

# ---------------- Pydantic Models ----------------
class HazardReq(BaseModel):
    text: str = Field(min_length=1)


class HazardResp(BaseModel):
    urgency: Literal["LOW", "MEDIUM", "HIGH", "UNKNOWN"]
    incident_type: str
    confidence: float = Field(ge=0, le=1)


class PresenceUpdateReq(BaseModel):
    lat: float = Field(ge=-90, le=90)
    lon: float = Field(ge=-180, le=180)
    accuracy_m: Optional[float] = Field(default=None, ge=0)
    status: Literal["SAFE", "CAUTION", "DANGER", "UNKNOWN"] = "UNKNOWN"

    @field_validator("accuracy_m")
    @classmethod
    def check_accuracy(cls, v):
        if v is None:
            return v
        return min(v, 5000.0)  # 5km trần


class PresenceUpdateResp(BaseModel):
    ok: bool
    display_until: datetime


class SosRaiseReq(BaseModel):
    alert_body: str = Field(min_length=1)
    lat: float = Field(ge=-90, le=90)
    lon: float = Field(ge=-180, le=180)
    radius_m: int = Field(ge=100, le=5000)
    ttl_min: int = Field(ge=5, le=180)


class SosRaiseResp(BaseModel):
    ok: bool
    sos_id: str
    center: tuple[float, float]
    radius_m: int
    expires_at: datetime


# ---------------- Load models in lifespan ----------------
urgency_model = None
type_model = None


def _looks_non_english(s: str) -> bool:
    """Heuristic: nếu có ký tự Unicode > 127 (dấu tiếng Việt, emoji, v.v.) thì coi như non-English."""
    if not s:
        return False
    return any(ord(ch) > 127 for ch in s)


def _try_load(path: Path):
    try:
        m = joblib.load(path)
        print(f"[ai-service] Loaded model: {path}")
        return m
    except Exception as e:
        print(f"[ai-service] Failed to load {path}: {e}")
        return None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global urgency_model, type_model
    print("[lifespan] startup begin")
    urgency_model = _try_load(URG_PATH)
    type_model = _try_load(TYP_PATH)
    print("[lifespan] startup done")
    try:
        yield
    finally:
        print("[lifespan] shutdown")


app = FastAPI(title="StormSafe AI Service", version="1.2.0", lifespan=lifespan)

# ---------------- Utilities ----------------
def _final_classes(model) -> list[str]:
    """Lấy classes_ từ estimator cuối của Pipeline/CalibratedClassifierCV."""
    named = getattr(model, "named_steps", None)
    if named and "clf" in named and hasattr(named["clf"], "classes_"):
        return list(named["clf"].classes_)
    if hasattr(model, "classes_"):
        return list(model.classes_)
    raise RuntimeError("Cannot resolve classes_ from model")


def _normalize_incident(label: str) -> str:
    lab = (label or "").upper()
    return lab if lab in _ALLOWED_INCIDENT else "UNKNOWN"


# ---------------- Endpoints ----------------
@app.get("/health")
def health():
    return {"ok": bool(urgency_model) and bool(type_model)}


@app.post("/classify/hazard-text", response_model=HazardResp)
def classify_hazard_text(req: HazardReq):
    if urgency_model is None or type_model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")

    orig_text = req.text.strip()
    if not orig_text:
        raise HTTPException(status_code=400, detail="Empty text")

    # 0) Nếu là tiếng Việt / non-English -> dịch sang tiếng Anh cho model
    text_for_model = orig_text
    if _looks_non_english(orig_text):
        try:
            text_for_model = translate_vi_to_en(orig_text)
            if not text_for_model:
                text_for_model = orig_text  # fallback
        except Exception as e:
            print("[ai-service] translate_vi_to_en failed:", e)
            text_for_model = orig_text

    # 1) Dự đoán với xác suất đã calibrate (CalibratedClassifierCV)
    up = urgency_model.predict_proba([text_for_model])[0]
    tp = type_model.predict_proba([text_for_model])[0]

    u_classes = _final_classes(urgency_model)
    t_classes = _final_classes(type_model)
    u_idx, t_idx = int(np.argmax(up)), int(np.argmax(tp))
    u_label = str(u_classes[u_idx]) if u_classes else "UNKNOWN"
    t_label = _normalize_incident(str(t_classes[t_idx]) if t_classes else "UNKNOWN")

    u_max, t_max = float(up[u_idx]), float(tp[t_idx])
    conf = float(min(u_max, t_max))  # confidence chung

    # 2) Hybrid guard cho tiếng Việt (dựa trên orig_text)
    incident_overridden = False
    urgency_overridden = False

    if conf < 0.90:
        vi = orig_text.lower()

        # incident-type cues
        if any(k in vi for k in ["mưa lớn", "mưa to", "mưa dông", "mưa bão", "mưa"]):
            t_label = "RAIN"
            incident_overridden = True
        if any(k in vi for k in ["ngập", "ngập nước", "lũ", "lũ lụt", "nước dâng", "ngập đến"]):
            t_label = "FLOOD"
            incident_overridden = True
        if any(k in vi for k in ["mất điện", "cúp điện", "điện bị cắt", "đứt dây điện", "trạm biến áp"]):
            t_label = "ELECTRIC"
            incident_overridden = True
        if any(k in vi for k in ["cây đổ", "cây ngã", "cây bật gốc", "cây gãy"]):
            t_label = "TREE_DOWN"
            incident_overridden = True
        if any(k in vi for k in ["sạt lở", "sụt lún", "lở đất"]):
            t_label = "LANDSLIDE"
            incident_overridden = True

        # urgency cues
        if any(k in vi for k in ["ngập sâu", "không di chuyển được", "mắc kẹt", "khẩn cấp", "nguy hiểm"]):
            u_label = "HIGH"
            urgency_overridden = True
        elif any(k in vi for k in ["đường bị chặn", "cản trở", "hư hỏng", "cúp điện", "trơn trượt"]):
            u_label = "MEDIUM"
            urgency_overridden = True

    # 3) Áp ngưỡng tin cậy CHỈ cho incident_type
    #    Urgency: luôn giữ LOW/MEDIUM/HIGH (hoặc heuristic), không set UNKNOWN chỉ vì prob thấp.
    if (not incident_overridden) and t_max < CONFIDENCE_MIN_TYPE:
        t_label = "UNKNOWN"

    # Ép urgency về tập 4 mức
    if u_label not in {"LOW", "MEDIUM", "HIGH", "UNKNOWN"}:
        mapping = {"LOW": "LOW", "MED": "MEDIUM", "MID": "MEDIUM", "HI": "HIGH"}
        u_label = mapping.get(u_label.upper(), "UNKNOWN")

    return HazardResp(urgency=u_label, incident_type=t_label, confidence=conf)


@app.post("/presence/update", response_model=PresenceUpdateResp)
def presence_update(req: PresenceUpdateReq):
    display_until = datetime.now(timezone.utc) + timedelta(minutes=30)
    return PresenceUpdateResp(ok=True, display_until=display_until)


@app.post("/sos/raise", response_model=SosRaiseResp)
def sos_raise(req: SosRaiseReq):
    sos_id = f"sos_{abs(hash((req.alert_body, req.lat, req.lon, req.radius_m))) % 10_000_000}"
    expires_at = datetime.now(timezone.utc) + timedelta(minutes=req.ttl_min)
    return SosRaiseResp(
        ok=True,
        sos_id=sos_id,
        center=(req.lat, req.lon),
        radius_m=req.radius_m,
        expires_at=expires_at,
    )
