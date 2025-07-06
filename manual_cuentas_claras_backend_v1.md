# CuentasClaras Backend - Manual de Implementación Completo

## Estado Actual: ✅ COMPLETAMENTE FUNCIONAL

Versión actual implementada con autenticación, cuentas, categorías, transacciones y **PRÉSTAMOS** funcionando perfectamente.

---

## 🎯 **CARACTERÍSTICAS COMPLETADAS**

### ✅ **Autenticación Completa**

- JWT Access Token (10 minutos)
- Refresh Token Rotation (30 días)
- Login único por usuario
- WebSocket para force_logout
- Device session tracking con refresh_token_id
- Invalidación automática de tokens anteriores

### ✅ **Modelos Implementados**

- **User**: Usuarios con datos encriptados
- **DeviceSession**: Control de sesiones únicas
- **Account**: Cuentas financieras del usuario
- **Category**: Categorías para clasificar gastos
- **Transaction**: Registro central de movimientos (con constraints BD)
- **Loan**: Sistema completo de préstamos dados/recibidos ✨
- **LoanPayment**: Pagos de préstamos con confirmación manual ✨

### ✅ **Seguridad Robusta**

- Encriptación AES-256 de campos sensibles
- Middleware de autenticación robusto
- Constraints de BD para consistencia de datos
- Refresh token ligado a access token
- Desencriptación correcta en responses
- Validación de tokens en tiempo real

### ✅ **API Endpoints Funcionales**

```
Auth:
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
GET    /api/v1/auth/profile
PUT    /api/v1/auth/profile
POST   /api/v1/auth/logout

Accounts:
POST   /api/v1/accounts
GET    /api/v1/accounts
GET    /api/v1/accounts/:id
PUT    /api/v1/accounts/:id
DELETE /api/v1/accounts/:id

Categories:
POST   /api/v1/categories
GET    /api/v1/categories
GET    /api/v1/categories/:id
PUT    /api/v1/categories/:id
DELETE /api/v1/categories/:id

Transactions:
POST   /api/v1/transactions
GET    /api/v1/transactions (con filtros)
GET    /api/v1/transactions/:id
PUT    /api/v1/transactions/:id
DELETE /api/v1/transactions/:id

Loans: ✨ NUEVO - COMPLETAMENTE FUNCIONAL
POST   /api/v1/loans                     # Crear préstamo + transaction automática
GET    /api/v1/loans                     # Listar préstamos con balances
GET    /api/v1/loans/:id                 # Ver préstamo específico
PUT    /api/v1/loans/:id                 # Actualizar préstamo
DELETE /api/v1/loans/:id                 # Eliminar préstamo
POST   /api/v1/loans/:id/payments        # Crear payment (pendiente)
PUT    /api/v1/loan-payments/:id/confirm # Confirmar payment (crea transaction)
```

---

## 🔧 **CONFIGURACIÓN ACTUAL**

### **Tecnologías**

- **Backend**: Go + Fiber v2 + GORM
- **BD Desarrollo**: SQLite
- **BD Producción**: PostgreSQL (configurado)
- **Encriptación**: AES-256-GCM

### **Variables de Entorno (.env)**

```bash
# Database
DB_DRIVER=sqlite
DB_NAME_SQLITE=cuentas_claras.db

# Prod - PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=cuentas_claras_db

# Server
PORT=3000

# JWT
JWT_SECRET=tu-jwt-secret-super-seguro-aqui

# Encryption
ENCRYPTION_KEY=mi-clave-super-secreta-32-chars!!
```

### **Estructura del Proyecto**

```
cuentas-claras-backend/
├── main.go
├── .env
├── config/
│   ├── database.go
│   ├── migrate.go
│   └── constraints.go
├── models/
│   ├── user.go
│   ├── device_session.go
│   ├── account.go
│   ├── category.go
│   ├── transaction.go
│   ├── loan.go           ✨ NUEVO
│   └── loan_payment.go   ✨ NUEVO
├── handlers/
│   ├── auth.go
│   ├── user.go
│   ├── account.go
│   ├── category.go
│   ├── transaction.go
│   └── loan.go           ✨ NUEVO
├── middleware/
│   └── auth.go
├── utils/
│   ├── crypto.go
│   └── tokens.go
└── routes/
    └── routes.go
```

---

## 📋 **TODO: PRÓXIMAS IMPLEMENTACIONES**

### 🎯 **PRIORIDAD ALTA**

#### **✅ TODO-1: Sistema de Préstamos - COMPLETADO**

- ✅ Loan con encriptación de datos sensibles
- ✅ LoanPayment con confirmación manual
- ✅ Integración completa con Transaction
- ✅ Cálculos dinámicos de balance y status
- ✅ Endpoints CRUD completos
- ✅ Validaciones de negocio
- ✅ Flujo: Crear → Pagar → Confirmar

#### **TODO-2: Modelo RecurringExpense**

