# 🚀 Como usar SEM repetir o factory

## ✅ **Resposta: NÃO, não precisa repetir toda vez!**

### 📋 **Configuração inicial (fazer UMA VEZ só):**

```go
// No seu main.go ou init():
func main() {
    // Configure o factory global UMA VEZ no início
    factory := func(db *sql.DB) *YourQueries {
        return db.NewYourQueries(db) // Seu SQLC gerado
    }

    connection.SetGlobalFactory(factory)

    // Resto da sua aplicação...
}
```

### 🎯 **Agora pode usar de forma MUITO mais simples:**

## **ANTES (repetindo factory):**

```go
factory := func(db *sql.DB) *Queries {
    return NewQueries(db)
}
queries, dbConn, err := connection.GetConnection(ctx, tenant, factory) // ❌ Repetitivo
```

## **DEPOIS (factory global):**

```go
// Super simples - só precisa do tenant:
queries, dbConn, err := connection.GetConnectionSimple[*Queries](ctx, tenant) // ✅ Clean!
```

## **OU AINDA MAIS SIMPLES (tenant no contexto):**

```go
// Ultra simples - nem tenant precisa passar:
queries, dbConn, err := connection.GetConnectionFromContextSimple[*Queries](ctx) // ✅ Perfect!
```

## 🔧 **Opções de uso:**

### **1. Simples (só tenant):**

```go
func YourHandler(tenant string, customerId string) error {
    queries, dbConn, err := connection.GetConnectionSimple[*Queries](ctx, tenant)
    if err != nil {
        return err
    }
    defer dbConn.Close()

    customer, err := queries.GetCustomerByID(ctx, customerId)
    // ...
}
```

### **2. Ultra simples (tenant do contexto):**

```go
func YourAPIHandler(w http.ResponseWriter, r *http.Request) {
    // Tenant já no contexto (middleware)
    queries, dbConn, err := connection.GetConnectionFromContextSimple[*Queries](r.Context())
    if err != nil {
        http.Error(w, "DB error", 500)
        return
    }
    defer dbConn.Close()

    customer, err := queries.GetCustomerByID(r.Context(), "123")
    // ...
}
```

### **3. Original (quando precisar de factory específico):**

```go
func SpecialCase(tenant string) error {
    // Factory customizado para casos especiais
    specialFactory := func(db *sql.DB) *SpecialQueries {
        return NewSpecialQueries(db)
    }

    queries, dbConn, err := connection.GetConnection(ctx, tenant, specialFactory)
    // ...
}
```

## 🎯 **Resumo das opções:**

| Função                           | Factory   | Tenant      | Quando usar     |
| -------------------------------- | --------- | ----------- | --------------- |
| `GetConnection`                  | ✅ Manual | ✅ Manual   | Casos especiais |
| `GetConnectionSimple`            | 🔄 Global | ✅ Manual   | Uso normal      |
| `GetConnectionFromContextSimple` | 🔄 Global | 🔄 Contexto | APIs/handlers   |

## 💡 **Recomendação:**

1. **Configure o factory global** no início da aplicação
2. **Use `GetConnectionSimple`** para casos normais
3. **Use `GetConnectionFromContextSimple`** para APIs
4. **Use `GetConnection`** só para casos especiais

**Resultado: Código muito mais limpo e sem repetição!** 🎉

## 📝 **Seu código fica assim:**

```go
// Era assim (repetitivo):
factory := func(db *sql.DB) *Queries { return NewQueries(db) }
queries, dbConn, err := GetConnection(ctx, tenant, factory)

// Agora fica assim (limpo):
queries, dbConn, err := GetConnectionSimple[*Queries](ctx, tenant)

// Ou ainda mais simples:
queries, dbConn, err := GetConnectionFromContextSimple[*Queries](ctx)
```
