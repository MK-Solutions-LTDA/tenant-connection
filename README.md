# Tenant Connection

Este repositório contém código para gerenciar a conexão de tenants em um 
aplicativo em Go. Ele fornece uma maneira fácil de estabelecer uma conexão 
segura com um banco de dados de catálogo e encapsular a conexão para facilitar 
o acesso aos dados do tenant.

## Como usar

Para iniciar o aplicativo em Go, é necessário chamar a conexão com o catálogo, 
passando a variável de ambiente com o endereço do banco de dados. 
Por exemplo:

> Arquivo main.go

```go
func main() {
	catalogDB := connection.GetCatalogConnection(os.Getenv("CATALOG_URL"))
	defer catalogDB.Close()

	// Resto do código...
}
```


Também é necessário "encapsular" a conexão de tenant da seguinte forma:

> Arquivo tenant.go

```go
package dbutils

import (
	"customer-portal.mknext.net/internal/db"

	connection "github.com/MK-Solutions-LTDA/tenant-connection"
)

func GetTenantConnection(tenant string) (*db.Queries, error) {
	dbCon, err := connection.GetTenantConnection(tenant)
	if err != nil {
		panic(err)
	}

	dbSqlc := db.New(dbCon.DB)

	return dbSqlc, nil
}
```

Para chamar a funcao `GetTenantConnection` basta passar o tenant <string>
como parâmetro

```go
db, err := dbutils.GetTenantConnection(tenant)
```
