from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional

from guardrails import anonymize_input, verify_output
from intent import classify_intent, needs_analysis, get_entities_for_analysis
from llm import generate_response
from services import analyze_entities

app = FastAPI(
    title="Fy Engine",
    description="Motor de lógica del asistente Fy de Trackfy",
    version="1.0.0"
)


# ============================================
# MODELOS
# ============================================

class ChatRequest(BaseModel):
    user_id: str
    message: str
    context: Optional[list[dict]] = None  # Historial previo


class ChatResponse(BaseModel):
    response: str
    mood: str
    pii_detected: bool
    intent: str
    analysis_performed: bool


# ============================================
# ENDPOINTS
# ============================================

@app.get("/health")
async def health():
    return {"status": "ok", "service": "fy-engine"}


@app.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest):
    """
    Endpoint principal de chat con Fy.
    
    Flujo:
    1. Guardrails INPUT: Anonimiza PII del mensaje
    2. Intent: Detecta qué quiere el usuario
    3. Si necesita análisis: Extrae entidades y llama al servicio
    4. LLM: Genera respuesta de Fy
    5. Guardrails OUTPUT: Verifica que no hay PII en respuesta
    """
    
    original_message = request.message
    
    # ─────────────────────────────────────────────
    # 1. GUARDRAILS INPUT: Anonimizar PII
    # ─────────────────────────────────────────────
    anonymized_message, pii_map, had_pii = anonymize_input(original_message)
    
    if had_pii:
        print(f"[Guardrails] PII detectado y anonimizado: {len(pii_map)} entidades")
    
    # ─────────────────────────────────────────────
    # 2. INTENT: Detectar qué quiere el usuario
    # ─────────────────────────────────────────────
    intent_result = classify_intent(original_message)  # Usar original para mejor detección
    intent = intent_result.intent.value
    
    print(f"[Intent] {intent} (confianza: {intent_result.confidence:.2f})")
    
    # ─────────────────────────────────────────────
    # 3. ANÁLISIS: Si lo necesita
    # ─────────────────────────────────────────────
    analysis_result = None
    analysis_performed = False
    
    if needs_analysis(intent_result):
        # Extraer entidades del mensaje ORIGINAL (URLs no son PII)
        entities = get_entities_for_analysis(original_message)
        
        if any(entities.values()):
            print(f"[Analysis] Entidades encontradas: {entities}")
            analysis_result = await analyze_entities(entities)
            analysis_performed = True
            print(f"[Analysis] Resultado: {analysis_result.get('verdict')} ({analysis_result.get('risk_score')}/100)")
    
    # ─────────────────────────────────────────────
    # 4. LLM: Generar respuesta de Fy
    # ─────────────────────────────────────────────
    llm_result = await generate_response(
        user_message=anonymized_message,  # Mensaje SIN PII al LLM
        intent=intent,
        context=request.context,
        analysis_result=analysis_result,
    )
    
    response_text = llm_result["response"]
    mood = llm_result["mood"]
    
    # ─────────────────────────────────────────────
    # 5. GUARDRAILS OUTPUT: Verificar respuesta
    # ─────────────────────────────────────────────
    is_safe, entities_found = verify_output(response_text)
    
    if not is_safe:
        print(f"[Guardrails] ⚠️ PII detectado en output: {entities_found}")
        # En producción podrías bloquear o limpiar la respuesta
        # Por ahora solo logueamos
    
    # ─────────────────────────────────────────────
    # 6. RESPUESTA
    # ─────────────────────────────────────────────
    
    # Añadir nota de privacidad si se detectó PII
    if had_pii:
        response_text += "\n\n🔒 _Tus datos personales no han sido compartidos._"
    
    return ChatResponse(
        response=response_text,
        mood=mood,
        pii_detected=had_pii,
        intent=intent,
        analysis_performed=analysis_performed,
    )



# ============================================
# RUN
# ============================================

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8082)
