# ✅ ATUALIZAÇÃO FINAL - Nova Assinatura com Close()

## 🎯 **Exatamente o que você pediu:**

### **ANTES (sua assinatura atual):**

```go
func GetTenantConnection(tenant string) (*db.Queries, error)
```

### **DEPOIS (nova assinatura V2 com Close):**

```go
func GetTenantConnection(tenant string) (*db.Queries, *connection.TenantConnectionV2, error)
```

---

## 💻 **Código completo do seu dbutils atualizado:**

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

// ✅ NOVA ASSINATURA: Retorna conexão para Close()
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

    return dbSqlc, tenantConn, nil // ← MUDANÇA: Retorna a conexão
}

// Função de catálogo permanece igual
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

### **EM CADA LUGAR QUE USA:**

**ANTES:**

```go
queries, err := dbutils.GetTenantConnection(tenant)
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}

customerExists, err := services.GetCustomerByID(customerId.String())
```

**DEPOIS:**

```go
queries, dbConn, err := dbutils.GetTenantConnection(tenant)
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}
defer dbConn.Close() // ← ADICIONAR SEMPRE

customerExists, err := services.GetCustomerByID(customerId.String())
```

---

## 🔍 **Exemplo prático em um handler:**

```go
func YourHandler(tenant string, customerId uuid.UUID) error {
    // Nova assinatura com dbConn
    queries, dbConn, err := dbutils.GetTenantConnection(tenant)
    if err != nil {
        return utils.NewApiError(http.StatusInternalServerError, err)
    }
    defer dbConn.Close() // ← SEMPRE adicionar

    // Código normal (sem mudanças)
    customerExists, err := queries.GetCustomerByID(ctx, customerId.String())
    if err != nil {
        return utils.NewApiError(http.StatusNotFound, err)
    }

    // Bonus: Agora você pode fazer health check
    if !dbConn.IsHealthy(ctx) {
        logger.Warn.Println("Database connection is unhealthy")
    }

    // Bonus: Ver métricas de conexão
    logger.Info.Printf("Connection age: %v", dbConn.GetAge())

    return nil
}
```

---

## ⚡ **Benefícios que você ganha:**

### ✅ **Controle Manual:**

```go
defer dbConn.Close() // Você controla quando fecha
```

### ✅ **Health Checks:**

```go
if !dbConn.IsHealthy(ctx) {
    // Conexão não está saudável
}
```

### ✅ **Métricas:**

```go
logger.Info.Printf("Connection age: %v", dbConn.GetAge())
```

### ✅ **Cache Melhorado:**

- Pool de conexões configurável
- Cache automático por tenant
- Reuso eficiente de conexões

### ✅ **Transações:**

```go
tx, err := dbConn.DB.BeginTx(ctx, nil)
defer tx.Rollback()
// usar transação...
tx.Commit()
```

---

## 🎯 **RESUMO DO QUE MUDAR:**

1. **No dbutils.go**: Atualizar a assinatura da função
2. **Em cada uso**: Adicionar `dbConn` na atribuição
3. **Em cada uso**: Adicionar `defer dbConn.Close()`

### **Mudança mínima necessária:**

```go
// Era assim:
queries, err := dbutils.GetTenantConnection(tenant)

// Agora é assim:
queries, dbConn, err := dbutils.GetTenantConnection(tenant)
defer dbConn.Close()
```

**E pronto! Você tem controle total das conexões com todos os benefícios da v2!** 🚀

## 💡 **Dica:**

Use Find & Replace no seu IDE:

- **Buscar**: `queries, err := dbutils.GetTenantConnection(`
- **Substituir**: `queries, dbConn, err := dbutils.GetTenantConnection(`
- Depois adicione `defer dbConn.Close()` em cada local
