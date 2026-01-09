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
        # URLs sin protocolo pero con www
        r'(?:www\.)[^\s<>"\']+',
        # URL shorteners
        r'(?:bit\.ly|tinyurl\.com|t\.co|goo\.gl|ow\.ly|is\.gd|buff\.ly)/[^\s<>"\']+',
        # Dominios sin protocolo (ej: dgt-incidencias.es, bbva-compra.com)
        r'\b[\w][\w-]*\.(?:es|com|org|net|info|tk|xyz|io|co|eu|me|tv|cc|gob\.es|com\.es|org\.es)\b',
    ],
    EntityType.EMAIL: [
        r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}',
    ],
    EntityType.PHONE: [
        # Español con +34 (formatos: +34 612 345 678, +34 612 34 56 78, +34612345678)
        r'\+34[\s.-]?[6789](?:[\s.-]?[0-9]){8}',
        # Español sin código país (formatos: 612 345 678, 612 34 56 78, 612345678)
        r'\b[6789](?:[\s.-]?[0-9]){8}\b',
        # Con paréntesis
        r'\(\+?34\)[\s.-]?[6789](?:[\s.-]?[0-9]){8}',
    ]
}


def extract_entities(text: str) -> list[Entity]:
    """Extrae URLs, emails y teléfonos del texto"""
    entities = []

    # Primero extraer emails para identificar sus dominios
    email_domains = set()
    email_pattern = PATTERNS[EntityType.EMAIL][0]
    for match in re.finditer(email_pattern, text, re.IGNORECASE):
        email = match.group().strip()
        # Extraer el dominio del email (parte después del @)
        domain = email.split('@')[1] if '@' in email else ''
        if domain:
            email_domains.add(domain.lower())
        entities.append(Entity(
            type=EntityType.EMAIL,
            value=email,
            start=match.start(),
            end=match.end()
        ))

    # Luego extraer URLs y teléfonos
    for entity_type, patterns in PATTERNS.items():
        if entity_type == EntityType.EMAIL:
            continue  # Ya procesados arriba

        for pattern in patterns:
            for match in re.finditer(pattern, text, re.IGNORECASE):
                # Limpiar el valor
                value = match.group().strip()

                # Evitar duplicados (mismo valor ya extraído)
                if any(e.value == value for e in entities):
                    continue

                # Si es una URL, verificar que no sea el dominio de un email
                if entity_type == EntityType.URL:
                    # Extraer solo el dominio de la URL para comparar
                    url_domain = value.lower()
                    if url_domain.startswith('http'):
                        # Extraer dominio de URL completa
                        url_domain = url_domain.split('//')[1].split('/')[0] if '//' in url_domain else url_domain
                    # Quitar www. si existe
                    url_domain = url_domain.replace('www.', '')

                    # Si el dominio es parte de un email, ignorar
                    if url_domain in email_domains:
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
