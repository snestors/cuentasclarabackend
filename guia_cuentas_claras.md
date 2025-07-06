# CuentasClaras - Guía Completa de Implementación

## Índice

1. [Descripción General](#descripción-general)
2. [Arquitectura del Sistema](#arquitectura-del-sistema)
3. [Modelos de Datos](#modelos-de-datos)
4. [Sistema de Autenticación](#sistema-de-autenticación)
5. [Lógica de Negocio](#lógica-de-negocio)
6. [API Endpoints](#api-endpoints)
7. [Seguridad](#seguridad)
8. [Notificaciones](#notificaciones)
9. [Estructura del Proyecto](#estructura-del-proyecto)
10. [Plan de Implementación](#plan-de-implementación)

---

## Descripción General

### Propósito

**CuentasClaras** es una aplicación móvil para **organizar y controlar gastos personales** sin gestión financiera activa ni recomendaciones. Solo seguimiento y organización de información financiera personal con total transparencia.

### Características Principales

- **Control de Gastos**: Registro de ingresos y gastos por categorías
- **Múltiples Cuentas**: Manejo de diferentes cuentas con distintas monedas
- **Gastos Recurrentes**: Configuración de gastos que se repiten con recordatorios
- **Préstamos**: Control de dinero prestado a terceros y recibido de otros
- **Recordatorios**: Notificaciones para gastos y vencimientos
- **Reportes Básicos**: Estadísticas simples sin recomendaciones financieras

### Stack Tecnológico

- **Frontend**: Flutter 3.22
- **Backend**: Go + Fiber v2 + GORM
- **Base de Datos**: PostgreSQL
- **Autenticación**: JWT + WebSocket
- **Notificaciones**: Firebase Cloud Messaging (FCM)

---

## Arquitectura del Sistema

### Diagrama de Arquitectura

```
┌─────────────────┐    HTTP/REST API    ┌─────────────────┐
│   Flutter App   │◄──────────────────►│   Go Backend    │
│   (Mobile)      │                     │   (Fiber)       │
└─────────────────┘                     └─────────────────┘
        │                                         │
        │ WebSocket                               │ GORM
        │ (Real-time)                             ▼
        └─────────────────────────────────►┌─────────────────┐
                                           │   PostgreSQL    │
                                           │   (Database)    │
                                           └─────────────────┘
```

### Principios de Diseño

- **Simplicidad**: Soluciones simples antes que complejas
- **Consistencia**: Enfoque contable con libro de transacciones
- **Seguridad**: Datos sensibles encriptados, login único
- **Escalabilidad**: Estructura preparada para crecimiento futuro

---

## Modelos de Datos

### Modelo Conceptual

El sistema se basa en un **enfoque contable** donde todas las transacciones financieras se registran en un libro único, y los balances se calculan dinámicamente.

### 1. Usuario (User)

**Propósito**: Información básica del usuario y configuraciones.

```go
type User struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
    Email       string    `json:"email" gorm:"unique;not null" validate:"required,email"`
    Password    string    `json:"-" gorm:"not null" validate:"required,min=6"`
    PhoneNumber string    `json:"phone_number" gorm:"column:phone_encrypted"` // 🔒 ENCRIPTADO

    // Configuraciones
    NotificationsEnabled   bool   `json:"notifications_enabled" gorm:"default:true"`
    PushNotifications     bool   `json:"push_notifications" gorm:"default:true"`
    InAppNotifications    bool   `json:"in_app_notifications" gorm:"default:true"` // No se puede desactivar
    Timezone              string `json:"timezone" gorm:"default:'America/Lima'"`
    QuietHoursStart       *time.Time `json:"quiet_hours_start,omitempty"`
    QuietHoursEnd         *time.Time `json:"quiet_hours_end,omitempty"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 2. Sesión de Dispositivo (DeviceSession)

**Propósito**: Control de login único y gestión de dispositivos.

```go
type DeviceSession struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`

    // Información del dispositivo
    DeviceID    string    `json:"device_id" gorm:"not null"`
    DeviceName  string    `json:"device_name" gorm:"not null"`
    DeviceType  string    `json:"device_type" gorm:"not null" validate:"oneof=ios android web"`
    DeviceModel string    `json:"device_model"`
    OSVersion   string    `json:"os_version"`

    // Información de red
    IPAddress   string    `json:"ip_address"`
    Country     string    `json:"country"`
    City        string    `json:"city"`
    UserAgent   string    `json:"user_agent"`

    // Tokens y control
    RefreshTokenHash string    `json:"-" gorm:"not null"`
    FCMToken         string    `json:"-" gorm:"column:fcm_token_encrypted"` // 🔒 ENCRIPTADO
    IsFCMActive      bool      `json:"is_fcm_active" gorm:"default:true"`
    IsActive         bool      `json:"is_active" gorm:"default:true"` // Solo UNA = true por usuario

    // Actividad
    LoginAt      time.Time  `json:"login_at"`
    LogoutAt     *time.Time `json:"logout_at,omitempty"`
    LogoutReason string     `json:"logout_reason,omitempty"`
    LastActivity time.Time  `json:"last_activity"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`

    // Relaciones
    User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Constraint: Solo una sesión activa por usuario
// UNIQUE INDEX ON (user_id) WHERE is_active = true
```

### 3. Cuenta (Account)

**Propósito**: Representar diferentes cuentas financieras del usuario.

```go
type Account struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    Name        string    `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
    Type        string    `json:"type" gorm:"not null" validate:"required,oneof=cash bank credit_card savings"`
    Currency    string    `json:"currency" gorm:"not null" validate:"required,len=3"` // PEN, USD, EUR
    Description string    `json:"description" gorm:"size:255"`

    // Personalización visual
    Color       string    `json:"color" gorm:"default:'#2196F3'"`
    Icon        string    `json:"icon" gorm:"default:'account_balance_wallet'"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`

    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relaciones
    User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:AccountID"`
}

// ❌ SIN campo balance - se calcula desde transactions
```

### 4. Categoría (Category)

**Propósito**: Clasificar gastos e ingresos para mejor organización.

```go
type Category struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    Name        string    `json:"name" gorm:"not null" validate:"required,min=2,max=50"`
    Description string    `json:"description" gorm:"size:255"`

    // Personalización visual
    Color       string    `json:"color" gorm:"default:'#4CAF50'"`
    Icon        string    `json:"icon" gorm:"default:'category'"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`

    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relaciones
    User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:CategoryID"`
}
```

### 5. Transacción (Transaction) - MODELO CENTRAL

**Propósito**: Registro unificado de todos los movimientos financieros.

```go
type Transaction struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    AccountID   uint      `json:"account_id" gorm:"not null"`

    // Información financiera
    Amount      float64   `json:"amount" gorm:"not null" validate:"required,ne=0"` // Con signo: + ingresos, - gastos
    Direction   string    `json:"direction" gorm:"not null" validate:"required,oneof=in out"`
    Description string    `json:"description" gorm:"not null" validate:"required,min=1,max=255"`
    Date        time.Time `json:"date" gorm:"not null"`
    Notes       string    `json:"notes" gorm:"size:500;column:notes_encrypted"` // 🔒 ENCRIPTADO

    // Clasificación
    Type        string    `json:"type" gorm:"not null" validate:"required,oneof=income expense loan_given loan_received loan_payment_given loan_payment_received"`
    CategoryID  *uint     `json:"category_id,omitempty"` // Opcional para préstamos

    // Referencias a otros objetos
    ReferenceID   *uint  `json:"reference_id,omitempty"`
    ReferenceType string `json:"reference_type,omitempty" validate:"omitempty,oneof=loan recurring_expense"`

    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relaciones
    User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Account  Account   `json:"account,omitempty" gorm:"foreignKey:AccountID"`
    Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}

// Constraints de consistencia
// CHECK ((direction = 'in' AND amount > 0) OR (direction = 'out' AND amount < 0))
// CHECK ((type IN ('income', 'loan_received', 'loan_payment_received') AND direction = 'in') OR
//        (type IN ('expense', 'loan_given', 'loan_payment_given') AND direction = 'out'))
```

### 6. Préstamo (Loan)

**Propósito**: Configuración y seguimiento de préstamos dados y recibidos.

```go
type Loan struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`

    // Información básica
    Amount      float64   `json:"amount" gorm:"not null" validate:"required,gt=0"`
    Description string    `json:"description" gorm:"not null" validate:"required,min=1,max=255"`
    PersonName  string    `json:"person_name" gorm:"column:person_name_encrypted"` // 🔒 ENCRIPTADO
    Type        string    `json:"type" gorm:"not null" validate:"required,oneof=given received"`

    // Control y fechas
    Status       string     `json:"status" gorm:"default:'pending'" validate:"oneof=pending partial_paid paid"`
    LoanDate     time.Time  `json:"loan_date" gorm:"not null"`
    DueDate      *time.Time `json:"due_date,omitempty"`
    InterestRate float64    `json:"interest_rate" gorm:"default:0"`
    Notes        string     `json:"notes" gorm:"size:500;column:notes_encrypted"` // 🔒 ENCRIPTADO

    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`

    // Relaciones
    User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ❌ SIN amount_paid/amount_owed - se calculan desde transactions
```

### 7. Gasto Recurrente (RecurringExpense)

**Propósito**: Configurar gastos que se repiten automáticamente.

```go
type RecurringExpense struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    AccountID   uint      `json:"account_id" gorm:"not null"`
    CategoryID  uint      `json:"category_id" gorm:"not null"`

    // Configuración del gasto
    Amount      float64   `json:"amount" gorm:"not null" validate:"required,gt=0"`
    Description string    `json:"description" gorm:"not null" validate:"required,min=1,max=255"`

    // Configuración de recurrencia
    Frequency    string    `json:"frequency" gorm:"not null" validate:"required,oneof=daily weekly monthly yearly"`
    StartDate    time.Time `json:"start_date" gorm:"not null"`
    EndDate      *time.Time `json:"end_date,omitempty"`
    NextDueDate  time.Time `json:"next_due_date" gorm:"not null"`

    // Control
    IsActive     bool      `json:"is_active" gorm:"default:true"`
    AutoGenerate bool      `json:"auto_generate" gorm:"default:false"` // Solo recordatorio, NO transacción automática
    Notes        string    `json:"notes" gorm:"size:500"`

    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`

    // Relaciones
    User     User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Account  Account  `json:"account,omitempty" gorm:"foreignKey:AccountID"`
    Category Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}
```

### 8. Recordatorio (Reminder)

**Propósito**: Sistema de notificaciones y alertas.

```go
type Reminder struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    UserID      uint      `json:"user_id" gorm:"not null"`
    Title       string    `json:"title" gorm:"not null" validate:"required,min=1,max=100"`
    Description string    `json:"description" gorm:"size:255"`
    Type        string    `json:"type" gorm:"not null" validate:"required,oneof=expense loan recurring_expense custom"`

    // Referencia al objeto relacionado
    ReferenceID   *uint  `json:"reference_id,omitempty"`
    ReferenceType string `json:"reference_type,omitempty" validate:"omitempty,oneof=expense loan recurring_expense"`

    // Configuración de recordatorio
    RemindAt    time.Time `json:"remind_at" gorm:"not null"`
    Frequency   string    `json:"frequency" gorm:"default:'once'" validate:"oneof=once daily weekly monthly"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    IsSent      bool      `json:"is_sent" gorm:"default:false"`

    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relaciones
    User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
```

---

## Sistema de Autenticación

### Enfoque: Login Único

- **Un usuario = Una sesión activa** en todo momento
- Nueva sesión cierra automáticamente la anterior
- Control en tiempo real vía WebSocket

### Flujo de Autenticación

#### 1. Registro de Usuario

```http
POST /auth/register
{
  "name": "Juan Pérez",
  "email": "juan@email.com",
  "password": "password123",
  "phone_number": "+51987654321"
}
```

#### 2. Login con Control de Dispositivo

```http
POST /auth/login
{
  "email": "juan@email.com",
  "password": "password123",
  "device_info": {
    "device_id": "uuid-generado-en-app",
    "device_name": "iPhone 13 Pro",
    "device_type": "ios",
    "os_version": "17.1",
    "fcm_token": "firebase-token-here"
  }
}
```

**Response:**

```json
{
  "access_token": "jwt-token",
  "refresh_token": "refresh-token",
  "device_session_id": 123,
  "user": { ... },
  "is_new_device": true
}
```

#### 3. Control de Sesión Única

Cuando hay login desde nuevo dispositivo:

1. Backend busca sesión activa existente
2. Si existe, envía `force_logout` vía WebSocket al dispositivo anterior
3. Marca sesión anterior como inactiva
4. Crea nueva sesión activa

### Tokens JWT

- **Access Token**: 15-30 minutos de duración
- **Refresh Token**: 7-30 días de duración
- **Refresh Token Rotation**: Se cambia en cada uso

### WebSocket para Control Real-Time

```
wss://api.app.com/ws?token=jwt_access_token
```

**Eventos WebSocket:**

```json
// Forzar logout
{
  "type": "force_logout",
  "message": "Nueva sesión iniciada desde otro dispositivo",
  "timestamp": "2025-01-15T10:30:00Z"
}

// Notificaciones en tiempo real
{
  "type": "notification",
  "data": {
    "title": "Recordatorio",
    "body": "Tienes un gasto recurrente pendiente"
  }
}

// Heartbeat
{
  "type": "ping",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

---

## Lógica de Negocio

### Cálculo de Balances

**Principio**: El balance se calcula dinámicamente desde las transacciones, NO se almacena.

```sql
-- Balance de una cuenta
SELECT SUM(amount) as balance
FROM transactions
WHERE account_id = ? AND user_id = ?

-- Balance por rango de fechas
SELECT SUM(amount) as balance
FROM transactions
WHERE account_id = ? AND user_id = ?
  AND date BETWEEN ? AND ?
```

### Gestión de Préstamos

#### Préstamo Dado (Yo presto dinero)

1. **Crear préstamo**: `Loan` con `type = 'given'`
2. **Transacción inicial**:
   - `amount = -2000, direction = 'out', type = 'loan_given'`
   - `reference_id = loan.id, reference_type = 'loan'`

#### Pago de Préstamo (Me pagan)

1. **Registrar pago**:
   - `amount = +500, direction = 'in', type = 'loan_payment_received'`
   - `reference_id = loan.id`

#### Cálculo automático de estado

```sql
-- Calcular monto pagado
SELECT SUM(amount) as amount_paid
FROM transactions
WHERE reference_id = ? AND type = 'loan_payment_received'

-- Actualizar status automáticamente
UPDATE loans SET status = CASE
  WHEN amount_paid = 0 THEN 'pending'
  WHEN amount_paid < amount THEN 'partial_paid'
  ELSE 'paid'
END
```

### Gastos Recurrentes

**Importante**: NO genera transacciones automáticamente, solo recordatorios.

#### Flujo:

1. Usuario configura gasto recurrente
2. Sistema crea recordatorios basados en frecuencia
3. Usuario ve recordatorio y confirma
4. Se crea transacción manual

#### Actualización de próxima fecha

```go
func UpdateNextDueDate(recurringExpense *RecurringExpense) {
    switch recurringExpense.Frequency {
    case "daily":
        recurringExpense.NextDueDate = recurringExpense.NextDueDate.AddDate(0, 0, 1)
    case "weekly":
        recurringExpense.NextDueDate = recurringExpense.NextDueDate.AddDate(0, 0, 7)
    case "monthly":
        recurringExpense.NextDueDate = recurringExpense.NextDueDate.AddDate(0, 1, 0)
    case "yearly":
        recurringExpense.NextDueDate = recurringExpense.NextDueDate.AddDate(1, 0, 0)
    }
}
```

### Validaciones de Negocio

#### Transacciones

- `amount` y `direction` deben ser consistentes
- `type` debe coincidir con `direction`
- `account_id` debe pertenecer al usuario
- `category_id` debe existir y pertenecer al usuario

#### Préstamos

- `person_name` requerido y no vacío
- `amount > 0`
- `loan_date <= now()`
- Pagos no pueden exceder el monto del préstamo

#### Cuentas

- `currency` debe ser código ISO válido (PEN, USD, EUR)
- `name` único por usuario
- `type` debe ser valor válido

---

## API Endpoints

### Autenticación

```http
POST   /auth/register          # Registro de usuario
POST   /auth/login             # Login con device info
POST   /auth/refresh           # Renovar access token
POST   /auth/logout            # Cerrar sesión actual
DELETE /auth/logout-all        # Cerrar todas las sesiones
GET    /auth/sessions          # Ver sesiones activas
DELETE /auth/sessions/{id}     # Cerrar sesión específica
PUT    /auth/fcm-token         # Actualizar FCM token
```

### Usuarios

```http
GET    /users/profile          # Obtener perfil
PUT    /users/profile          # Actualizar perfil
PUT    /users/notifications    # Configurar notificaciones
DELETE /users/account          # Eliminar cuenta
```

### Cuentas

```http
GET    /accounts               # Listar cuentas del usuario
POST   /accounts               # Crear nueva cuenta
GET    /accounts/{id}          # Obtener cuenta específica
PUT    /accounts/{id}          # Actualizar cuenta
DELETE /accounts/{id}          # Eliminar cuenta (soft delete)
GET    /accounts/{id}/balance  # Obtener balance actual
GET    /accounts/{id}/history  # Historial de transacciones
```

### Categorías

```http
GET    /categories             # Listar categorías del usuario
POST   /categories             # Crear nueva categoría
GET    /categories/{id}        # Obtener categoría específica
PUT    /categories/{id}        # Actualizar categoría
DELETE /categories/{id}        # Eliminar categoría (soft delete)
```

### Transacciones

```http
GET    /transactions           # Listar transacciones (con filtros)
POST   /transactions           # Crear nueva transacción
GET    /transactions/{id}      # Obtener transacción específica
PUT    /transactions/{id}      # Actualizar transacción
DELETE /transactions/{id}      # Eliminar transacción

# Filtros disponibles
GET    /transactions?account_id=1&category_id=2&date_from=2025-01-01&date_to=2025-01-31&type=expense
```

### Préstamos

```http
GET    /loans                  # Listar préstamos del usuario
POST   /loans                  # Crear nuevo préstamo
GET    /loans/{id}             # Obtener préstamo específico
PUT    /loans/{id}             # Actualizar préstamo
DELETE /loans/{id}             # Eliminar préstamo
POST   /loans/{id}/payments    # Registrar pago de préstamo
GET    /loans/{id}/payments    # Historial de pagos
```

### Gastos Recurrentes

```http
GET    /recurring-expenses     # Listar gastos recurrentes
POST   /recurring-expenses     # Crear gasto recurrente
GET    /recurring-expenses/{id} # Obtener gasto recurrente específico
PUT    /recurring-expenses/{id} # Actualizar gasto recurrente
DELETE /recurring-expenses/{id} # Eliminar gasto recurrente
POST   /recurring-expenses/{id}/execute # Ejecutar gasto recurrente manualmente
```

### Recordatorios

```http
GET    /reminders              # Listar recordatorios del usuario
POST   /reminders              # Crear nuevo recordatorio
GET    /reminders/{id}         # Obtener recordatorio específico
PUT    /reminders/{id}         # Actualizar recordatorio
DELETE /reminders/{id}         # Eliminar recordatorio
PUT    /reminders/{id}/mark-sent # Marcar como enviado
```

### Reportes

```http
GET    /reports/summary        # Resumen general
GET    /reports/monthly/{year}/{month} # Reporte mensual
GET    /reports/categories     # Gastos por categoría
GET    /reports/accounts       # Balances por cuenta
GET    /reports/loans          # Estado de préstamos
```

---

## Seguridad

### Encriptación de Datos Sensibles

**Campos encriptados a nivel columna (AES-256):**

- `User.phone_number`
- `DeviceSession.fcm_token`
- `Transaction.notes`
- `Loan.person_name`
- `Loan.notes`

### Implementación de Encriptación

```go
// Hooks automáticos en GORM
func (u *User) BeforeSave(tx *gorm.DB) error {
    if u.PhoneNumber != "" {
        u.PhoneNumber = EncryptField(u.PhoneNumber)
    }
    return nil
}

func (u *User) AfterFind(tx *gorm.DB) error {
    if u.PhoneNumber != "" {
        u.PhoneNumber = DecryptField(u.PhoneNumber)
    }
    return nil
}
```

### Validaciones de Seguridad

- **Rate Limiting**: 100 requests por hora por usuario
- **Input Validation**: Sanitización de todos los inputs
- **SQL Injection**: Uso de GORM con prepared statements
- **XSS Protection**: Sanitización de outputs
- **CORS**: Configurado específicamente para la app móvil

### Auditoría y Logs

**Eventos loggeados:**

- Login attempts (exitosos y fallidos)
- Operaciones financieras (crear/editar transacciones)
- Cambios en préstamos
- Accesos desde IPs/dispositivos nuevos

**NO loggeados:**

- Passwords
- Montos específicos (solo que hubo transacción)
- Datos personales de terceros

### Gestión de Errores

```go
// ❌ Malo - revela información
return errors.New("User with email user@email.com not found")

// ✅ Bueno - genérico
return errors.New("Invalid credentials")
```

---

## Notificaciones

### Sistema Híbrido: Push + In-App

#### Notificaciones Push (FCM)

- **Configurables**: Usuario puede desactivar completamente
- **Horarios**: Respeta quiet hours configurados
- **Tipos**: Recordatorios, vencimientos, nuevos logins

#### Notificaciones In-App

- **NO configurables**: Siempre activas
- **Críticas**: Recordatorios de gastos recurrentes
- **Aparecen**: Cuando usuario abre la aplicación

### Configuración por Usuario

```json
{
  "push_notifications": true,
  "quiet_hours_start": "22:00",
  "quiet_hours_end": "08:00",
  "notification_types": {
    "recurring_expenses": true,
    "loan_reminders": true,
    "login_alerts": true,
    "monthly_reports": false
  }
}
```

### Implementación de Notificaciones

```go
type NotificationService struct {
    FCMClient *messaging.Client
}

func (ns *NotificationService) SendReminder(userID uint, title, body string) {
    // 1. Obtener configuraciones del usuario
    user := getUserByID(userID)

    // 2. Verificar si push notifications están habilitadas
    if !user.PushNotifications {
        // Solo crear notificación in-app
        createInAppNotification(userID, title, body)
        return
    }

    // 3. Verificar quiet hours
    if isQuietHour(user.QuietHoursStart, user.QuietHoursEnd) {
        // Programar para después de quiet hours
        scheduleNotification(userID, title, body, user.QuietHoursEnd)
        return
    }

    // 4. Enviar push notification
    sendPushNotification(user.FCMToken, title, body)

    // 5. Crear notificación in-app como backup
    createInAppNotification(userID, title, body)
}
```

### Tipos de Notificaciones

#### Recordatorios de Gastos Recurrentes

```json
{
  "title": "Recordatorio de Gasto",
  "body": "Tienes pendiente: Alquiler - S/ 1,200",
  "type": "recurring_expense",
  "reference_id": 123
}
```

#### Vencimientos de Préstamos

```json
{
  "title": "Préstamo por Vencer",
  "body": "Juan debe pagarte S/ 500 mañana",
  "type": "loan_reminder",
  "reference_id": 456
}
```

#### Alertas de Seguridad

```json
{
  "title": "Nuevo Inicio de Sesión",
  "body": "Sesión iniciada desde iPhone 13 Pro",
  "type": "security_alert"
}
```

---

## Estructura del Proyecto

### Backend (Go)
