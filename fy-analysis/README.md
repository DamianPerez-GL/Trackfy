# Fy-Analysis - Servicio de Análisis de Amenazas

API REST en Go para detectar emails, URLs y números de teléfono maliciosos.

## Índice

1. [Inicio Rápido](#inicio-rápido)
2. [Endpoints](#endpoints)
3. [Cómo Detecta Amenazas](#cómo-detecta-amenazas)
4. [Ejemplos de Uso](#ejemplos-de-uso)
5. [Testing con Burp Suite](#testing-con-burp-suite)
6. [Estructura del Proyecto](#estructura-del-proyecto)

---

## Inicio Rápido

### Requisitos
- Docker Desktop

### Ejecutar

```bash
cd fy-analysis

# Construir e iniciar
docker compose up -d --build

# Verificar que está corriendo
docker ps

# Ver logs
docker compose logs -f

# Detener
docker compose down
```

### Verificar funcionamiento

```bash
curl http://localhost:9090/health
```

Respuesta esperada:
```json
{"status":"healthy","version":"1.0.0","timestamp":"2024-..."}
```

---

## Endpoints

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/analyze/email` | Analizar email |
| POST | `/api/v1/analyze/url` | Analizar URL |
| POST | `/api/v1/analyze/phone` | Analizar teléfono |
| POST | `/api/v1/analyze/batch` | Análisis en lote |

---

## Cómo Detecta Amenazas

### 1. Análisis de Emails

```
Email recibido
      │
      ▼
┌─────────────────┐
│ Validar formato │ ──► ¿Tiene formato válido de email?
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ Verificar blacklist │ ──► ¿Dominio conocido como spam/phishing?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Verificar temporal  │ ──► ¿Es email desechable (tempmail, etc)?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Verificar DNS MX    │ ──► ¿El dominio puede recibir correos?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Análisis heurístico │ ──► Patrones sospechosos + contexto
└────────┬────────────┘
         │
         ▼
    RESULTADO
```

#### Qué detecta:

| Criterio | Descripción | Ejemplo |
|----------|-------------|---------|
| **Blacklist** | Dominios de spam/phishing conocidos | `phishing-site.net` |
| **Email desechable** | Servicios de email temporal | `tempmail.com`, `guerrillamail.com` |
| **Sin MX** | Dominio sin servidor de correo | No puede recibir respuestas |
| **Patrones sospechosos** | Palabras en el nombre | `admin`, `security`, `verify` |
| **Contexto urgente** | Texto que acompaña al email | "URGENTE", "cuenta suspendida" |

#### Dominios desechables detectados:
```
tempmail.com, guerrillamail.com, 10minutemail.com, mailinator.com,
yopmail.com, trashmail.com, fakeinbox.com, temp-mail.org, maildrop.cc
```

---

### 2. Análisis de URLs

```
URL recibida
      │
      ▼
┌─────────────────┐
│  Parsear URL    │ ──► Extraer dominio, path, parámetros
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ Verificar blacklist │ ──► ¿Dominio de malware conocido?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Detectar acortador  │ ──► ¿Es bit.ly, tinyurl, etc?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Verificar HTTPS     │ ──► ¿Usa conexión segura?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Análisis heurístico │ ──► 10 verificaciones de patrones
└────────┬────────────┘
         │
         ▼
    RESULTADO
```

#### Qué detecta:

| Criterio | Descripción | Ejemplo |
|----------|-------------|---------|
| **Blacklist** | Dominios de malware | `malware-distribution.com` |
| **TLD sospechoso** | Extensiones usadas para phishing | `.tk`, `.ml`, `.xyz`, `.click` |
| **IP en URL** | Usa IP en lugar de dominio | `http://192.168.1.1/login` |
| **URL acortada** | Destino desconocido | `bit.ly/xxx`, `tinyurl.com/xxx` |
| **Homógrafos** | Caracteres que parecen otros | `pаypal.com` (а cirílico) |
| **Sin HTTPS** | Conexión no segura | `http://banco.com/login` |
| **Muchos guiones** | Typosquatting | `paypal-secure-login-verify.com` |
| **Params sospechosos** | Redirecciones | `?redirect=`, `?url=`, `?goto=` |
| **@ en URL** | Técnica de ofuscación | `http://google.com@evil.com` |
| **Palabras phishing** | Keywords de estafa | `login`, `verify`, `password`, `bank` |

---

### 3. Análisis de Teléfonos

```
Teléfono recibido
      │
      ▼
┌─────────────────┐
│ Limpiar número  │ ──► Quitar espacios, guiones
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ Validar formato     │ ──► ¿Longitud y formato correcto?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Detectar país       │ ──► +34=España, +52=México, etc.
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Verificar scam DB   │ ──► ¿Número reportado como estafa?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Detectar premium    │ ──► ¿Es número de tarificación especial?
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Análisis contexto   │ ──► ¿Mensaje de estafa?
└────────┬────────────┘
         │
         ▼
    RESULTADO
```

#### Qué detecta:

| Criterio | Descripción | Ejemplo |
|----------|-------------|---------|
| **Números premium** | Tarificación especial (cobran extra) | 806, 807, 900 en España |
| **Scam DB** | Números reportados como estafa | Base de datos interna |
| **Contexto scam** | Texto típico de estafas | "Has ganado", "lotería", "premio" |

#### Prefijos premium por país:

| País | Prefijos |
|------|----------|
| España | 803, 806, 807, 905 |
| México | 900 |
| USA | 900, 976 |

---

## Niveles de Amenaza

| Nivel | Color | Significado |
|-------|-------|-------------|
| `safe` | 🟢 Verde | Sin amenazas |
| `low` | 🟡 Amarillo | Sospecha menor |
| `medium` | 🟠 Naranja | Precaución |
| `high` | 🔴 Rojo | Probable amenaza |
| `critical` | ⛔ Rojo oscuro | Amenaza confirmada |

---

## Ejemplos de Uso

### Con curl (CMD)

```bash
# Health check
curl http://localhost:9090/health

# Analizar email seguro
curl -X POST http://localhost:9090/api/v1/analyze/email -H "Content-Type: application/json" -d "{\"email\":\"usuario@gmail.com\"}"

# Analizar email sospechoso
curl -X POST http://localhost:9090/api/v1/analyze/email -H "Content-Type: application/json" -d "{\"email\":\"security@banco-falso.tk\",\"context\":\"URGENTE: Verifique su cuenta\"}"

# Analizar URL de phishing
curl -X POST http://localhost:9090/api/v1/analyze/url -H "Content-Type: application/json" -d "{\"url\":\"http://paypal-login.tk/verify\"}"

# Analizar URL con IP
curl -X POST http://localhost:9090/api/v1/analyze/url -H "Content-Type: application/json" -d "{\"url\":\"http://192.168.1.100/admin\"}"

# Analizar teléfono premium
curl -X POST http://localhost:9090/api/v1/analyze/phone -H "Content-Type: application/json" -d "{\"phone\":\"+34806123456\",\"country_code\":\"ES\"}"

# Análisis en lote
curl -X POST http://localhost:9090/api/v1/analyze/batch -H "Content-Type: application/json" -d "{\"emails\":[\"good@gmail.com\",\"bad@tempmail.com\"],\"urls\":[\"https://google.com\",\"http://evil.tk\"],\"phones\":[\"+34612345678\"]}"
```

### Con PowerShell

```powershell
# Health check
Invoke-RestMethod -Uri "http://localhost:9090/health"

# Analizar email
$body = @{ email = "test@tempmail.com" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/analyze/email" -Method POST -Body $body -ContentType "application/json"

# Analizar URL
$body = @{ url = "http://phishing-site.tk/login" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/analyze/url" -Method POST -Body $body -ContentType "application/json"

# Analizar teléfono
$body = @{ phone = "+34806123456"; country_code = "ES" } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/analyze/phone" -Method POST -Body $body -ContentType "application/json"
```

### Ejemplos de respuestas

**Email desechable:**
```json
{
  "email": "test@tempmail.com",
  "analysis": {
    "is_malicious": false,
    "threat_level": "medium",
    "threat_types": ["disposable_email"],
    "confidence": 0.95,
    "reasons": ["Email de dominio desechable/temporal"]
  },
  "domain_info": {
    "domain": "tempmail.com",
    "is_disposable": true,
    "is_freemail": false,
    "has_mx_records": true
  },
  "recommendations": ["Solicitar un email corporativo o personal permanente"]
}
```

**URL de phishing:**
```json
{
  "url": "http://paypal-secure-login.tk/verify",
  "analysis": {
    "is_malicious": true,
    "threat_level": "high",
    "threat_types": ["phishing"],
    "confidence": 0.85,
    "reasons": [
      "TLD frecuentemente usado en sitios maliciosos",
      "Dominio con múltiples guiones (posible typosquatting)",
      "Contiene palabras clave asociadas a phishing: paypal",
      "No usa conexión segura (HTTPS)"
    ]
  },
  "url_info": {
    "domain": "paypal-secure-login.tk",
    "scheme": "http",
    "is_shortened": false,
    "ssl_valid": false
  },
  "recommendations": ["Verificar la autenticidad del sitio antes de ingresar datos"]
}
```

**Teléfono premium:**
```json
{
  "phone": "+34806123456",
  "analysis": {
    "is_malicious": false,
    "threat_level": "medium",
    "threat_types": ["fraud"],
    "confidence": 0.9,
    "reasons": ["Número de tarificación especial (premium)"]
  },
  "phone_info": {
    "country_code": "ES",
    "country": "España",
    "type": "premium",
    "is_valid": true,
    "is_premium_rate": true
  },
  "recommendations": ["Llamar a este número puede generar cargos elevados"]
}
```

---

## Testing con Burp Suite

Configuración:
- **API**: `http://127.0.0.1:9090`
- **Burp Proxy**: `http://127.0.0.1:8080`

### Comandos con proxy (CMD)

```bash
# Health check
curl -x http://127.0.0.1:8080 http://127.0.0.1:9090/health

# Analizar email
curl -x http://127.0.0.1:8080 -X POST http://127.0.0.1:9090/api/v1/analyze/email -H "Content-Type: application/json" -d "{\"email\":\"test@tempmail.com\"}"

# Analizar URL
curl -x http://127.0.0.1:8080 -X POST http://127.0.0.1:9090/api/v1/analyze/url -H "Content-Type: application/json" -d "{\"url\":\"http://evil.tk/login\"}"

# Analizar teléfono
curl -x http://127.0.0.1:8080 -X POST http://127.0.0.1:9090/api/v1/analyze/phone -H "Content-Type: application/json" -d "{\"phone\":\"+34806123456\",\"country_code\":\"ES\"}"

# Batch
curl -x http://127.0.0.1:8080 -X POST http://127.0.0.1:9090/api/v1/analyze/batch -H "Content-Type: application/json" -d "{\"emails\":[\"a@gmail.com\",\"b@tempmail.com\"],\"urls\":[\"https://google.com\"]}"
```

> **Nota**: Usa `127.0.0.1` en lugar de `localhost` para que el tráfico pase por Burp.

---

## Estructura del Proyecto

```
fy-analysis/
├── cmd/server/main.go           # Punto de entrada
├── internal/
│   ├── api/
│   │   ├── handlers/            # Handlers HTTP (email, url, phone, batch)
│   │   ├── middleware/          # Logging
│   │   └── router.go            # Definición de rutas
│   ├── analyzer/
│   │   ├── email/validator.go   # Lógica detección emails
│   │   ├── url/analyzer.go      # Lógica detección URLs
│   │   └── phone/analyzer.go    # Lógica detección teléfonos
│   ├── models/                  # Request/Response DTOs
│   └── config/                  # Configuración
├── tests/                       # Tests unitarios
├── scripts/                     # Scripts PowerShell
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## Configuración

Variables de entorno (en `docker-compose.yml`):

| Variable | Default | Descripción |
|----------|---------|-------------|
| `PORT` | 9090 | Puerto de la API |
| `ENVIRONMENT` | development | Entorno |
| `LOG_LEVEL` | info | Nivel de logs |
| `RATE_LIMIT` | 100 | Peticiones por minuto por IP |