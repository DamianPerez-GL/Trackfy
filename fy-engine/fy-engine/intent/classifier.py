import re
from enum import Enum
from dataclasses import dataclass


class Intent(Enum):
    ANALYSIS = "analysis"       # Quiere analizar algo (URL, email, phone, mensaje)
    QUESTION = "question"       # Pregunta sobre ciberseguridad
    RESCUE = "rescue"           # Emergencia, ha sido víctima
    SMALLTALK = "smalltalk"     # Saludo, charla casual


@dataclass
class IntentResult:
    intent: Intent
    confidence: float
    trigger_words: list[str]


# Patrones por intent
INTENT_PATTERNS = {
    Intent.ANALYSIS: {
        "keywords": [
            "mira esto", "es seguro", "es legítimo", "me llegó", "me han enviado",
            "qué opinas de", "analiza", "verifica", "comprueba", "revisar",
            "este enlace", "esta url", "este link", "este mensaje", "este sms",
            "este email", "este correo", "este número", "este teléfono",
            "me llamaron", "me escribieron", "es real", "es falso", "es phishing",
            "es estafa", "parece sospechoso", "no me fío", "será verdad",
        ],
        "patterns": [
            r"https?://",           # URLs
            r"bit\.ly|tinyurl",     # URL shorteners
            r"@\w+\.\w+",           # Emails
            r"\+?34[\s.-]?[67]",    # Teléfonos españoles
        ]
    },
    Intent.RESCUE: {
        "keywords": [
            "me han estafado", "he sido víctima", "me robaron", "me engañaron",
            "di mis datos", "metí mi tarjeta", "puse mi contraseña",
            "he instalado", "descargué algo", "ayuda urgente", "emergencia",
            "qué hago ahora", "es tarde", "ya di", "ya puse", "ya metí",
            "creo que me han", "me hackearon",
        ],
        "patterns": []
    },
    Intent.QUESTION: {
        "keywords": [
            "qué es", "cómo funciona", "cómo puedo", "qué significa",
            "por qué", "explícame", "cuéntame", "dime", "qué hago si",
            "cómo sé si", "cómo protegerme", "es seguro usar", "recomiendas",
            "qué opinas", "consejos", "tips",
        ],
        "patterns": []
    },
    Intent.SMALLTALK: {
        "keywords": [
            "hola", "buenas", "hey", "qué tal", "cómo estás",
            "gracias", "vale", "ok", "perfecto", "genial",
            "adiós", "hasta luego", "chao",
        ],
        "patterns": []
    }
}


def classify_intent(text: str) -> IntentResult:
    """Clasifica el intent del mensaje del usuario"""
    text_lower = text.lower()
    scores = {intent: 0.0 for intent in Intent}
    triggers = {intent: [] for intent in Intent}
    
    for intent, patterns in INTENT_PATTERNS.items():
        # Check keywords
        for keyword in patterns["keywords"]:
            if keyword in text_lower:
                scores[intent] += 1.0
                triggers[intent].append(keyword)
        
        # Check regex patterns
        for pattern in patterns["patterns"]:
            if re.search(pattern, text, re.IGNORECASE):
                scores[intent] += 2.0  # Patterns tienen más peso
                triggers[intent].append(f"pattern:{pattern}")
    
    # Si hay URL/email/phone, casi seguro es análisis
    if scores[Intent.ANALYSIS] >= 2.0:
        return IntentResult(
            intent=Intent.ANALYSIS,
            confidence=min(scores[Intent.ANALYSIS] / 5.0, 1.0),
            trigger_words=triggers[Intent.ANALYSIS]
        )
    
    # Rescue tiene prioridad si hay match
    if scores[Intent.RESCUE] > 0:
        return IntentResult(
            intent=Intent.RESCUE,
            confidence=min(scores[Intent.RESCUE] / 3.0, 1.0),
            trigger_words=triggers[Intent.RESCUE]
        )
    
    # Buscar el intent con mayor score
    best_intent = max(scores, key=scores.get)
    best_score = scores[best_intent]
    
    # Si no hay match claro, default a question
    if best_score == 0:
        return IntentResult(
            intent=Intent.QUESTION,
            confidence=0.3,
            trigger_words=[]
        )
    
    return IntentResult(
        intent=best_intent,
        confidence=min(best_score / 3.0, 1.0),
        trigger_words=triggers[best_intent]
    )


def needs_analysis(intent_result: IntentResult) -> bool:
    """Determina si el intent requiere llamar al servicio de análisis"""
    return intent_result.intent == Intent.ANALYSIS
