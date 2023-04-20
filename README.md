# Tenant Connection

Este repositório contém código para gerenciar a conexão de inquilinos em um 
aplicativo em Go. Ele fornece uma maneira fácil de estabelecer uma conexão 
segura com um banco de dados de catálogo e encapsular a conexão de inquilino 
para facilitar o acesso aos dados do inquilino.

## Como usar

Para iniciar o aplicativo em Go, é necessário chamar a conexão com o catálogo, 
passando a variável de ambiente com o endereço do banco de dados de catálogo. 
Por exemplo:

> Arquivo main.go

```go
func main() {
	catalogDB := connection.GetCatalogConnection(os.Getenv("CATALOG_URL"))
	defer catalogDB.Close()

	// Resto do código...
}
```

--

Também é necessário "encapsular" a conexão de inquilino da seguinte forma:

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

Para chamar a funcao `GetTenantConnection` basta fazer:

```go
db, err := dbutils.GetTenantConnection(tenant)
```