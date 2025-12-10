import httpx
from config import ANALYSIS_SERVICE_URL


async def analyze_entities(entities: dict) -> dict | None:
    """
    Llama al servicio de análisis con las entidades extraídas.
    
    Args:
        entities: {
            "urls": ["https://..."],
            "emails": ["test@example.com"],
            "phones": ["+34612345678"]
        }
    
    Returns:
        {
            "type": "url" | "email" | "phone",
            "content": str,
            "risk_score": 0-100,
            "verdict": "safe" | "suspicious" | "dangerous",
            "reasons": ["razón 1", "razón 2"],
        }
    """
    # Prioridad: URLs > Emails > Phones
    if entities.get("urls"):
        return await analyze_url(entities["urls"][0])
    elif entities.get("emails"):
        return await analyze_email(entities["emails"][0])
    elif entities.get("phones"):
        return await analyze_phone(entities["phones"][0])
    
    return None


async def analyze_url(url: str) -> dict:
    """Analiza una URL"""
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(
                f"{ANALYSIS_SERVICE_URL}/analyze/url",
                json={"url": url}
            )
            
            if response.status_code == 200:
                return response.json()
            
    except Exception as e:
        print(f"Error calling analysis service: {e}")
    
    # Fallback si el servicio no responde
    return {
        "type": "url",
        "content": url,
        "risk_score": 50,
        "verdict": "unknown",
        "reasons": ["No se pudo analizar el enlace. Procede con precaución."]
    }


async def analyze_email(email: str) -> dict:
    """Analiza un email"""
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(
                f"{ANALYSIS_SERVICE_URL}/analyze/email",
                json={"email": email}
            )
            
            if response.status_code == 200:
                return response.json()
            
    except Exception as e:
        print(f"Error calling analysis service: {e}")
    
    return {
        "type": "email",
        "content": email,
        "risk_score": 50,
        "verdict": "unknown",
        "reasons": ["No se pudo verificar el email."]
    }


async def analyze_phone(phone: str) -> dict:
    """Analiza un teléfono"""
    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(
                f"{ANALYSIS_SERVICE_URL}/analyze/phone",
                json={"phone": phone}
            )
            
            if response.status_code == 200:
                return response.json()
            
    except Exception as e:
        print(f"Error calling analysis service: {e}")
    
    return {
        "type": "phone",
        "content": phone,
        "risk_score": 50,
        "verdict": "unknown",
        "reasons": ["No se pudo verificar el número."]
    }
