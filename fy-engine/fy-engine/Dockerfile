FROM python:3.11-slim

WORKDIR /app

# Instalar dependencias del sistema para spacy
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copiar requirements e instalar dependencias
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Descargar modelo de spacy para español
RUN python -m spacy download es_core_news_sm

# Copiar código
COPY . .

# Puerto
EXPOSE 8082

# Comando
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8082"]
