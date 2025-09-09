# Como usar no seu projeto

## ✅ **Sim, funciona perfeitamente com SQLC!**

### 📋 **Uso igual ao seu outro projeto:**

```go
// EXATAMENTE como você já usa:
queries, dbConn, err := connection.GetConnection(context.Background(), "meu_tenant", factory)
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}
defer dbConn.Close()

customerExists, err := services.GetCustomerByID(customerId.String())
```

## 🚀 **Formas de usar:**

### 1. **Forma Básica (igual ao seu padrão):**

```go
package main

import (
    "context"
    "database/sql"
    "net/http"

    connection "github.com/MK-Solutions-LTDA/tenant-connection"
    "your-project/services" // Seus services com SQLC
    "your-project/utils"
)

func YourHandler(tenant string, customerId uuid.UUID) error {
    // Factory do seu SQLC
    factory := func(db *sql.DB) *services.Queries {
        return services.NewQueries(db)
    }

    // USO IGUAL AO SEU OUTRO PROJETO
    queries, dbConn, err := connection.GetConnection(context.Background(), tenant, factory)
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, err)
    }
    defer dbConn.Close()

    // Usar services normalmente
    customerExists, err := services.GetCustomerByID(customerId.String())
    if err != nil {
        return utils.NewApiError(http.StatusNotFound, err)
    }

    return nil
}
```

### 2. **Com tenant no contexto (para APIs):**

```go
// No seu middleware, adicione o tenant ao contexto:
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenant := r.Header.Get("X-Tenant-ID") // ou de onde vier
        ctx := context.WithValue(r.Context(), connection.TenantContextKey, tenant)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// No seu handler:
func YourAPIHandler(w http.ResponseWriter, r *http.Request) {
    factory := func(db *sql.DB) *services.Queries {
        return services.NewQueries(db)
    }

    // Pega tenant automaticamente do contexto
    queries, dbConn, err := connection.GetConnectionFromContext(r.Context(), factory)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dbConn.Close()

    // Usar normalmente...
}
```

### 3. **Com configurações avançadas:**

```go
func YourAdvancedHandler(tenant string, customerId uuid.UUID) error {
    factory := func(db *sql.DB) *services.Queries {
        return services.NewQueries(db)
    }

    opts := connection.TenantConnectOptions{
        Tenant:          tenant,
        MaxOpenConns:    50,
        MaxIdleConns:    25,
        ForceUTC:        true,
        CacheEnabled:    true,
    }

    queries, dbConn, err := connection.GetConnectionWithOptions(context.Background(), factory, opts)
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, err)
    }
    defer dbConn.Close()

    customerExists, err := services.GetCustomerByID(customerId.String())
    if err != nil {
        return utils.NewApiError(http.StatusNotFound, err)
    }

    return nil
}
```

## 🔧 **Setup no seu go.mod:**

Apenas importe normalmente:

```go
import connection "github.com/MK-Solutions-LTDA/tenant-connection"
```

## ⚡ **Vantagens da v2:**

- ✅ **Mesmo padrão de uso** que você já conhece
- ✅ **Cache automático** de conexões
- ✅ **Pool de conexões configurável**
- ✅ **Health checks** automáticos
- ✅ **Logging** de queries opcional
- ✅ **Compatibilidade total** com SQLC
- ✅ **Cleanup automático** com `defer dbConn.Close()`

## 📝 **Sua chamada fica igual:**

```go
// Antes (seu outro projeto):
queries, dbConn, err := dbutils.GetConnection(context.Background())

// Agora (com tenant):
queries, dbConn, err := connection.GetConnection(context.Background(), tenant, factory)

// Resto igual:
defer dbConn.Close()
customerExists, err := services.GetCustomerByID(customerId.String())
```

**Só adiciona o `tenant` e o `factory` na chamada!** 🎯
