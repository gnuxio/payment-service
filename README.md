# Payment Service

Microservicio de pagos con Stripe, construido en Go puro (net/http) para gestionar suscripciones de múltiples SAAS.

## Stack Tecnológico

- **Go 1.23+** (net/http puro, sin frameworks)
- **Stripe** para procesamiento de pagos
- **PostgreSQL** para persistencia
- **Docker** para containerización

## Características

- ✅ Creación de sesiones de checkout con Stripe
- ✅ Gestión de suscripciones (crear, consultar, cancelar)
- ✅ Webhooks de Stripe para sincronización automática
- ✅ Multi-tenant (soporta múltiples SAAS)
- ✅ Notificaciones a backend vía webhook
- ✅ Historial de facturas
- ✅ Autenticación con API Key

## Arquitectura

```
Frontend → menuum-backend → payment-service → Stripe
                ↑                    ↓
                └──── webhook ───────┘
```

El frontend NO llama directamente a payment-service. Solo menuum-backend se comunica con él.

## Endpoints

### Públicos (sin autenticación)

- `GET /payments/health` - Health check
- `POST /payments/webhook` - Recibir eventos de Stripe

### Protegidos (requieren API Key en header `X-API-Key`)

- `POST /payments/checkout` - Crear sesión de pago
- `GET /payments/subscription/:userId` - Ver estado de suscripción
- `POST /payments/cancel/:userId` - Cancelar suscripción

## Documentación y Ejemplos

Este proyecto incluye ejemplos de integración y scripts de prueba:

- **`examples/menuum-backend-integration.py`** - Código de ejemplo completo para integrar con menuum-backend (Python/FastAPI)
- **`examples/curl-examples.sh`** - Script con ejemplos de curl para probar todos los endpoints
- **`CHANGELOG.md`** - Registro de cambios del proyecto

## Configuración de Stripe

**⚠️ IMPORTANTE**: Antes de usar el servicio, DEBES actualizar los Price IDs en `internal/stripe/client.go` con los IDs reales de tus productos en Stripe (ver paso 2 más abajo).

### 1. Obtener API Keys

