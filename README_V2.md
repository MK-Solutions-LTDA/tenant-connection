# Tenant Connection V2

Esta é a versão v2 da biblioteca de conexão com PostgreSQL com suporte a multi-tenant. A v2 mantém compatibilidade com a versão original mas adiciona novos recursos avançados.

## Novos Recursos na V2

- ✅ **Gerenciamento avançado de conexões**: Configuração detalhada de pool de conexões
- ✅ **Close manual**: Método `Close()` para fechar conexões explicitamente
- ✅ **Health checks**: Método `IsHealthy()` para verificar saúde da conexão
- ✅ **Query logging**: Sistema de logging customizável para queries
- ✅ **Timezone UTC forçado**: Opção para forçar timezone UTC
- ✅ **Controle de cache**: Opção para habilitar/desabilitar cache de conexões
- ✅ **Métricas de conexão**: Idade da conexão e outras métricas
- ✅ **Factory para SQLC**: Integração facilitada com SQLC
- ✅ **Context support**: Suporte completo a context.Context

## Uso Básico

```go
package main

import (
    "context"
    "log"

    connection "github.com/MK-Solutions-LTDA/tenant-connection"
)

func main() {
    ctx := context.Background()

    // Configuração básica
    opts := connection.TenantConnectOptions{
        Tenant: "meu_tenant",
    }

    // Obtém conexão
    tenantConn, err := connection.GetTenantConnectionV2(ctx, opts)
    if err != nil {
        log.Fatal(err)
    }
    defer tenantConn.Close() // Importante: sempre feche a conexão

    // Usa a conexão
    result, err := tenantConn.ExecWithLog(ctx, "SELECT 1")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Query executada: %v", result)
}
```

## Configuração Avançada

```go
// Logger customizado
customLogger := func(ctx context.Context, query string, args ...any) {
    log.Printf("[CUSTOM] Query: %s | Args: %v", query, args)
}

opts := connection.TenantConnectOptions{
    Tenant:          "meu_tenant",
    MaxOpenConns:    50,                // Máximo de conexões abertas
    MaxIdleConns:    25,                // Máximo de conexões idle
    ConnMaxIdle:     30 * time.Minute,  // Tempo máximo idle
    ConnMaxLifetime: 2 * time.Hour,     // Tempo máximo de vida
    ForceUTC:        true,              // Força timezone UTC
    QueryLogger:     customLogger,      // Logger customizado
    CacheEnabled:    true,              // Habilita cache
    CacheTTL:        1 * time.Hour,     // TTL do cache
}

tenantConn, err := connection.GetTenantConnectionV2(ctx, opts)
if err != nil {
    log.Fatal(err)
}
defer tenantConn.Close()
```

## Integração com SQLC

```go
// Assuming you have generated SQLC code
type Queries struct {
    db *sql.DB
}

func NewQueries(db *sql.DB) *Queries {
    return &Queries{db: db}
}

// Factory function
factory := func(db *sql.DB) *Queries {
    return NewQueries(db)
}

opts := connection.TenantConnectOptions{
    Tenant: "meu_tenant",
}

// Cria queries e conexão em uma chamada
queries, tenantConn, err := connection.NewSqlcWithTenantConnection(ctx, factory, opts)
if err != nil {
    log.Fatal(err)
}
defer tenantConn.Close()

// Usa o SQLC normalmente
user, err := queries.GetUser(ctx, 123)
if err != nil {
    log.Fatal(err)
}
```

## Métodos Disponíveis

### TenantConnectionV2

- `Close() error` - Fecha a conexão e remove do cache
- `IsHealthy(ctx context.Context) bool` - Verifica se a conexão está saudável
- `GetAge() time.Duration` - Retorna a idade da conexão
- `ExecWithLog(ctx, query, args...) (sql.Result, error)` - Executa query com log
- `QueryWithLog(ctx, query, args...) (*sql.Rows, error)` - Executa query com log
- `QueryRowWithLog(ctx, query, args...) *sql.Row` - Executa query row com log

### Funções Globais

- `GetTenantConnectionV2(ctx, opts) (*TenantConnectionV2, error)` - Obtém conexão v2
- `NewSqlcWithTenantConnection[T](ctx, factory, opts) (T, *TenantConnectionV2, error)` - Cria SQLC com conexão
- `CloseAllTenantConnections() error` - Fecha todas as conexões (placeholder)

## Configurações Padrão

| Opção             | Valor Padrão        | Descrição                  |
| ----------------- | ------------------- | -------------------------- |
| `MaxOpenConns`    | 25                  | Máximo de conexões abertas |
| `MaxIdleConns`    | 25                  | Máximo de conexões idle    |
| `ConnMaxIdle`     | 1 hora              | Tempo máximo idle          |
| `ConnMaxLifetime` | 1 hora              | Tempo máximo de vida       |
| `CacheEnabled`    | true                | Cache habilitado           |
| `CacheTTL`        | 55 minutos          | TTL do cache               |
| `ForceUTC`        | false               | Não força UTC              |
| `QueryLogger`     | DefaultTenantLogger | Logger padrão              |

## Compatibilidade

A versão v2 é completamente compatível com a versão original. Você pode usar ambas no mesmo projeto:

```go
// Versão original (ainda funciona)
conn, err := connection.GetTenantConnection("meu_tenant")

// Versão v2 (novos recursos)
connV2, err := connection.GetTenantConnectionV2(ctx, opts)
```

## Cache e Performance

- O cache é **habilitado por padrão** na v2
- Conexões são automaticamente reutilizadas entre chamadas
- TTL padrão de 55 minutos (configurável)
- Conexões inválidas são automaticamente removidas do cache
- Use `CacheEnabled: false` para desabilitar o cache quando necessário

## Logging

A v2 inclui sistema de logging robusto:

```go
// Logger padrão
connection.DefaultTenantLogger(ctx, "SELECT * FROM users", 123)
// Output: [TenantQuery] SELECT * FROM users | args: [123]

// Logger customizado
myLogger := func(ctx context.Context, query string, args ...any) {
    // Seu código de log aqui
}
```

## Migrações (Futuro)

Para adicionar suporte a migrações, você pode adicionar as dependências necessárias:

```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
```

Em seguida, descomente a função `MigrateTenantDatabase` no arquivo `tenant_con_v2.go`.

## Exemplo Completo

Veja o arquivo `example_v2.go` para exemplos completos de uso da biblioteca.
