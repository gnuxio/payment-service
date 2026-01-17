# Setup Checklist

Usa este checklist para configurar payment-service por primera vez.

## Pre-requisitos

- [ ] Go 1.23+ instalado
- [ ] Docker y Docker Compose instalados
- [ ] Cuenta de Stripe creada

## Configuración de Stripe

### 1. API Keys
- [ ] Ir a [Stripe Dashboard](https://dashboard.stripe.com/)
- [ ] Ir a **Developers → API Keys**
- [ ] Copiar **Secret Key** (sk_test_...)
- [ ] Guardar en `.env` como `STRIPE_SECRET_KEY`

### 2. Crear Productos
- [ ] Ir a **Products → Add Product**
- [ ] Crear producto "Premium Monthly":
  - Nombre: Premium Monthly
  - Precio: $9.99
  - Facturación: Mensual recurrente
- [ ] Copiar el **Price ID** (price_...)
- [ ] Crear producto "Premium Yearly":
  - Nombre: Premium Yearly
  - Precio: $99
  - Facturación: Anual recurrente
- [ ] Copiar el **Price ID** (price_...)

### 3. Actualizar Price IDs en el Código
- [ ] Abrir `internal/stripe/client.go`
- [ ] Buscar la función `GetPriceID`
- [ ] Reemplazar `"price_premium_monthly"` con tu Price ID mensual real
- [ ] Reemplazar `"price_premium_yearly"` con tu Price ID anual real

**Ejemplo:**
```go
switch plan {
case models.PlanPremiumMonthly:
    return "price_1Abc123Xyz456", nil  // ← Tu Price ID real
case models.PlanPremiumYearly:
    return "price_1Def789Uvw012", nil  // ← Tu Price ID real
}
```

### 4. Configurar Webhook
- [ ] Ir a **Developers → Webhooks**
- [ ] Click **Add endpoint**
- [ ] URL: `https://TU_DOMINIO/payments/webhook`
  - Para desarrollo local: usar ngrok o Stripe CLI
- [ ] Seleccionar eventos:
  - [ ] `checkout.session.completed`
  - [ ] `customer.subscription.created`
  - [ ] `customer.subscription.updated`
  - [ ] `customer.subscription.deleted`
  - [ ] `invoice.paid`
  - [ ] `invoice.payment_failed`
- [ ] Copiar **Signing Secret** (whsec_...)
- [ ] Guardar en `.env` como `STRIPE_WEBHOOK_SECRET`

## Configuración del Proyecto

### 1. Clonar e Instalar
- [ ] Clonar repositorio
- [ ] Ejecutar `go mod download`

### 2. Variables de Entorno
- [ ] Copiar `.env.example` a `.env`
- [ ] Generar API Key: `openssl rand -hex 32`
- [ ] Completar todas las variables en `.env`:
  - [ ] `PORT=8081`
  - [ ] `DATABASE_URL=postgres://...`
  - [ ] `STRIPE_SECRET_KEY=sk_test_...`
  - [ ] `STRIPE_WEBHOOK_SECRET=whsec_...`
  - [ ] `API_KEY=...` (el generado arriba)
  - [ ] `BACKEND_WEBHOOK_URL=http://localhost:8000`

### 3. Base de Datos
- [ ] Iniciar PostgreSQL (viene con docker-compose)
- [ ] Las migraciones se ejecutan automáticamente al iniciar

## Iniciar el Servicio

### Opción 1: Docker (Recomendado)
- [ ] Ejecutar `docker-compose up -d`
- [ ] Verificar: `curl http://localhost:8081/payments/health`

### Opción 2: Local (Sin Docker)
- [ ] Iniciar PostgreSQL: `docker run -d --name payment-postgres -e POSTGRES_DB=payment_service -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 5433:5432 postgres:16-alpine`
- [ ] Cargar variables: `export $(cat .env | xargs)`
- [ ] Ejecutar: `go run cmd/server/main.go`
- [ ] Verificar: `curl http://localhost:8081/payments/health`

## Configurar menuum-backend

### 1. Agregar Variables de Entorno
- [ ] Agregar a `.env` de menuum-backend:
  - `PAYMENT_SERVICE_URL=http://localhost:8081`
  - `PAYMENT_SERVICE_API_KEY=...` (mismo que en payment-service)

### 2. Implementar Webhook
- [ ] Crear endpoint `POST /webhooks/subscription` en menuum-backend
- [ ] Ver ejemplo en `examples/menuum-backend-integration.py`
- [ ] Implementar lógica para actualizar `is_premium` en base de datos

### 3. Integrar Checkout
- [ ] Agregar rutas para crear checkout sessions
- [ ] Agregar rutas para consultar estado de suscripción
- [ ] Agregar rutas para cancelar suscripciones
- [ ] Ver ejemplos completos en `examples/menuum-backend-integration.py`

## Testing

### 1. Test Manual con curl
- [ ] Ejecutar `./examples/curl-examples.sh`
- [ ] Verificar que todos los endpoints respondan correctamente

### 2. Test de Integración con Stripe
- [ ] Instalar Stripe CLI: `brew install stripe/stripe-cli/stripe`
- [ ] Login: `stripe login`
- [ ] Forward webhooks: `stripe listen --forward-to localhost:8081/payments/webhook`
- [ ] Trigger test: `stripe trigger checkout.session.completed`
- [ ] Verificar logs: `docker logs -f payment-service`

### 3. Test de Checkout Real
- [ ] Crear checkout session desde menuum-backend
- [ ] Usar tarjeta de test: `4242 4242 4242 4242`
- [ ] Completar pago
- [ ] Verificar que se reciba webhook en menuum-backend
- [ ] Verificar que `is_premium` se actualice correctamente

## Verificación Final

- [ ] Health endpoint responde: `{"status":"healthy"}`
- [ ] Checkout session se crea correctamente
- [ ] Stripe webhook funciona (ver logs)
- [ ] menuum-backend recibe notificación
- [ ] Base de datos tiene subscripción creada
- [ ] Usuario puede acceder a features premium
- [ ] Cancelación funciona correctamente

## Troubleshooting Común

### "Missing API key"
- Verificar que `X-API-Key` header esté incluido
- Verificar que el valor coincida con `.env`

### "X-Tenant-ID header is required"
- Agregar header `X-Tenant-ID: menuum` a todas las peticiones

### Webhook no funciona
- Verificar `STRIPE_WEBHOOK_SECRET` en `.env`
- Para local, usar Stripe CLI o ngrok
- Revisar logs: `docker logs -f payment-service`

### Base de datos no conecta
- Verificar PostgreSQL: `docker ps`
- Verificar `DATABASE_URL` en `.env`
- Probar conexión: `psql $DATABASE_URL`

## Recursos

- **README.md** - Documentación completa
- **examples/menuum-backend-integration.py** - Código de integración Python
- **examples/curl-examples.sh** - Tests con curl
- **CHANGELOG.md** - Historial de cambios
- **Stripe Dashboard** - https://dashboard.stripe.com/

## Próximos Pasos

Una vez completado el setup:

1. [ ] Configurar webhooks para producción
2. [ ] Implementar tests automatizados
3. [ ] Configurar monitoreo y alertas
4. [ ] Documentar flujo de producción
5. [ ] Configurar CI/CD

---

**¿Problemas?** Consulta el README.md o revisa los logs del servicio.