```go
// Gastos recurrentes (alquiler, servicios)
type RecurringExpense struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    AccountID   uint      `json:"account_id" gorm:"not null"`
    CategoryID  uint      `json:"category_id" gorm:"not null"`
    Amount      float64   `json:"amount" gorm:"not null"`
    Description string    `json:"description" gorm:"not null"`
    Frequency   string    `json:"frequency" gorm:"not null"` // daily, weekly, monthly, yearly
    NextDueDate time.Time `json:"next_due_date" gorm:"not null"`
    AutoGenerate bool     `json:"auto_generate" gorm:"default:false"`
    // ... resto de campos
}
```

**Endpoints necesarios:**

- `POST /api/v1/recurring-expenses` - Crear gasto recurrente
- `GET /api/v1/recurring-expenses` - Listar gastos recurrentes
- `POST /api/v1/recurring-expenses/:id/execute` - Ejecutar manualmente

#### **TODO-3: Modelo Reminder**

```go
// Sistema de recordatorios
type Reminder struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    Title       string    `json:"title" gorm:"not null"`
    Description string    `json:"description"`
    Type        string    `json:"type" gorm:"not null"`
    ReferenceID *uint     `json:"reference_id,omitempty"`
    RemindAt    time.Time `json:"remind_at" gorm:"not null"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    IsSent      bool      `json:"is_sent" gorm:"default:false"`
    // ... resto de campos
}
```

### 🎯 **PRIORIDAD MEDIA**

#### **TODO-4: Endpoints Adicionales de Account**

- `GET /api/v1/accounts/:id/balance` - Balance específico con filtros
- `GET /api/v1/accounts/:id/history` - Historial detallado de transacciones

#### **TODO-5: Sistema de Reportes**

```
GET /api/v1/reports/summary        # Resumen general
GET /api/v1/reports/monthly/:year/:month # Reporte mensual
GET /api/v1/reports/categories     # Gastos por categoría
GET /api/v1/reports/accounts       # Balances por cuenta
GET /api/v1/reports/loans          # Estado de préstamos
```

#### **TODO-6: Validaciones Adicionales**

- Validar monedas ISO (PEN, USD, EUR)
- Validar que account_id y category_id pertenezcan al usuario
- Límites de transacciones por día/mes

### 🎯 **PRIORIDAD BAJA**

#### **TODO-7: Sistema de Notificaciones**

- Integración con Firebase Cloud Messaging (FCM)
- WebSocket para notificaciones real-time
- Configuración de quiet hours

#### **TODO-8: Optimizaciones**

- Índices de BD para queries frecuentes
- Paginación en listados
- Cache de balances frecuentes

#### **TODO-9: Logs y Auditoría**

- Log de operaciones financieras
- Auditoría de cambios importantes
- Métricas de uso

#### **TODO-10: Testing**

- Unit tests para handlers
- Integration tests para endpoints
- Tests de constraints de BD

---

## 🚀 **CÓMO EJECUTAR**

### **Desarrollo**

```bash
# Instalar dependencias
go mod tidy

# Configurar .env (usar SQLite)
DB_DRIVER=sqlite
DB_NAME_SQLITE=cuentas_claras.db

# Ejecutar
go run main.go
```

### **Producción**

```bash
# Configurar .env (usar PostgreSQL)
DB_DRIVER=postgres
DB_HOST=tu-host-postgres
DB_NAME=cuentas_claras_db

# Ejecutar
go build
./cuentas-claras
```

---

## 📝 **NOTAS IMPORTANTES**

### **Seguridad Implementada**

- Todos los campos sensibles están encriptados (AES-256-GCM)
- JWT tokens con refresh rotation
- Login único garantizado por refresh_token_id
- Constraints de BD para consistencia total
- Middleware de autenticación robusto

### **Encriptación Automática**

Los siguientes campos se encriptan automáticamente:

- `User.name_encrypted`
- `User.phone_encrypted`
- `Transaction.notes_encrypted`
- `Loan.description_encrypted` ✨
- `Loan.person_name_encrypted` ✨
- `Loan.notes_encrypted` ✨
- `LoanPayment.description_encrypted` ✨
- `LoanPayment.notes_encrypted` ✨

### **Sistema de Préstamos Implementado** ✨

**Flujo completo:**

1. **Crear Loan** → Genera Transaction automática (dinero entra/sale)
2. **Crear LoanPayment** → Registro pendiente (sin Transaction)
3. **Confirmar Payment** → Crea Transaction y actualiza status
4. **Cálculos automáticos** → Balance y status desde transacciones

**Casos soportados:**

- Préstamos dados (type: "given")
- Préstamos recibidos (type: "received")
- Pagos parciales y completos
- Encriptación de datos sensibles
- Validaciones de negocio

### **Constraints de BD Implementados**

```sql
-- Direction solo 'in' o 'out'
-- Amount positivo para 'in', negativo para 'out'
-- Type consistente con direction
-- Amount no puede ser cero
-- Types válidos solamente
-- Refresh token único por sesión activa
```

### **Balance Dinámico**

Los balances se calculan en tiempo real desde transacciones:

```sql
SELECT SUM(amount) as balance
FROM transactions
WHERE account_id = ? AND user_id = ? AND deleted_at IS NULL
```

### **Problemas Resueltos** ✅

- ✅ Refresh token funcionando correctamente
- ✅ Encriptación/desencriptación en responses
- ✅ Hooks GORM manejados apropiadamente
- ✅ Balance dinámico correcto con soft delete
- ✅ Login único con invalidación automática
- ✅ Constraints de BD para consistencia total

---

## 🧪 **TESTING - COMANDOS COMPLETOS**

### **Setup inicial:**

```bash
# Registro
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@test.com","password":"123456"}'

# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"123456","device_info":{"device_id":"test"}}'

export TOKEN="tu_access_token_aqui"
```

### **Crear datos básicos:**

```bash
# Cuenta
curl -X POST http://localhost:3000/api/v1/accounts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Cuenta Principal","type":"bank","currency":"PEN"}'

# Categoría
curl -X POST http://localhost:3000/api/v1/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Préstamos","color":"#FF5722"}'
```

### **Sistema de Préstamos:**

```bash
# Crear préstamo DADO
curl -X POST http://localhost:3000/api/v1/loans \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "amount": 5000,
    "description": "Préstamo de emergencia para Juan",
    "person_name": "Juan Pérez García",
    "type": "given",
    "loan_date": "2025-01-15T00:00:00Z",
    "notes": "Préstamo sin interés, pago en 3 meses"
  }'

# Listar préstamos
curl -X GET http://localhost:3000/api/v1/loans \
  -H "Authorization: Bearer $TOKEN"

# Crear pago
curl -X POST http://localhost:3000/api/v1/loans/1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "amount": 1500,
    "date": "2025-02-15T00:00:00Z",
    "description": "Primer pago parcial del préstamo",
    "notes": "Pago recibido en efectivo"
  }'

# Confirmar pago (crea transaction)
curl -X PUT http://localhost:3000/api/v1/loan-payments/1/confirm \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🔄 **PRÓXIMO PASO RECOMENDADO**

**Implementar Modelo RecurringExpense** para completar los gastos recurrentes (alquiler, servicios), que es la siguiente característica importante de CuentasClaras.

**¿Qué tenemos ahora?**

- ✅ Sistema de usuarios completo
- ✅ Cuentas y categorías
- ✅ Transacciones con constraints
- ✅ Préstamos completos con pagos
- 🔄 **SIGUIENTE**: Gastos recurrentes + recordatorios

---

## 📊 **ARQUITECTURA ACTUAL**

```
┌─────────────────┐    HTTP/REST API    ┌─────────────────┐
│   Flutter App   │◄──────────────────►│   Go Backend    │
│   (Mobile)      │                     │   (Fiber)       │
└─────────────────┘                     └─────────────────┘
        │                                         │
        │ WebSocket                               │ GORM
        │ (Real-time)                             ▼
        └─────────────────────────────────►┌─────────────────┐
                                           │   SQLite/       │
                                           │   PostgreSQL    │
                                           └─────────────────┘
```

**Características técnicas:**

- ✅ Login único con device sessions
- ✅ Refresh token rotation automática
- ✅ Encriptación AES-256 de datos sensibles
- ✅ Balance dinámico calculado en tiempo real
- ✅ Constraints de BD para consistencia
- ✅ Soft delete en todos los modelos
- ✅ WebSocket preparado para notificaciones

---

## 📞 **SOPORTE**

- **Documentación completa**: Ver `guia_cuentas_claras.md`
- **API Testing**: Usar Postman/curl con Bearer tokens
- **Base de datos**: SQLite para desarrollo, PostgreSQL para producción
- **Comandos de prueba**: Incluidos en este manual
- **Estado actual**: Sistema de préstamos completamente funcional

---

## 🎉 **LOGROS ALCANZADOS**

### **✅ Sistema Financiero Robusto:**

- Control total de ingresos y gastos
- Múltiples cuentas con diferentes monedas
- Sistema completo de préstamos dados/recibidos
- Cálculos automáticos de balances y estados
- Encriptación de datos sensibles

### **✅ Seguridad de Nivel Empresarial:**

- Autenticación JWT con refresh rotation
- Login único con control de dispositivos
- Encriptación AES-256 de campos sensibles
- Constraints de BD para consistencia
- Middleware de autenticación robusto

### **✅ Arquitectura Escalable:**

- Modelos bien estructurados
- API RESTful consistente
- Separación clara de responsabilidades
- Configuración flexible (SQLite/PostgreSQL)
- Preparado para notificaciones y WebSocket

**CuentasClaras** es ahora una aplicación financiera completa y segura! 💰🔐🚀
