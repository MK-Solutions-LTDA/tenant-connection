# 🎯 Migração do seu dbutils - Passo a Passo

## 📋 **RECOMENDAÇÃO: Migração Gradual e Segura**

### **PASSO 1: Atualizar função existente (mantém compatibilidade)**

No seu `dbutils/dbutils.go`, faça apenas estas mudanças:

```go
package dbutils

import (
    "account-payable-service/internal/config"
    "account-payable-service/internal/db"
    "account-payable-service/internal/logger"
    "account-payable-service/internal/utils"
    "context"  // ← ADICIONAR APENAS ISSO
    "errors"

    connection "github.com/MK-Solutions-LTDA/tenant-connection"
)

// Mesma assinatura - ZERO mudanças no resto do código
func GetTenantConnection(tenant string) (*db.Queries, error) {
    ctx := context.Background()

    // Troca só essa parte ↓
    opts := connection.TenantConnectOptions{
        Tenant:       tenant,
        CacheEnabled: true,  // Cache melhorado da v2
        MaxOpenConns: 25,    // Pool configurável
        ForceUTC:     true,  // Força UTC para consistência
    }

    tenantConn, err := connection.GetTenantConnectionV2(ctx, opts)
    // ↑ Troca só essa parte

    if err != nil {
        logger.Error.Println("::::::GetTenantConnection - Ocorreu um erro ao conectar ao tenant::::::", tenant, err)
        utils.SendDiscordMessage("GetTenantConnection - Houve um erro ao conectar ao tenant: " + tenant + " Erro: " + err.Error())
        return nil, err
    }

    dbSqlc := db.New(tenantConn.DB) // Mesmo código

    return dbSqlc, nil
}

// Função de catálogo permanece igual
func GetCatalogConnection() (*db.Queries, error) {
    // ... código atual sem mudanças
}
```

**✅ RESULTADO:**

- Todo seu código atual continua funcionando igual
- Ganha cache melhorado, pool de conexões, etc.
- Zero mudanças em services, handlers, etc.

### **PASSO 2: (Opcional) Adicionar versão v2 para novos casos**

```go
// Adicionar nova função para casos que precisam de controle manual
func GetTenantConnectionV2(tenant string) (*db.Queries, *connection.TenantConnectionV2, error) {
    ctx := context.Background()

    opts := connection.TenantConnectOptions{
        Tenant:       tenant,
        CacheEnabled: true,
        MaxOpenConns: 25,
        ForceUTC:     true,
    }

    tenantConn, err := connection.GetTenantConnectionV2(ctx, opts)
    if err != nil {
        logger.Error.Println("::::::GetTenantConnectionV2 - Erro ao conectar ao tenant::::::", tenant, err)
        utils.SendDiscordMessage("GetTenantConnectionV2 - Erro ao conectar ao tenant: " + tenant + " Erro: " + err.Error())
        return nil, nil, err
    }

    dbSqlc := db.New(tenantConn.DB)

    return dbSqlc, tenantConn, nil
}
```

### **PASSO 3: (Futuro) Factory Global para casos novos**

Se quiser ainda mais simplicidade para códigos novos:

```go
// No main.go ou init()
func init() {
    dbutils.SetupGlobalFactory()
}

// No dbutils.go
func SetupGlobalFactory() {
    factory := func(dbConn *sql.DB) *db.Queries {
        return db.New(dbConn)
    }

    connection.SetGlobalFactory(factory)
    logger.Info.Println("Global factory configurado para dbutils")
}

// Nova função super simples para códigos futuros
func GetConnection(tenant string) (*db.Queries, *connection.TenantConnectionV2, error) {
    ctx := context.Background()

    queries, tenantConn, err := connection.GetConnectionSimple[*db.Queries](ctx, tenant)
    if err != nil {
        logger.Error.Println("::::::GetConnection - Erro::::::", tenant, err)
        utils.SendDiscordMessage("GetConnection - Erro: " + tenant + " - " + err.Error())
        return nil, nil, err
    }

    return queries, tenantConn, nil
}
```

## 📝 **Como fica seu código atual:**

### **ANTES:**

```go
queries, err := dbutils.GetTenantConnection("tenant123")
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}

customerExists, err := services.GetCustomerByID(customerId.String())
```

### **DEPOIS (PASSO 1 - zero mudanças):**

```go
queries, err := dbutils.GetTenantConnection("tenant123") // ← IGUAL!
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}

customerExists, err := services.GetCustomerByID(customerId.String())
// Mas agora tem cache melhorado, pool configurável, etc. automaticamente!
```

### **DEPOIS (PASSO 2 - controle manual opcional):**

```go
queries, dbConn, err := dbutils.GetTenantConnectionV2("tenant123")
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}
defer dbConn.Close() // ← Controle manual

customerExists, err := services.GetCustomerByID(customerId.String())
```

### **DEPOIS (PASSO 3 - super simples para códigos novos):**

```go
queries, dbConn, err := dbutils.GetConnection("tenant123")
if err != nil {
    return utils.NewApiError(http.StatusInternalServerError, err)
}
defer dbConn.Close()

customerExists, err := services.GetCustomerByID(customerId.String())
```

## 🎯 **RESUMO:**

1. **PASSO 1**: Atualize só o `GetTenantConnection` - zero impacto no resto
2. **PASSO 2**: Adicione `GetTenantConnectionV2` para casos novos
3. **PASSO 3**: Configure factory global para máxima simplicidade

**Começe pelo PASSO 1 - é 100% seguro e já dá todos os benefícios!** ✨
