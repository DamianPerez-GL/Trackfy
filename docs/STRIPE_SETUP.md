# Configuración de Stripe - Trackfy Premium

## Resumen del Sistema

Trackfy Premium es un sistema de suscripción mensual que permite a los usuarios acceder a mensajes ilimitados con Fy.

| Plan | Precio | Mensajes/mes |
|------|--------|--------------|
| Free | Gratis | 5 |
| Premium | 4,99€/mes | Ilimitados |

---

## Arquitectura

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│   Flutter App   │────▶│   API Gateway    │────▶│   Stripe    │
│   (Android/iOS) │◀────│   (localhost:8080)│◀────│    API      │
└─────────────────┘     └──────────────────┘     └─────────────┘
                               │
                        ┌──────┴──────┐
                        ▼             ▼
                  ┌──────────┐  ┌─────────┐
                  │PostgreSQL│  │  Redis  │
                  │ (5433)   │  │ (6379)  │
                  └──────────┘  └─────────┘
```

---

## Configuración Inicial

### 1. Crear Cuenta de Stripe

1. Ir a [stripe.com](https://stripe.com) y crear cuenta gratuita
2. Activar **modo test** (toggle en la esquina superior derecha)

### 2. Obtener API Keys

1. Ir a **Developers > API Keys**
2. Copiar **Secret key** (empieza con `sk_test_`)

### 3. Crear Producto y Precio

1. Ir a **Products > Add Product**
2. Configurar:
   - **Nombre**: Trackfy Premium
   - **Descripción**: Mensajes ilimitados con Fy
   - **Precio**: 4,99 EUR
   - **Tipo**: Recurrente (mensual)
3. Guardar y copiar el **Price ID** (empieza con `price_`)

### 4. Configurar Variables de Entorno

Editar el archivo `.env` en la raíz del proyecto:

```env
# Stripe - Modo Test
STRIPE_SECRET_KEY=sk_test_xxxxxxxxxxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
STRIPE_PRICE_ID=price_xxxxxxxxxxxxx
```

---

## Configuración de Webhooks

### Desarrollo Local (con Stripe CLI)

#### Paso 1: Instalar Stripe CLI

**Windows (con Scoop):**
```powershell
# Instalar Scoop si no lo tienes
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.dev | iex

# Instalar Stripe CLI
scoop install stripe
```

**Windows (descarga directa):**
1. Descargar de: https://github.com/stripe/stripe-cli/releases/latest
2. Extraer `stripe_x.x.x_windows_x86_64.zip`
3. Añadir la carpeta al PATH del sistema

**macOS:**
```bash
brew install stripe/stripe-cli/stripe
```

**Linux:**
```bash
# Debian/Ubuntu
curl -s https://packages.stripe.dev/api/security/keypair/stripe-cli-gpg/public | gpg --dearmor | sudo tee /usr/share/keyrings/stripe.gpg
echo "deb [signed-by=/usr/share/keyrings/stripe.gpg] https://packages.stripe.dev/stripe-cli-debian-local stable main" | sudo tee -a /etc/apt/sources.list.d/stripe.list
sudo apt update && sudo apt install stripe
```

#### Paso 2: Autenticarse en Stripe

```bash
stripe login
```
Esto abrirá el navegador para autenticarte.

#### Paso 3: Iniciar el Listener de Webhooks

```bash
stripe listen --forward-to localhost:8080/webhook/stripe
```

Esto mostrará:
```
> Ready! Your webhook signing secret is whsec_abc123...
```

**Importante**: Copia el `whsec_...` y actualiza tu `.env`:
```env
STRIPE_WEBHOOK_SECRET=whsec_abc123...
```

#### Paso 4: Mantener el Listener Activo

Deja esta terminal abierta mientras desarrollas. El listener reenviará todos los eventos de Stripe a tu localhost.

### Producción (Webhook Real)

1. Ir a **Developers > Webhooks** en Stripe Dashboard
2. Click en **Add endpoint**
3. Configurar:
   - **URL**: `https://tu-dominio.com/webhook/stripe`
   - **Eventos a escuchar**:
     - `checkout.session.completed`
     - `customer.subscription.created`
     - `customer.subscription.updated`
     - `customer.subscription.deleted`
     - `invoice.paid`
     - `invoice.payment_failed`
4. Copiar **Signing secret** al `.env` de producción

---

## Base de Datos

### Migración SQL

La migración `002_subscriptions.sql` crea:

**Tablas:**
- `subscriptions` - Estado de suscripción por usuario
- `payment_history` - Historial de pagos

**Funciones:**
- `increment_message_count(user_id)` - Incrementa contador y verifica límite
- `get_subscription_status(user_id)` - Obtiene estado actual
- `upgrade_to_premium(user_id, stripe_customer_id, stripe_subscription_id)` - Actualiza a premium
- `cancel_premium(user_id)` - Cancela y vuelve a free

### Ejecutar Migración

```bash
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway < api-gateway/scripts/002_subscriptions.sql
```

