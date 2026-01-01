# Sistema de Reportes de Usuarios - Trackfy

## Resumen

El sistema de reportes permite a los usuarios de Trackfy reportar URLs sospechosas que se integran en el análisis de amenazas de Fy. Incluye un **sistema anti-spam** robusto que previene que grupos pequeños de usuarios inflen artificialmente la peligrosidad de URLs legítimas.

## Arquitectura

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   Frontend   │ ──▶  │ API Gateway  │ ──▶  │ Fy-Analysis  │
│   (Flutter)  │      │   :8080      │      │    :9090     │
└──────────────┘      └──────────────┘      └──────────────┘
                                                    │
                                                    ▼
                                            ┌──────────────┐
                                            │  PostgreSQL  │
                                            │  (fy_threats)│
                                            └──────────────┘
```

## Base de Datos

### Tablas Principales

#### `user_trust_scores`
Sistema de confianza por usuario.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| user_id | VARCHAR(64) | ID único del usuario |
| trust_score | SMALLINT | Puntuación 0-100 (default: 50) |
| total_reports | INTEGER | Total de reportes realizados |
| confirmed_reports | INTEGER | Reportes confirmados como válidos |
| rejected_reports | INTEGER | Reportes rechazados |
| flags | SMALLINT | bit 0: is_banned, bit 1: is_trusted_reviewer |

**Niveles de Confianza:**
- `0-19`: Usuario no confiable / baneado
- `20-49`: Confianza baja (nuevos usuarios)
- `50-79`: Confianza normal
- `80-100`: Usuario muy confiable

#### `reported_urls`
URLs agregadas con puntuación anti-spam.

| Campo | Tipo | Descripción |
|-------|------|-------------|
| url_hash | BYTEA | Hash SHA256 de la URL (PK) |
| url | VARCHAR(2048) | URL normalizada |
| domain | VARCHAR(253) | Dominio extraído |
| primary_threat_type | ENUM | Tipo de amenaza más reportado |
| aggregated_score | SMALLINT | Score 0-100 calculado anti-spam |
| total_reports | INTEGER | Número total de reportes |
| unique_reporters | INTEGER | Usuarios únicos que reportaron |
| weighted_trust_sum | NUMERIC | Suma ponderada de trust de reportadores |
| status | ENUM | pending, reviewed, confirmed, rejected |

#### `user_url_reports`
Reportes individuales (con deduplicación).

| Campo | Tipo | Descripción |
|-------|------|-------------|
| url_hash | BYTEA | Referencia a reported_urls |
| user_id | VARCHAR(64) | Usuario que reporta |
| user_trust_at_report | SMALLINT | Trust del usuario al momento del reporte |
| threat_type | ENUM | Tipo de amenaza reportada |
| description | VARCHAR(500) | Descripción opcional |
| UNIQUE | (url_hash, user_id) | Evita reportes duplicados |

## Sistema Anti-Spam

### Problema
Sin protección, 4 usuarios coordinados podrían reportar una URL legítima y hacer que Fy la marque como peligrosa.

### Solución
El sistema implementa un **algoritmo de agregación ponderada**:

#### 1. Confianza Base por Usuario
- Usuarios nuevos empiezan con `trust_score = 50`
- Cada reporte confirmado: `+3 puntos`
- Cada reporte rechazado: `-10 puntos`
- Si `trust_score < 10`: usuario es baneado automáticamente

#### 2. Multiplicador por Cantidad de Reportadores

| Reportadores Únicos | Multiplicador |
|---------------------|---------------|
| 1 | x0.3 |
| 2 | x0.5 |
| 3-4 | x0.6 |
| 5-9 | x0.7 |
| 10-19 | x0.85 |
| 20+ | x1.0 |

#### 3. Cap Máximo de Score

| Reportadores Únicos | Score Máximo |
|---------------------|--------------|
| 1 | 30 |
| 2 | 45 |
| 3-4 | 60 |
| 5-9 | 75 |
| 10+ | 100 |

### Fórmula de Cálculo

```sql
base_score = AVG(user_trust) de todos los reportadores
multiplied_score = base_score * multiplicador_por_cantidad
final_score = MIN(multiplied_score, cap_por_cantidad)
```

### Ejemplo Práctico

**Escenario: 4 usuarios intentan inflar una URL**

| Usuario | Trust Score |
|---------|-------------|
| A | 50 |
| B | 50 |
| C | 30 |
| D | 50 |

```
Promedio trust = (50 + 50 + 30 + 50) / 4 = 45
Multiplicador (3-4 reportadores) = 0.6
Score calculado = 45 * 0.6 = 27
Cap máximo (3-4 reportadores) = 60
Final score = MIN(27, 60) = 27
```

**Resultado:** La URL obtiene score 27, que es nivel WARNING bajo, NO danger.
Se necesitarían muchos más reportadores confiables para alcanzar niveles altos.

## API Endpoints

### Reportar URL
```http
POST /api/v1/report
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "https://example-phishing.com/login",
  "threat_type": "phishing",
  "description": "Recibido por SMS suplantando a mi banco"
}
```

**Respuesta:**
```json
{
  "success": true,
  "message": "Reporte registrado correctamente",
  "url_score": 27
}
```

**Tipos de amenaza válidos:**
- `phishing` - Suplantación de identidad
- `malware` - Software malicioso
- `scam` - Estafa
- `spam` - Contenido no deseado
- `vishing` - Phishing por voz
- `smishing` - Phishing por SMS
- `other` - Otros

### Estadísticas de Reportes (Admin)
```http
GET /api/v1/reports/stats
```

**Respuesta:**
```json
{
  "reported_urls": {
    "total": 150,
    "pending_review": 45,
    "confirmed": 80,
    "rejected": 25,
    "high_score": 20,
    "medium_score": 35,
    "low_score": 95,
    "avg_reporters": 3.2,
    "total_reports": 480
  },
  "reporters": {
    "total_users": 1200,
    "banned_users": 15,
    "avg_trust_score": 52.3
  }
}
```

## Integración en Fy

### Peso en el Análisis
El checker `user_reports` tiene peso **0.10** (10%) en el análisis total:

| Fuente | Peso |
|--------|------|
| LocalDB (amenazas verificadas) | 0.30 |
| URLhaus | 0.15 |
| PhishTank | 0.15 |
| Google Web Risk | 0.15 |
| URLScan.io | 0.10 |
| **User Reports** | **0.10** |
| Heurísticas | 0.15 |

### Umbrales de Activación
El checker solo considera URLs reportadas que cumplan:
- `unique_reporters >= 2` (mínimo 2 usuarios diferentes)
- `aggregated_score >= 40` para WARNING
- `aggregated_score >= 70` para DANGER

### Boost por Confirmación
Si un revisor confirma el reporte:
- Confianza del resultado aumenta x1.3
- Se muestra "URL confirmada como maliciosa por revisores"

## Flujo de Revisión

```
Usuario reporta → pending
        ↓
  Revisor analiza
        ↓
    ┌───┴───┐