1. Ve a [Stripe Dashboard](https://dashboard.stripe.com/)
2. Ve a **Developers → API Keys**
3. Copia tu **Secret Key** (empieza con `sk_test_` para modo test)
4. Guárdala en el archivo `.env` como `STRIPE_SECRET_KEY`

### 2. Crear Productos y Precios

1. Ve a **Products → Add Product**
2. Crea dos productos:
   - **Premium Monthly**: $9.99/mes
   - **Premium Yearly**: $99/año
3. Después de crear cada producto, copia el **Price ID** (empieza con `price_`)
4. Actualiza el archivo `internal/stripe/client.go`:

```go
switch plan {
case models.PlanPremiumMonthly:
    return "price_TU_PRICE_ID_MENSUAL_AQUI", nil
case models.PlanPremiumYearly:
    return "price_TU_PRICE_ID_ANUAL_AQUI", nil
}
```

### 3. Configurar Webhook

1. Ve a **Developers → Webhooks**
2. Clic en **Add endpoint**
3. Endpoint URL: `https://TU_DOMINIO/payments/webhook`
   - Para desarrollo local, usa [ngrok](https://ngrok.com/) o similar
4. Selecciona estos eventos:
   - `checkout.session.completed`
   - `customer.subscription.created`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.paid`
   - `invoice.payment_failed`
5. Copia el **Signing Secret** (empieza con `whsec_`)
6. Guárdalo en `.env` como `STRIPE_WEBHOOK_SECRET`

## Instalación

### Requisitos Previos

- Go 1.23+
- Docker y Docker Compose
- Cuenta de Stripe

### Paso 1: Clonar e Instalar Dependencias

```bash
cd payment-service
go mod download
```

### Paso 2: Configurar Variables de Entorno

```bash
cp .env.example .env
```

Edita `.env` con tus valores reales:

```bash
# Genera un API Key seguro
openssl rand -hex 32

# Configura tu .env
PORT=8081
DATABASE_URL=postgres://postgres:postgres@localhost:5433/payment_service?sslmode=disable
STRIPE_SECRET_KEY=sk_test_tu_key_aqui
STRIPE_WEBHOOK_SECRET=whsec_tu_secret_aqui
API_KEY=tu_api_key_generado_aqui
BACKEND_WEBHOOK_URL=http://localhost:8000
```

### Paso 3: Iniciar con Docker

```bash
docker-compose up -d
```

Esto iniciará:
- PostgreSQL en puerto `5433`
- Payment Service en puerto `8081`

### Paso 4: Verificar

```bash
curl http://localhost:8081/payments/health
```

Deberías recibir: `{"status":"healthy"}`

## Desarrollo Local (sin Docker)

### 1. Iniciar PostgreSQL

```bash
docker run -d \
  --name payment-postgres \
  -e POSTGRES_DB=payment_service \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 \
  postgres:16-alpine
```

### 2. Ejecutar el Servicio

```bash
# Cargar variables de entorno
export $(cat .env | xargs)

# Ejecutar
go run cmd/server/main.go
```

## Base de Datos

### Esquema

#### Tabla: `subscriptions`

```sql
- id (serial)
- user_id (varchar)
- tenant (varchar)
- stripe_customer_id (varchar)
- stripe_subscription_id (varchar)
- status (varchar)
- plan (varchar)
- current_period_start (timestamp)
- current_period_end (timestamp)
- cancel_at_period_end (boolean)
- created_at (timestamp)
- updated_at (timestamp)
```

#### Tabla: `invoices`

```sql
- id (serial)
- subscription_id (integer)
- stripe_invoice_id (varchar)
- user_id (varchar)
- tenant (varchar)
- amount_paid (integer)
- currency (varchar)
- status (varchar)
- invoice_pdf (varchar)
- hosted_invoice_url (varchar)
- period_start (timestamp)
- period_end (timestamp)
- created_at (timestamp)
```

Las migraciones se ejecutan automáticamente al iniciar el servicio.

## Uso desde menuum-backend

### 1. Compartir API Key

Copia el `API_KEY` del `.env` de payment-service y guárdalo en menuum-backend.

### 2. Crear Sesión de Checkout

```python
import requests

headers = {
    "X-API-Key": "tu_api_key_aqui",
    "X-Tenant-ID": "menuum"
}

payload = {
    "user_id": "cognito_user_id_aqui",
    "plan": "premium_monthly",  # o "premium_yearly"
    "success_url": "https://menuum.com/success",
    "cancel_url": "https://menuum.com/cancel"
}

response = requests.post(
    "http://localhost:8081/payments/checkout",
    json=payload,
    headers=headers
)

# Redirigir usuario a session_url
session_url = response.json()["session_url"]
```

### 3. Consultar Suscripción

```python
user_id = "cognito_user_id_aqui"

response = requests.get(
    f"http://localhost:8081/payments/subscription/{user_id}",
    headers=headers
)

subscription = response.json()
print(subscription["status"])  # "active", "canceled", etc.
```

### 4. Cancelar Suscripción

```python
response = requests.post(
    f"http://localhost:8081/payments/cancel/{user_id}",
    headers=headers
)
```

### 5. Recibir Webhooks

Crea el endpoint `POST /webhooks/subscription` en menuum-backend:

```python
from fastapi import APIRouter, Request

router = APIRouter()

@router.post("/webhooks/subscription")
async def subscription_webhook(request: Request):
    payload = await request.json()

    user_id = payload["user_id"]
    status = payload["status"]
    plan = payload["plan"]
    subscription_id = payload["subscription_id"]

    # Actualizar is_premium en tu base de datos
    if status == "active":
        update_user_premium(user_id, is_premium=True)
    elif status == "canceled":
        update_user_premium(user_id, is_premium=False)

    return {"status": "success"}
```

## Planes Disponibles

- `premium_monthly`: $9.99/mes
- `premium_yearly`: $99/año

## Multi-Tenancy

Cada petición debe incluir el header `X-Tenant-ID` para identificar el SAAS:

```bash
X-Tenant-ID: menuum
X-Tenant-ID: otro-saas
```

Esto permite usar el mismo payment-service para múltiples aplicaciones.

## Testing con Stripe

### Tarjetas de Prueba

- **Éxito**: `4242 4242 4242 4242`
- **Requiere autenticación**: `4000 0025 0000 3155`
- **Declinada**: `4000 0000 0000 9995`

Usa cualquier fecha futura y CVC válido (ej: 123).

### Webhook Local con Stripe CLI

```bash
# Instalar Stripe CLI
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Escuchar webhooks
stripe listen --forward-to localhost:8081/payments/webhook

# Copiar el webhook secret que se muestra
# Actualizarlo en .env como STRIPE_WEBHOOK_SECRET
```

## Logs

El servicio registra eventos importantes:

```bash
# Ver logs con Docker
docker logs -f payment-service

# Logs incluyen:
# - Conexiones a base de datos
# - Webhooks recibidos de Stripe
# - Notificaciones enviadas a menuum-backend
# - Errores y warnings
```

## Estructura del Proyecto

```
payment-service/
├── cmd/
│   └── server/
│       └── main.go                 # Punto de entrada
├── internal/
│   ├── config/
│   │   └── config.go               # Configuración
│   ├── database/
│   │   ├── postgres.go             # Conexión DB
│   │   └── migrations/             # Migraciones SQL
│   ├── models/
│   │   ├── subscription.go         # Modelo Subscription
│   │   └── invoice.go              # Modelo Invoice
│   ├── repository/
│   │   ├── subscription_repository.go
│   │   └── invoice_repository.go
│   ├── handlers/
│   │   ├── health.go               # GET /payments/health
│   │   ├── checkout.go             # POST /payments/checkout
│   │   ├── webhook.go              # POST /payments/webhook
│   │   ├── subscription.go         # GET /payments/subscription/:userId
│   │   ├── cancel.go               # POST /payments/cancel/:userId
│   │   └── utils.go                # Utilidades
│   ├── middleware/
│   │   └── auth.go                 # Autenticación API Key
│   ├── stripe/
│   │   └── client.go               # Cliente Stripe
│   └── webhook/
│       └── client.go               # Cliente webhook a backend
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── .gitignore
└── README.md
```

## Seguridad

- ✅ Autenticación con API Key
- ✅ Validación de webhooks de Stripe
- ✅ Uso de HTTPS en producción (recomendado)
- ✅ Variables de entorno para secrets
- ✅ PostgreSQL con credenciales seguras

## Troubleshooting

### Error: "Missing API key"

Asegúrate de incluir el header `X-API-Key` en tus peticiones.

### Error: "X-Tenant-ID header is required"

Todas las peticiones protegidas requieren el header `X-Tenant-ID`.

### Webhook de Stripe no funciona

1. Verifica que el `STRIPE_WEBHOOK_SECRET` sea correcto
2. Para local, usa Stripe CLI o ngrok
3. Revisa los logs: `docker logs -f payment-service`

### Base de datos no conecta

1. Verifica que PostgreSQL esté corriendo: `docker ps`
2. Verifica el `DATABASE_URL` en `.env`
3. Prueba la conexión: `psql $DATABASE_URL`

## Próximos Pasos

- [ ] Agregar tests unitarios
- [ ] Implementar rate limiting
- [ ] Agregar métricas (Prometheus)
- [ ] Implementar retry logic para webhooks fallidos
- [ ] Agregar soporte para cupones de descuento
- [ ] Implementar cambio de plan

## Contribuir

Este es un proyecto privado de Naventro. Para cambios, contacta al equipo.

## Licencia

Privado - Naventro © 2026