### Verificar Estado

```bash
# Ver todas las suscripciones
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway -c "SELECT * FROM subscriptions;"

# Ver estado de un usuario específico
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway -c "SELECT * FROM get_subscription_status('USER_ID_AQUI'::uuid);"
```

---

## Endpoints de la API

| Endpoint | Método | Auth | Descripción |
|----------|--------|------|-------------|
| `/api/v1/subscription/status` | GET | JWT | Estado actual de suscripción |
| `/api/v1/subscription/checkout` | POST | JWT | Crear sesión de pago Stripe |
| `/api/v1/subscription/portal` | POST | JWT | URL del portal de gestión |
| `/webhook/stripe` | POST | Firma Stripe | Webhook para eventos de Stripe |

### Ejemplo: Obtener Estado

```bash
curl -X GET http://localhost:8080/api/v1/subscription/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Respuesta:
```json
{
  "plan": "free",
  "status": "active",
  "messages_used": 2,
  "messages_limit": 5,
  "messages_remaining": 3,
  "is_premium": false,
  "period_end": "2026-02-01T00:00:00Z"
}
```

### Ejemplo: Crear Checkout

```bash
curl -X POST http://localhost:8080/api/v1/subscription/checkout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Respuesta:
```json
{
  "checkout_url": "https://checkout.stripe.com/c/pay/cs_test_..."
}
```

---

## Flujo de Pago

```
1. Usuario pulsa "Suscribirme" en la app
2. App → POST /api/v1/subscription/checkout
3. Backend crea Stripe Customer (si no existe)
4. Backend crea Checkout Session
5. App abre checkout_url en navegador
6. Usuario completa pago con tarjeta
7. Stripe envía webhook checkout.session.completed
8. Backend actualiza BD → Usuario es premium
9. Usuario vuelve a la app
10. App refresca estado → Mensajes ilimitados
```

---

## Testing

### Tarjetas de Prueba

| Número | Resultado |
|--------|-----------|
| 4242 4242 4242 4242 | Pago exitoso |
| 4000 0000 0000 0002 | Tarjeta rechazada |
| 4000 0000 0000 3220 | Requiere 3D Secure |

- **Fecha**: Cualquier fecha futura (ej: 12/34)
- **CVC**: Cualquier 3 dígitos (ej: 123)

### Probar Manualmente

```bash
# 1. Hacer premium a un usuario manualmente
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway -c "
  UPDATE subscriptions
  SET plan = 'premium', messages_limit = -1, status = 'active'
  WHERE user_id = 'USER_ID'::uuid;
"

# 2. Volver a free
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway -c "
  UPDATE subscriptions
  SET plan = 'free', messages_limit = 5, messages_used = 0
  WHERE user_id = 'USER_ID'::uuid;
"

# 3. Resetear contador de mensajes
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway -c "
  UPDATE subscriptions SET messages_used = 0 WHERE user_id = 'USER_ID'::uuid;
"
```

### Simular Webhook con Stripe CLI

```bash
# Simular pago completado
stripe trigger checkout.session.completed

# Simular cancelación
stripe trigger customer.subscription.deleted
```

---

## Troubleshooting

### El webhook no llega

1. Verificar que Stripe CLI está corriendo: `stripe listen --forward-to localhost:8080/webhook/stripe`
2. Verificar logs del api-gateway: `docker logs api-gateway -f`
3. Verificar que el `STRIPE_WEBHOOK_SECRET` coincide

### Error 429 al enviar mensajes

El usuario alcanzó el límite. Verificar en BD:
```bash
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway -c "
  SELECT messages_used, messages_limit FROM subscriptions WHERE user_id = 'USER_ID'::uuid;
"
```

### El pago fue exitoso pero no se actualizó

1. Verificar logs del webhook
2. Actualizar manualmente:
```bash
docker exec -i fy-postgres-gateway psql -U trackfy -d trackfy_gateway -c "
  SELECT upgrade_to_premium('USER_ID'::uuid, 'cus_xxx', 'sub_xxx');
"
```

---

## Variables de Entorno

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `STRIPE_SECRET_KEY` | API Key secreta | `sk_test_51Sng...` |
| `STRIPE_WEBHOOK_SECRET` | Firma del webhook | `whsec_abc123...` |
| `STRIPE_PRICE_ID` | ID del precio mensual | `price_1Snin9...` |
| `STRIPE_SUCCESS_URL` | URL tras pago exitoso | `http://localhost:3000/success` |
| `STRIPE_CANCEL_URL` | URL si cancela | `http://localhost:3000/cancel` |

---

## Checklist de Despliegue

- [ ] Crear cuenta Stripe y activar modo live
- [ ] Crear producto "Trackfy Premium" en producción
- [ ] Configurar webhook con URL de producción
- [ ] Actualizar `.env.production` con claves live (`sk_live_...`)
- [ ] Ejecutar migración SQL en BD de producción
- [ ] Probar flujo completo con tarjeta real
