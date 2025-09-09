# 🔧 Atualização do dbutils com Close() - Nova Assinatura V2

## 📋 **Seu código atual:**

```go
func GetTenantConnection(tenant string) (*db.Queries, error)
```

## 🚀 **Nova assinatura V2 com Close():**

```go
func GetTenantConnection(tenant string) (*db.Queries, *connection.TenantConnectionV2, error)
```

---

## 💻 **Código atualizado completo:**

```go
package dbutils

import (
    "account-payable-service/internal/config"
    "account-payable-service/internal/db"
    "account-payable-service/internal/logger"
    "account-payable-service/internal/utils"
    "context"
    "errors"

    connection "github.com/MK-Solutions-LTDA/tenant-connection"
)

// Nova assinatura V2 - retorna conexão para Close()
func GetTenantConnection(tenant string) (*db.Queries, *connection.TenantConnectionV2, error) {
    ctx := context.Background()

    opts := connection.TenantConnectOptions{
        Tenant:       tenant,
        CacheEnabled: true,
        MaxOpenConns: 25,
        MaxIdleConns: 25,
        ForceUTC:     true,
    }

    tenantConn, err := connection.GetTenantConnectionV2(ctx, opts)
    if err != nil {
        logger.Error.Println("::::::GetTenantConnection - Ocorreu um erro ao conectar ao tenant::::::", tenant, err)
        utils.SendDiscordMessage("GetTenantConnection - Houve um erro ao conectar ao tenant: " + tenant + " Erro: " + err.Error())
        return nil, nil, err
    }

    dbSqlc := db.New(tenantConn.DB)

    return dbSqlc, tenantConn, nil // ← Retorna a conexão para Close()
}

func GetCatalogConnection() (*db.Queries, error) {
    dbCon := connection.GetCatalogConnection(config.GetCatalogDBConnectionString())

    if dbCon == nil {
        logger.Error.Println("::::::GetCatalogConnection - Ocorreu um erro ao conectar ao banco de dados catalog::::::")
        utils.SendDiscordMessage("GetCatalogConnection - Houve um erro ao conectar ao banco de dados catalog")
        return nil, errors.New("failed to connect to catalog database")
    }

    dbSqlc := db.New(dbCon)

    return dbSqlc, nil
}
```

---

## 📝 **Como atualizar seu código existente:**

### **ANTES:**

```go
queries, err := dbutils.GetTenantConnection(tenant)
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}

customerExists, err := services.GetCustomerByID(customerId.String())
```

### **DEPOIS:**

```go
queries, dbConn, err := dbutils.GetTenantConnection(tenant)
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}
defer dbConn.Close() // ← NOVO: Controle manual da conexão

customerExists, err := services.GetCustomerByID(customerId.String())
```

---

## 🎯 **Benefícios da nova assinatura:**

### ✅ **Controle Manual:**

```go
func YourHandler(tenant string, customerId uuid.UUID) error {
    queries, dbConn, err := dbutils.GetTenantConnection(tenant)
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, err)
    }
    defer dbConn.Close() // ← Fecha quando sair da função

    // Usar queries normalmente...
    customer, err := queries.GetCustomerByID(ctx, customerId)
    return nil
}
```

### ✅ **Health Check:**

```go
func YourHealthCheck(tenant string) error {
    queries, dbConn, err := dbutils.GetTenantConnection(tenant)
    if err != nil {
        return err
    }
    defer dbConn.Close()

    // Verifica se a conexão está saudável
    if !dbConn.IsHealthy(context.Background()) {
        return errors.New("database connection unhealthy")
    }

    return nil
}
```

### ✅ **Métricas de Conexão:**

```go
func LogConnectionMetrics(tenant string) {
    queries, dbConn, err := dbutils.GetTenantConnection(tenant)
    if err != nil {
        return
    }
    defer dbConn.Close()

    logger.Info.Printf("Connection age for tenant %s: %v", tenant, dbConn.GetAge())
}
```

### ✅ **Transações com Controle:**

```go
func YourTransactionHandler(tenant string) error {
    queries, dbConn, err := dbutils.GetTenantConnection(tenant)
    if err != nil {
        return err
    }
    defer dbConn.Close()

    // Começar transação
    tx, err := dbConn.DB.BeginTx(context.Background(), nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Usar transação...
    queriesWithTx := queries.WithTx(tx)

    // Se tudo ok, commit
    return tx.Commit()
}
```

---

## 🔄 **Padrão de uso recomendado:**

```go
// Padrão consistente em todos os handlers:
func AnyHandler() error {
    queries, dbConn, err := dbutils.GetTenantConnection("my-tenant")
    if err != nil {
        return err
    }
    defer dbConn.Close() // ← Sempre usar defer

    // Lógica de negócio aqui...
    return nil
}
```

---

## ⚠️ **Mudanças necessárias no seu código:**

1. **Adicionar `dbConn` em todas as chamadas**
2. **Adicionar `defer dbConn.Close()`**
3. **Atualizar handlers/services que usam GetTenantConnection**

### **Exemplo de refatoração:**

**Arquivo de exemplo - handler.go:**

```go
// ANTES:
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
    queries, err := dbutils.GetTenantConnection(tenant)
    // ...
}

// DEPOIS:
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
    queries, dbConn, err := dbutils.GetTenantConnection(tenant)
    if err != nil {
        // handle error
    }
    defer dbConn.Close() // ← ADICIONAR
    // ...
}
```

---

## 🎯 **RESULTADO:**

✅ **Controle total** sobre o ciclo de vida das conexões  
✅ **Performance melhorada** com cache e pool configurável  
✅ **Health checks** automáticos  
✅ **Métricas** de conexão disponíveis  
✅ **Cleanup automático** com `defer dbConn.Close()`

**A mudança é simples: só adicionar `dbConn` na assinatura e `defer dbConn.Close()` onde usar!** 🚀
