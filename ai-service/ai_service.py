# ai-service/ai_service.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, field_validator
from typing import Literal, Optional
from pathlib import Path
import joblib
from datetime import datetime, timedelta, timezone
import math

# ---------- Models ----------
class HazardReq(BaseModel):
    text: str = Field(min_length=1)

class HazardResp(BaseModel):
    urgency: Literal["LOW","MEDIUM","HIGH"]
    incident_type: str
    confidence: float = Field(ge=0, le=1)

class PresenceUpdateReq(BaseModel):
    lat: float = Field(ge=-90, le=90)
    lon: float = Field(ge=-180, le=180)
    accuracy_m: Optional[float] = Field(default=None, ge=0)
    status: Literal["SAFE","CAUTION","DANGER","UNKNOWN"] = "UNKNOWN"

    @field_validator("accuracy_m")
    @classmethod
    def check_accuracy(cls, v):
        # nếu không có accuracy -> OK; nếu có mà quá tệ thì cảnh báo/clip
        if v is None: return v
        return min(v, 5000.0)

class PresenceUpdateResp(BaseModel):
    ok: bool
    display_until: datetime

class SosRaiseReq(BaseModel):
    alert_body: str = Field(min_length=1)
    lat: float = Field(ge=-90, le=90)
    lon: float = Field(ge=-180, le=180)
    radius_m: int = Field(ge=100, le=5000)  # giao diện gợi ý 2km → cho phép 100..5000m
    ttl_min: int = Field(ge=5, le=180)

class SosRaiseResp(BaseModel):
    ok: bool
    sos_id: str
    center: tuple[float, float]
    radius_m: int
    expires_at: datetime

# ---------- Load models ----------
ROOT = Path(__file__).resolve().parent
URG_PATH = ROOT / "models" / "urgency_clf.pkl"
TYP_PATH = ROOT / "models" / "type_clf.pkl"

def try_load(path: Path):
    try:
        return joblib.load(path)
    except Exception:
        return None

urgency_model = try_load(URG_PATH)
type_model    = try_load(TYP_PATH)

app = FastAPI(title="StormSafe AI Service", version="1.1.0")

@app.get("/health")
def health():
    return {"ok": urgency_model is not None and type_model is not None}

# ---------- Core: classify hazard text ----------
@app.post("/classify/hazard-text", response_model=HazardResp)
def classify_hazard_text(req: HazardReq):
    if urgency_model is None or type_model is None:
        raise HTTPException(status_code=503, detail="Model not loaded")
    text = req.text.strip()
    if not text:
        raise HTTPException(status_code=400, detail="Empty text")

    up = urgency_model.predict_proba([text])[0]
    tp = type_model.predict_proba([text])[0]

    u_label = str(urgency_model.classes_[up.argmax()])
    t_label = str(type_model.classes_[tp.argmax()])
    conf = float(min(float(up.max()), float(tp.max())))

    # ép về 3 mức LOW/MEDIUM/HIGH nếu training labels là 3 mức; nếu khác thì map tại đây.
    return HazardResp(urgency=u_label, incident_type=t_label, confidence=conf)

# ---------- NEW: presence/update ----------
@app.post("/presence/update", response_model=PresenceUpdateResp)
def presence_update(req: PresenceUpdateReq):
    # Ở AI service ta chỉ validate/chuẩn hóa; broadcast STOMP là nhiệm vụ BE.
    # Quy ước TTL hiển thị: 30 phút; có thể để client gửi.
    display_until = datetime.now(timezone.utc) + timedelta(minutes=30)
    return PresenceUpdateResp(ok=True, display_until=display_until)

# ---------- NEW: sos/raise ----------
@app.post("/sos/raise", response_model=SosRaiseResp)
def sos_raise(req: SosRaiseReq):
    # AI có thể chạy NLP nhanh trên alert_body (vd gắn cờ ELECTRIC) nếu cần.
    sos_id = f"sos_{abs(hash((req.alert_body, req.lat, req.lon, req.radius_m)))%10_000_000}"
    expires_at = datetime.now(timezone.utc) + timedelta(minutes=req.ttl_min)
    return SosRaiseResp(
        ok=True, sos_id=sos_id,
        center=(req.lat, req.lon),
        radius_m=req.radius_m,
        expires_at=expires_at
    )
