from presidio_analyzer import AnalyzerEngine, PatternRecognizer, Pattern
from presidio_analyzer.nlp_engine import NlpEngineProvider

class PIIDetector:
    def __init__(self):
        # Configurar NLP engine con soporte para español
        configuration = {
            "nlp_engine_name": "spacy",
            "models": [
                {"lang_code": "es", "model_name": "es_core_news_sm"},
            ],
        }
        provider = NlpEngineProvider(nlp_configuration=configuration)
        nlp_engine = provider.create_engine()

        self.analyzer = AnalyzerEngine(nlp_engine=nlp_engine, supported_languages=["es"])
        self._add_spanish_recognizers()
    
    def _add_spanish_recognizers(self):
        # DNI español: 8 números + letra
        dni_pattern = Pattern(name="dni", regex=r"\b[0-9]{8}[A-Za-z]\b", score=0.95)
        dni_recognizer = PatternRecognizer(
            supported_entity="ES_DNI",
            patterns=[dni_pattern],
            supported_language="es"
        )
        
        # NIE español: X/Y/Z + 7 números + letra
        nie_pattern = Pattern(name="nie", regex=r"\b[XYZxyz][0-9]{7}[A-Za-z]\b", score=0.95)
        nie_recognizer = PatternRecognizer(
            supported_entity="ES_NIE",
            patterns=[nie_pattern],
            supported_language="es"
        )
        
        # IBAN español
        iban_pattern = Pattern(
            name="iban_es",
            regex=r"\bES[0-9]{2}[\s]?([0-9]{4}[\s]?){5}\b",
            score=0.95
        )
        iban_recognizer = PatternRecognizer(
            supported_entity="ES_IBAN",
            patterns=[iban_pattern],
            supported_language="es"
        )
        
        # NOTA: Teléfonos y emails NO se anonimizan porque son necesarios para análisis
        # El usuario puede pedir "¿Es seguro este número 612345678?" y necesitamos
        # que llegue al servicio de análisis sin anonimizar

        # Tarjeta de crédito (reforzado)
        card_pattern = Pattern(
            name="credit_card",
            regex=r"\b[0-9]{4}[\s.-]?[0-9]{4}[\s.-]?[0-9]{4}[\s.-]?[0-9]{4}\b",
            score=0.9
        )
        card_recognizer = PatternRecognizer(
            supported_entity="CREDIT_CARD_ES",
            patterns=[card_pattern],
            supported_language="es"
        )
        
        self.analyzer.registry.add_recognizer(dni_recognizer)
        self.analyzer.registry.add_recognizer(nie_recognizer)
        self.analyzer.registry.add_recognizer(iban_recognizer)
        self.analyzer.registry.add_recognizer(card_recognizer)
    
    def detect(self, text: str) -> list:
        """Detecta PII en el texto"""
        results = self.analyzer.analyze(
            text=text,
            language="es",
            entities=[
                # Estándar Presidio
                "PERSON",
                "CREDIT_CARD",
                "IBAN_CODE",
                # NOTA: LOCATION excluido porque genera muchos falsos positivos
                # y no es información crítica para un chat anti-fraude
                # Custom españoles
                "ES_DNI",
                "ES_NIE",
                "ES_IBAN",
                "CREDIT_CARD_ES",
                # NOTA: EMAIL_ADDRESS, PHONE_NUMBER y ES_PHONE excluidos
                # porque son necesarios para análisis de fraude
            ]
        )
        return results


# Singleton
pii_detector = PIIDetector()
