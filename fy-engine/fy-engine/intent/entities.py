import re
from dataclasses import dataclass
from enum import Enum


class EntityType(Enum):
    URL = "url"
    EMAIL = "email"
    PHONE = "phone"


@dataclass
class Entity:
    type: EntityType
    value: str
    start: int
    end: int


# Patterns para extracción
PATTERNS = {
    EntityType.URL: [
        # URLs completas
        r'https?://[^\s<>"\']+',
        # URLs sin protocolo pero con dominio conocido
        r'(?:www\.)[^\s<>"\']+',
        # URL shorteners
        r'(?:bit\.ly|tinyurl\.com|t\.co|goo\.gl|ow\.ly|is\.gd|buff\.ly)/[^\s<>"\']+',
    ],
    EntityType.EMAIL: [
        r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}',
    ],
    EntityType.PHONE: [
        # Español con +34
        r'\+34[\s.-]?[6789][\s.-]?[0-9]{2}[\s.-]?[0-9]{3}[\s.-]?[0-9]{3}',
        # Español sin código país
        r'\b[6789][0-9]{2}[\s.-]?[0-9]{3}[\s.-]?[0-9]{3}\b',
        # Con paréntesis
        r'\(\+?34\)[\s.-]?[6789][0-9]{8}',
    ]
}


def extract_entities(text: str) -> list[Entity]:
    """Extrae URLs, emails y teléfonos del texto"""
    entities = []
    
    for entity_type, patterns in PATTERNS.items():
        for pattern in patterns:
            for match in re.finditer(pattern, text, re.IGNORECASE):
                # Limpiar el valor
                value = match.group().strip()
                
                # Evitar duplicados (mismo valor ya extraído)
                if any(e.value == value for e in entities):
                    continue
                
                entities.append(Entity(
                    type=entity_type,
                    value=value,
                    start=match.start(),
                    end=match.end()
                ))
    
    return entities


def extract_urls(text: str) -> list[str]:
    """Extrae solo URLs"""
    entities = extract_entities(text)
    return [e.value for e in entities if e.type == EntityType.URL]


def extract_emails(text: str) -> list[str]:
    """Extrae solo emails"""
    entities = extract_entities(text)
    return [e.value for e in entities if e.type == EntityType.EMAIL]


def extract_phones(text: str) -> list[str]:
    """Extrae solo teléfonos"""
    entities = extract_entities(text)
    return [e.value for e in entities if e.type == EntityType.PHONE]


def get_entities_for_analysis(text: str) -> dict:
    """
    Devuelve las entidades formateadas para enviar al servicio de análisis.
    
    Returns:
        {
            "urls": ["https://..."],
            "emails": ["test@example.com"],
            "phones": ["+34612345678"]
        }
    """
    entities = extract_entities(text)
    
    return {
        "urls": [e.value for e in entities if e.type == EntityType.URL],
        "emails": [e.value for e in entities if e.type == EntityType.EMAIL],
        "phones": [e.value for e in entities if e.type == EntityType.PHONE],
    }
