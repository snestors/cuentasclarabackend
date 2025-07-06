# CuentasClaras Backend - Manual de Implementación

## Estado Actual: ✅ FUNCIONAL

Versión actual implementada con autenticación, cuentas, categorías y transacciones funcionando.

---

## 🎯 **CARACTERÍSTICAS COMPLETADAS**

### ✅ **Autenticación Completa**

- JWT Access Token (10 minutos)
- Refresh Token Rotation (30 días)
- Login único por usuario
- WebSocket para force_logout
- Device session tracking

### ✅ **Modelos Implementados**

- **User**: Usuarios con datos encriptados
- **DeviceSession**: Control de sesiones únicas
- **Account**: Cuentas financieras del usuario
- **Category**: Categorías para clasificar gastos
- **Transaction**: Registro central de movimientos (con constraints BD)

### ✅ **Seguridad**

- Encriptación AES-256 de campos sensibles
- Middleware de autenticación robusto
- Constraints de BD para consistencia de datos
- Refresh token ligado a access token

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
```

---

## 🔧 **CONFIGURACIÓN ACTUAL**

### **Tecnologías**

- **Backend**: Go + Fiber v2 + GORM
- **BD Desarrollo**: SQLite
- **BD Producción**: PostgreSQL (configurado)
- **Encriptación**: AES-256

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
│   └── transaction.go
├── handlers/
│   ├── auth.go
│   ├── user.go
│   ├── account.go
│   ├── category.go
│   └── transaction.go
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

#### **TODO-1: Modelo Loan**

```go
// Implementar modelo de préstamos
type Loan struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    Amount      float64   `json:"amount" gorm:"not null"`
    Description string    `json:"description" gorm:"not null"`
    PersonName  string    `json:"person_name" gorm:"column:person_name_encrypted"`
    Type        string    `json:"type" gorm:"not null"` // 'given' or 'received'
    Status      string    `json:"status" gorm:"default:'pending'"`
    LoanDate    time.Time `json:"loan_date" gorm:"not null"`
    DueDate     *time.Time `json:"due_date,omitempty"`
    Notes       string    `json:"notes" gorm:"column:notes_encrypted"`
    // ... resto de campos
}
```

**Endpoints necesarios:**

- `POST /api/v1/loans` - Crear préstamo
- `GET /api/v1/loans` - Listar préstamos
- `POST /api/v1/loans/:id/payments` - Registrar pago
- `GET /api/v1/loans/:id/payments` - Historial de pagos

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

### **Seguridad**

- Todos los campos sensibles están encriptados (AES-256)
- JWT tokens con refresh rotation
- Login único garantizado
- Constraints de BD para consistencia

### **Encriptación Automática**

Los siguientes campos se encriptan automáticamente:

- `User.name_encrypted`
- `User.phone_encrypted`
- `Transaction.notes_encrypted`
- `Loan.person_name_encrypted` (TODO)
- `Loan.notes_encrypted` (TODO)

### **Constraints de BD Implementados**

```sql
-- Direction solo 'in' o 'out'
-- Amount positivo para 'in', negativo para 'out'
-- Type consistente con direction
-- Amount no puede ser cero
-- Types válidos solamente
```

### **Balance Dinámico**

Los balances se calculan en tiempo real desde transacciones:

```sql
SELECT SUM(amount) as balance
FROM transactions
WHERE account_id = ? AND user_id = ?
```

---

## 🔄 **PRÓXIMO PASO RECOMENDADO**

**Implementar Modelo Loan** para completar la funcionalidad de préstamos, que es una característica central de CuentasClaras.

---

## 📞 **SOPORTE**

- **Documentación completa**: Ver `guia_cuentas_claras.md`
- **API Testing**: Usar Postman/curl con Bearer tokens
- **Base de datos**: SQLite para desarrollo, PostgreSQL para producción
