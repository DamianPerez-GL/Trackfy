from openai import OpenAI
from config import OPENAI_API_KEY, OPENAI_MODEL
from .prompts import FY_SYSTEM_PROMPT, get_prompt_for_intent, get_mood_from_risk


client = OpenAI(api_key=OPENAI_API_KEY)


def _detect_missing_info_context(message: str) -> str:
    """
    Detecta qué tipo de información falta en el mensaje del usuario.
    Retorna un string describiendo el contexto para el prompt.
    """
    message_lower = message.lower()
    contexts = []

    # Detectar menciones de SMS/mensaje
    if any(word in message_lower for word in ["sms", "mensaje", "whatsapp", "texto"]):
        contexts.append("SMS o mensaje de texto (falta el contenido o número del remitente)")

    # Detectar menciones de llamada/teléfono
    if any(word in message_lower for word in ["llamaron", "llamada", "teléfono", "número", "llamó"]):
        contexts.append("llamada telefónica (falta el número)")

    # Detectar menciones de email/correo
    if any(word in message_lower for word in ["email", "correo", "mail", "gmail", "outlook"]):
        contexts.append("email o correo (falta el remitente o contenido)")

    # Detectar menciones de enlace/link/url
    if any(word in message_lower for word in ["enlace", "link", "url", "página", "web", "sitio"]):
        contexts.append("enlace o URL (falta la dirección web)")

    # Si no detectamos nada específico
    if not contexts:
        contexts.append("contacto sospechoso (falta el número, email o enlace para verificar)")

    return " / ".join(contexts)


async def generate_response(
    user_message: str,
    intent: str,
    context: list[dict] = None,
    analysis_result: dict = None,
) -> dict:
    """
    Genera una respuesta de Fy.
    
    Args:
        user_message: Mensaje del usuario (ya anonimizado)
        intent: Intent detectado (analysis, question, rescue, smalltalk)
        context: Historial de conversación previo
        analysis_result: Resultado del servicio de análisis (si aplica)
    
    Returns:
        {
            "response": str,
            "mood": str,  # happy, thinking, warning, danger
        }
    """
    messages = [
        {"role": "system", "content": FY_SYSTEM_PROMPT}
    ]
    
    # Añadir contexto previo si existe
    if context:
        messages.extend(context[-10:])  # Últimos 10 mensajes
    
    # Construir el mensaje según el intent
    if intent == "analysis" and analysis_result:
        # Formatear resultado del análisis
        reasons_text = "\n".join(f"- {r}" for r in analysis_result.get("reasons", []))
        
        prompt_addition = get_prompt_for_intent(
            intent="analysis",
            entity_type=analysis_result.get("type", "desconocido"),
            content=analysis_result.get("content", ""),
            risk_level=analysis_result.get("risk_score", 0),
            verdict=analysis_result.get("verdict", "desconocido"),
            reasons=reasons_text or "- Sin información adicional"
        )
        
        full_message = f"{user_message}\n\n{prompt_addition}"
        mood = get_mood_from_risk(analysis_result.get("risk_score", 0))
        
    elif intent == "rescue":
        prompt_addition = get_prompt_for_intent(
            intent="rescue",
            situation=user_message
        )
        full_message = f"{user_message}\n\n{prompt_addition}"
        mood = "danger"
        
    elif intent == "question":
        prompt_addition = get_prompt_for_intent(
            intent="question",
            topic=user_message
        )
        full_message = f"{user_message}\n\n{prompt_addition}"
        mood = "thinking"

    elif intent == "needs_info":
        # Detectar qué tipo de información falta basándose en el mensaje
        detected_context = _detect_missing_info_context(user_message)
        prompt_addition = get_prompt_for_intent(
            intent="needs_info",
            message=user_message,
            detected_context=detected_context
        )
        full_message = f"{user_message}\n\n{prompt_addition}"
        mood = "thinking"

    else:  # smalltalk
        prompt_addition = get_prompt_for_intent(
            intent="smalltalk",
            message=user_message
        )
        full_message = f"{user_message}\n\n{prompt_addition}"
        mood = "happy"
    
    messages.append({"role": "user", "content": full_message})

    # DEBUG: Ver lo que se manda al LLM
    print("\n" + "=" * 60)
    print("📤 ENVIANDO AL LLM:")
    print("=" * 60)
    for msg in messages:
        print(f"\n[{msg['role'].upper()}]:")
        print(msg['content'][:500] + "..." if len(msg['content']) > 500 else msg['content'])
    print("=" * 60 + "\n")

    # Llamar a GPT-4o mini
    try:
        response = client.chat.completions.create(
            model=OPENAI_MODEL,
            messages=messages,
            max_tokens=500,
            temperature=0.7,
        )
        
        response_text = response.choices[0].message.content
        
        return {
            "response": response_text,
            "mood": mood,
        }
        
    except Exception as e:
        # Fallback si falla el LLM
        return {
            "response": "Ups, algo ha fallado por mi parte. ¿Puedes intentarlo de nuevo? 🙏",
            "mood": "thinking",
            "error": str(e)
        }
