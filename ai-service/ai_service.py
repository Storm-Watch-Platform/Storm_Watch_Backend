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

# ---------------- Config / ENV ----------------
ROOT = Path(__file__).resolve().parent
MODEL_DIR = Path(os.getenv("MODEL_DIR", ROOT / "models"))
URG_PATH = MODEL_DIR / "urgency_clf.pkl"
TYP_PATH = MODEL_DIR / "type_clf.pkl"

# Ngưỡng "an toàn" sau khi probabilities đã calibrate (Platt/Temp)
CONFIDENCE_MIN = float(os.getenv("CONFIDENCE_MIN", "0.65"))

_ALLOWED_INCIDENT = {
    "FLOOD", "RAIN", "LANDSLIDE", "ELECTRIC", "TREE_DOWN",
    "WIND", "STORM", "OTHER", "UNKNOWN"
}

# ---------------- Models ----------------
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
        # nếu không có accuracy -> OK; nếu có mà quá tệ thì clip
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
    radius_m: int = Field(ge=100, le=5000)  # UI gợi ý 2km nhưng cho phép 100..5000
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
        # dọn tài nguyên nếu cần
        print("[lifespan] shutdown")
        # (joblib models là in-memory; đặt None để GC)
        # Không cần đóng file descriptor vì đã load xong.
app = FastAPI(title="StormSafe AI Service", version="1.1.0", lifespan=lifespan)

# ---------------- Utilities ----------------
def _final_classes(model) -> list[str]:
    """
    Lấy classes_ từ estimator cuối của Pipeline/CalibratedClassifierCV.
    """
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

    text = req.text.strip()
    if not text:
        raise HTTPException(status_code=400, detail="Empty text")

    # 1) Dự đoán với xác suất đã calibrate (CalibratedClassifierCV)
    up = urgency_model.predict_proba([text])[0]
    tp = type_model.predict_proba([text])[0]

    u_classes = _final_classes(urgency_model)
    t_classes = _final_classes(type_model)
    u_idx, t_idx = int(np.argmax(up)), int(np.argmax(tp))
    u_label = str(u_classes[u_idx]) if u_classes else "UNKNOWN"
    t_label = _normalize_incident(str(t_classes[t_idx]) if t_classes else "UNKNOWN")

    u_max, t_max = float(up[u_idx]), float(tp[t_idx])
    conf = float(min(u_max, t_max))  # confidence chung

    # 2) Hybrid guard cho TV (chỉ nắn khi model chưa quá tự tin)
    if conf < 0.90:
        vi = text.lower()
        # incident-type cues
        if any(k in vi for k in ["mưa lớn", "mưa to", "mưa dông", "mưa bão", "mưa"]):
            t_label = "RAIN"
        if any(k in vi for k in ["ngập", "ngập nước", "lũ", "lũ lụt", "nước dâng", "ngập đến"]):
            t_label = "FLOOD"
        if any(k in vi for k in ["mất điện", "cúp điện", "điện bị cắt", "đứt dây điện", "trạm biến áp"]):
            t_label = "ELECTRIC"
        if any(k in vi for k in ["cây đổ", "cây ngã", "cây bật gốc", "cây gãy"]):
            t_label = "TREE_DOWN"
        if any(k in vi for k in ["sạt lở", "sụt lún", "lở đất"]):
            t_label = "LANDSLIDE"
        # urgency cues
        if any(k in vi for k in ["ngập sâu", "không di chuyển được", "mắc kẹt", "khẩn cấp", "nguy hiểm"]):
            u_label = "HIGH"
        elif any(k in vi for k in ["đường bị chặn", "cản trở", "hư hỏng", "cúp điện", "trơn trượt"]):
            u_label = "MEDIUM"

    # 3) Áp ngưỡng tin cậy sau calibration (tránh ảo giác)
    if u_max < CONFIDENCE_MIN:
        u_label = "UNKNOWN"
    if t_max < CONFIDENCE_MIN:
        t_label = "UNKNOWN"

    # Ép urgency về tập 4 mức
    if u_label not in {"LOW", "MEDIUM", "HIGH", "UNKNOWN"}:
        m = {"LOW": "LOW", "MED": "MEDIUM", "MID": "MEDIUM", "HI": "HIGH"}
        u_label = m.get(u_label.upper(), "UNKNOWN")

    return HazardResp(urgency=u_label, incident_type=t_label, confidence=conf)

@app.post("/presence/update", response_model=PresenceUpdateResp)
def presence_update(req: PresenceUpdateReq):
    # AI service chỉ validate/chuẩn hoá; broadcast/realtime do BE đảm nhiệm
    display_until = datetime.now(timezone.utc) + timedelta(minutes=30)
    return PresenceUpdateResp(ok=True, display_until=display_until)

@app.post("/sos/raise", response_model=SosRaiseResp)
def sos_raise(req: SosRaiseReq):
    # Có thể gắn NLP nhẹ trên alert_body nếu cần
    sos_id = f"sos_{abs(hash((req.alert_body, req.lat, req.lon, req.radius_m)))%10_000_000}"
    expires_at = datetime.now(timezone.utc) + timedelta(minutes=req.ttl_min)
    return SosRaiseResp(
        ok=True,
        sos_id=sos_id,
        center=(req.lat, req.lon),
        radius_m=req.radius_m,
        expires_at=expires_at,
    )