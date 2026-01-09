import re
from enum import Enum
from dataclasses import dataclass


class Intent(Enum):
    ANALYSIS = "analysis"       # Quiere analizar algo (URL, email, phone, mensaje)
    NEEDS_INFO = "needs_info"   # Menciona algo sospechoso pero NO incluye el dato
    QUESTION = "question"       # Pregunta sobre ciberseguridad
    RESCUE = "rescue"           # Emergencia, ha sido víctima
    REPORT = "report"           # Quiere reportar una estafa
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
            r"https?://",           # URLs con protocolo
            r"bit\.ly|tinyurl",     # URL shorteners
            r"@\w+\.\w+",           # Emails
            r"\+?34[\s.-]?[6789]",  # Teléfonos españoles con código país
            r"\b[6789][\s.-]?\d{2}[\s.-]?\d{2,3}[\s.-]?\d{2,3}\b",  # Teléfonos españoles sin código país
            r"\b[\w-]+\.(?:es|com|org|net|info|tk|xyz|gob\.es)\b",  # Dominios sin protocolo
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
    Intent.NEEDS_INFO: {
        # Usuario menciona algo sospechoso pero NO incluye URL/email/teléfono
        "keywords": [
            "me llegó un mensaje", "me llegó un sms", "me ha llegado",
            "recibí un mensaje", "recibí un sms", "recibí un email", "recibí un correo",
            "me llamaron", "me han llamado", "llamada de un número",
            "número desconocido", "número que no conozco", "número raro",
            "mensaje sospechoso", "sms sospechoso", "email sospechoso", "correo sospechoso",
            "mensaje raro", "sms raro", "email raro", "correo raro",
            "me escribieron", "me contactaron", "me mandaron algo",
            "no sé quién es", "no sé de quién es", "no reconozco",
            "dice que soy", "dicen que debo", "dice que tengo",
            "supuestamente de", "haciéndose pasar", "se hace pasar",
        ],
        "patterns": []  # Sin patrones de URL/email/phone, eso lo diferencia de ANALYSIS
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
    Intent.REPORT: {
        "keywords": [
            "reportar", "reportar estafa", "quiero reportar", "denunciar",
            "quiero denunciar", "reportar fraude", "avisar de una estafa",
            "informar de estafa", "reportar número", "reportar enlace",
            "reportar email", "reportar página",
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

    # Detectar si hay entidades analizables (URL, email, teléfono)
    has_analyzable_entity = False
    analysis_patterns = INTENT_PATTERNS[Intent.ANALYSIS]["patterns"]
    for pattern in analysis_patterns:
        if re.search(pattern, text, re.IGNORECASE):
            has_analyzable_entity = True
            break

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

    # Report tiene alta prioridad
    if scores[Intent.REPORT] > 0:
        return IntentResult(
            intent=Intent.REPORT,
            confidence=min(scores[Intent.REPORT] / 2.0, 1.0),
            trigger_words=triggers[Intent.REPORT]
        )

    # NEEDS_INFO: menciona algo sospechoso pero NO incluye URL/email/phone
    # Solo se activa si hay keywords de NEEDS_INFO pero NO hay entidades analizables
    if scores[Intent.NEEDS_INFO] > 0 and not has_analyzable_entity:
        return IntentResult(
            intent=Intent.NEEDS_INFO,
            confidence=min(scores[Intent.NEEDS_INFO] / 3.0, 1.0),
            trigger_words=triggers[Intent.NEEDS_INFO]
        )

    # Buscar el intent con mayor score (excluyendo NEEDS_INFO si ya lo procesamos)
    remaining_intents = {k: v for k, v in scores.items() if k != Intent.NEEDS_INFO}
    best_intent = max(remaining_intents, key=remaining_intents.get)
    best_score = remaining_intents[best_intent]

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