confirmed  rejected
    │         │
    ↓         ↓
 +3 trust   -10 trust
 al usuario  al usuario
```

## Configuración

### Variables de Entorno (fy-analysis)

```bash
# Habilitar sistema de reportes
ENABLE_USER_REPORTS=true

# Umbrales (opcionales)
USER_REPORTS_MIN_WARNING=40
USER_REPORTS_MIN_DANGER=70
USER_REPORTS_MIN_REPORTERS=2
```

## Casos de Uso

### Caso 1: Reporte Legítimo
1. Usuario A (trust: 75) recibe SMS phishing con URL
2. Reporta la URL como "phishing"
3. Score inicial: 75 * 0.3 = 22.5 (cap 30) → 22
4. Usuario B (trust: 80) también reporta
5. Score: ((75+80)/2) * 0.5 = 38.75 (cap 45) → 38
6. Más usuarios reportan...
7. Con 10+ reportadores confiables, score puede llegar a ~80

### Caso 2: Intento de Abuso
1. 4 cuentas nuevas (trust: 50 cada una) reportan URL legítima
2. Score: 50 * 0.6 = 30 (cap 60) → 30
3. Fy mostrará WARNING leve, NO DANGER
4. Si se revisa y rechaza:
   - Los 4 usuarios bajan a trust 40
   - Si reinciden, serán baneados

### Caso 3: Usuario Baneado
1. Usuario C tiene trust: 8 (por múltiples reportes falsos)
2. Intenta reportar una URL
3. Sistema rechaza: "Usuario no puede reportar"
4. Sus reportes anteriores no afectan scores

## Mantenimiento

### Limpieza de Particiones
Las particiones de `user_url_reports` se crean mensualmente. Eliminar particiones antiguas:

```sql
DROP TABLE IF EXISTS user_reports_2024_01;
```

### Monitoreo de Abuso
```sql
-- Usuarios con alto ratio de rechazos
SELECT user_id, trust_score, total_reports, rejected_reports,
       (rejected_reports::float / NULLIF(total_reports, 0)) as reject_ratio
FROM user_trust_scores
WHERE total_reports > 5
ORDER BY reject_ratio DESC
LIMIT 20;
```

### Promoción a Threat DB
URLs con alto score confirmado pueden promoverse a `threat_domains`:

```sql
-- Ver URLs candidatas a promoción
SELECT url, domain, aggregated_score, unique_reporters
FROM reported_urls
WHERE status = 'confirmed'
  AND aggregated_score >= 80
  AND unique_reporters >= 10
  AND promoted_to_threats = FALSE;
```

## Seguridad

- **Rate limiting:** 10 reportes/minuto por usuario
- **Autenticación:** Solo usuarios verificados pueden reportar
- **Deduplicación:** Un usuario no puede reportar la misma URL dos veces
- **IP tracking:** Se registra IP para detectar patrones de abuso
- **Trust decay:** Usuarios inactivos no pierden trust, pero reportes rechazados sí penalizan

## FAQ

**P: ¿Puedo reportar la misma URL varias veces?**
R: No. El sistema detecta duplicados y devuelve "Ya reportaste esta URL anteriormente".

**P: ¿Cómo subo mi nivel de confianza?**
R: Reportando URLs que luego sean confirmadas como maliciosas por revisores.

**P: ¿Por qué mi reporte tiene score bajo?**
R: El score depende de cuántos usuarios únicos hayan reportado y su nivel de confianza.

**P: ¿Pueden banearme por reportar?**
R: Solo si múltiples reportes tuyos son rechazados como falsos.
