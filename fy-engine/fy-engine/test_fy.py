"""
Script para probar el motor de Fy localmente.
Ejecuta: python test_fy.py
"""
import asyncio
from dotenv import load_dotenv

load_dotenv()

from guardrails import anonymize_input, verify_output
from intent import classify_intent, needs_analysis, get_entities_for_analysis
from llm import generate_response


async def test_fy(message: str):
    print("=" * 60)
    print(f"📝 MENSAJE: {message}")
    print("=" * 60)
    
    # 1. Guardrails INPUT
    anonymized, pii_map, had_pii = anonymize_input(message)
    print(f"\n🛡️ GUARDRAILS INPUT:")
    print(f"   PII detectado: {had_pii}")
    if had_pii:
        print(f"   Entidades: {list(pii_map.keys())}")
        print(f"   Mensaje anonimizado: {anonymized}")
    
    # 2. Intent
    intent_result = classify_intent(message)
    print(f"\n🎯 INTENT:")
    print(f"   Tipo: {intent_result.intent.value}")
    print(f"   Confianza: {intent_result.confidence:.2f}")
    print(f"   Triggers: {intent_result.trigger_words}")
    
    # 3. Entidades para análisis
    if needs_analysis(intent_result):
        entities = get_entities_for_analysis(message)
        print(f"\n🔍 ENTIDADES PARA ANÁLISIS:")
        print(f"   URLs: {entities['urls']}")
        print(f"   Emails: {entities['emails']}")
        print(f"   Phones: {entities['phones']}")
    
    # 4. Mock analysis result
    analysis_result = None
    if needs_analysis(intent_result):
        entities = get_entities_for_analysis(message)
        if entities.get("urls"):
            analysis_result = {
                "type": "url",
                "content": entities["urls"][0],
                "risk_score": 85,
                "verdict": "dangerous",
                "reasons": [
                    "URL acortada que oculta el destino real",
                    "Redirige a dominio sospechoso: santander-verificacion.tk",
                    "Dominio .tk gratuito, usado frecuentemente en phishing",
                    "Registrado hace solo 3 días"
                ]
            }
            print(f"\n⚠️ RESULTADO ANÁLISIS (mock):")
            print(f"   Riesgo: {analysis_result['risk_score']}/100")
            print(f"   Veredicto: {analysis_result['verdict']}")
    
    # 5. LLM
    print(f"\n🤖 LLAMANDO AL LLM...")
    result = await generate_response(
        user_message=anonymized,
        intent=intent_result.intent.value,
        context=None,
        analysis_result=analysis_result,
    )
    
    print(f"\n💬 RESPUESTA FY:")
    print(f"   Mood: {result['mood']}")
    print(f"   Texto:\n")
    print(result['response'])
    
    # 6. Guardrails OUTPUT
    is_safe, entities_found = verify_output(result['response'])
    print(f"\n🛡️ GUARDRAILS OUTPUT:")
    print(f"   Seguro: {is_safe}")
    if not is_safe:
        print(f"   ⚠️ PII en respuesta: {entities_found}")
    
    print("\n" + "=" * 60)


# Tests
async def main():
    # Test 1: Mensaje con PII y URL
    await test_fy(
        "Hola, me llamo Juan García y me llegó este SMS: "
        "Tu tarjeta 4532-1234-5678-9012 tiene un cargo. "
        "Verifica en bit.ly/santander-verificar. Mi DNI es 12345678A"
    )
    
    print("\n\n")
    
    # Test 2: Pregunta normal
    await test_fy("¿Qué es el phishing y cómo puedo protegerme?")
    
    print("\n\n")
    
    # Test 3: Rescate
    await test_fy("Creo que me han estafado, di mis datos de la tarjeta en una web")
    
    print("\n\n")
    
    # Test 4: Smalltalk
    await test_fy("Hola Fy, ¿qué tal estás?")


if __name__ == "__main__":
    asyncio.run(main())
