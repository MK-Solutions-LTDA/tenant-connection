# 🔧 Como atualizar seu dbutils para usar v2

Baseado no seu código atual, aqui estão as opções para migrar para v2:

## 📋 **Seu código atual:**

```go
package dbutils

import (
    "account-payable-service/internal/config"
    "account-payable-service/internal/db"
    "account-payable-service/internal/logger"
    "account-payable-service/internal/utils"
    "errors"

    connection "github.com/MK-Solutions-LTDA/tenant-connection"
)

func GetTenantConnection(tenant string) (*db.Queries, error) {
    dbCon, err := connection.GetTenantConnection(tenant)
    if err != nil {
        logger.Error.Println("::::::GetTenantConnection - Ocorreu um erro ao conectar ao tenant::::::", tenant, err)
        utils.SendDiscordMessage("GetTenantConnection - Houve um erro ao conectar ao tenant: " + tenant + " Erro: " + err.Error())
        return nil, err
    }

    dbSqlc := db.New(dbCon.DB)
    return dbSqlc, nil
}
```

---

## 🚀 **OPÇÃO 1: Migração Simples (Recomendada)**

Apenas adicione `context` e atualize internamente para v2:

```go
package dbutils

import (
    "account-payable-service/internal/config"
    "account-payable-service/internal/db"
    "account-payable-service/internal/logger"
    "account-payable-service/internal/utils"
    "context"  // ← ADICIONAR
    "errors"

    connection "github.com/MK-Solutions-LTDA/tenant-connection"
)

// Mantém a mesma assinatura (compatibilidade total)
func GetTenantConnection(tenant string) (*db.Queries, error) {
    ctx := context.Background()

    // Usa v2 internamente com benefícios melhorados
    opts := connection.TenantConnectOptions{
        Tenant:       tenant,
        CacheEnabled: true,        // Cache melhorado
        MaxOpenConns: 25,          // Pool configurável
        ForceUTC:     true,        // Força UTC
    }

    tenantConn, err := connection.GetTenantConnectionV2(ctx, opts)
    if err != nil {
        logger.Error.Println("::::::GetTenantConnection - Erro ao conectar ao tenant::::::", tenant, err)
        utils.SendDiscordMessage("GetTenantConnection - Erro ao conectar ao tenant: " + tenant + " Erro: " + err.Error())
        return nil, err
    }

    dbSqlc := db.New(tenantConn.DB)
    return dbSqlc, nil
}

// Resto das funções igual...
```

**✅ Vantagens:**

- Zero mudanças no resto do seu código
- Aproveita melhorias da v2 automaticamente
- Cache mais eficiente
- Pool de conexões configurável

---

## 🎯 **OPÇÃO 2: Nova Função V2 (Controle Manual)**

Adicione uma nova função que retorna a conexão para controle manual:

```go
// Nova função v2 (adicionar junto com a atual)
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
        logger.Error.Println("::::::GetTenantConnectionV2 - Erro::::::", tenant, err)
        utils.SendDiscordMessage("GetTenantConnectionV2 - Erro: " + tenant + " - " + err.Error())
        return nil, nil, err
    }

    dbSqlc := db.New(tenantConn.DB)
    return dbSqlc, tenantConn, nil // ← Retorna conexão para Close()
}
```

**Como usar:**

```go
// Onde você usava:
queries, err := dbutils.GetTenantConnection(tenant)

// Agora pode usar também:
queries, dbConn, err := dbutils.GetTenantConnectionV2(tenant)
defer dbConn.Close() // ← Controle manual
```

---

## ⚡ **OPÇÃO 3: Factory Global (Mais Elegante)**

Configure uma vez e use de forma super simples:

```go
// No seu main.go (configurar UMA VEZ):
func main() {
    dbutils.SetupGlobalFactory()
    // resto...
}

// No dbutils.go:
func SetupGlobalFactory() {
    factory := func(dbConn *sql.DB) *db.Queries {
        return db.New(dbConn)
    }

    connection.SetGlobalFactory(factory)
    logger.Info.Println("Global factory configurado")
}

// Nova função super simples:
func GetTenantConnectionSimple(tenant string) (*db.Queries, *connection.TenantConnectionV2, error) {
    ctx := context.Background()

    queries, tenantConn, err := connection.GetConnectionSimple[*db.Queries](ctx, tenant)
    if err != nil {
        logger.Error.Println("::::::GetTenantConnectionSimple - Erro::::::", tenant, err)
        utils.SendDiscordMessage("GetTenantConnectionSimple - Erro: " + tenant + " - " + err.Error())
        return nil, nil, err
    }

    return queries, tenantConn, nil
}
```

**Como usar:**

```go
queries, dbConn, err := dbutils.GetTenantConnectionSimple(tenant)
defer dbConn.Close()
```

---

## 🏆 **RECOMENDAÇÃO:**

**Use a OPÇÃO 1 primeiro** - é a mais segura e mantém compatibilidade total.

```go
// Só adicionar context e atualizar a implementação interna:
func GetTenantConnection(tenant string) (*db.Queries, error) {
    ctx := context.Background()
    opts := connection.TenantConnectOptions{
        Tenant:       tenant,
        CacheEnabled: true,
        MaxOpenConns: 25,
        ForceUTC:     true,
    }

    tenantConn, err := connection.GetTenantConnectionV2(ctx, opts)
    if err != nil {
        // seu log de erro atual
        return nil, err
    }

    return db.New(tenantConn.DB), nil
}
```

**Benefícios imediatos:**

- ✅ Cache melhorado
- ✅ Pool de conexões configurável
- ✅ Health checks automáticos
- ✅ Zero mudanças no resto do código
- ✅ Logs de performance opcionais

**Depois, gradualmente, pode adicionar as versões v2 para casos que precisam de controle manual!**
