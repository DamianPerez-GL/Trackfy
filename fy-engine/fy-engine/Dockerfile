FROM python:3.11-slim

WORKDIR /app

# Instalar dependencias del sistema para spacy
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copiar requirements e instalar dependencias
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Descargar modelo de spacy para español (usando pip directamente)
RUN pip install https://github.com/explosion/spacy-models/releases/download/es_core_news_sm-3.7.0/es_core_news_sm-3.7.0-py3-none-any.whl

# Copiar código
COPY . .

# Puerto
EXPOSE 8082

# Comando
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8082"]
