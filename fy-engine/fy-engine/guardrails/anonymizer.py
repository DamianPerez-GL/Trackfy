from presidio_anonymizer import AnonymizerEngine
from presidio_anonymizer.entities import OperatorConfig
from .pii_detector import pii_detector


class PIIAnonymizer:
    def __init__(self):
        self.anonymizer = AnonymizerEngine()
    
    def anonymize(self, text: str) -> tuple[str, dict, bool]:
        """
        Anonimiza PII del texto.
        
        Returns:
            - texto_anonimizado: str
            - pii_map: dict con mapeo placeholder -> valor real
            - had_pii: bool si se encontró PII
        """
        results = pii_detector.detect(text)
        
        if not results:
            return text, {}, False
        
        # Crear mapeo antes de anonimizar
        pii_map = {}
        sorted_results = sorted(results, key=lambda x: x.start, reverse=True)
        
        anonymized_text = text
        for i, result in enumerate(sorted_results):
            entity_type = result.entity_type
            original_value = text[result.start:result.end]
            placeholder = f"[{entity_type}_{i}]"
            
            pii_map[placeholder] = original_value
            anonymized_text = (
                anonymized_text[:result.start] + 
                placeholder + 
                anonymized_text[result.end:]
            )
        
        return anonymized_text, pii_map, True
    
    def check_output(self, text: str) -> tuple[bool, list]:
        """
        Verifica que el output no contenga PII.
        
        Returns:
            - is_safe: bool
            - entities_found: list de entidades encontradas
        """
        results = pii_detector.detect(text)
        
        if not results:
            return True, []
        
        entities_found = [
            {"type": r.entity_type, "score": r.score}
            for r in results
        ]
        
        return False, entities_found


# Singleton
anonymizer = PIIAnonymizer()


# Funciones helper para usar directamente
def anonymize_input(text: str) -> tuple[str, dict, bool]:
    """Anonimiza PII del input del usuario"""
    return anonymizer.anonymize(text)


def verify_output(text: str) -> tuple[bool, list]:
    """Verifica que el output no tenga PII"""
    return anonymizer.check_output(text)
