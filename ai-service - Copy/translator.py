# ai-service/translator.py
import os
import requests

PAPAGO_URL = "https://papago.apigw.ntruss.com/nmt/v1/translation"

PAPAGO_KEY_ID = os.getenv("4e1dl8i4pq")
PAPAGO_KEY = os.getenv("fZWODSGPzoH59AbGC2YGgIysueT6KzdwQIKSuVVf")


def translate_vi_to_en(text: str) -> str:
    """
    Dịch TV -> EN bằng Naver Papago.
    Nếu lỗi thì trả lại original text (fail-safe).
    """
    if not text:
        return text

    if PAPAGO_KEY_ID is None or PAPAGO_KEY is None:
        # Nếu chưa set env thì thôi, trả nguyên văn
        print("[translator] PAPAGO keys not set, skip translation")
        return text

    headers = {
        "X-NCP-APIGW-API-KEY-ID": PAPAGO_KEY_ID,
        "X-NCP-APIGW-API-KEY": PAPAGO_KEY,
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
    }
    data = {
        "source": "vi",
        "target": "en",
        "text": text,
    }

    try:
        resp = requests.post(PAPAGO_URL, headers=headers, data=data, timeout=3)
        resp.raise_for_status()
        j = resp.json()
        # cấu trúc chuẩn của Papago
        return j["message"]["result"]["translatedText"]
    except Exception as e:
        print("[translator] Papago error:", e)
        return text
