package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/MK-Solutions-LTDA/common-utils/cache"
	_ "github.com/lib/pq"
)

const prefixConnection = "con-"

type Connection struct {
	DB         *sql.DB
	SearchPath string
}

func GetTenantConnection(tenant string) (Connection, error) {
	cache.Mutex.Lock()
	defer cache.Mutex.Unlock()

	// Verifica se já existe uma conexão no cache para o tenant
	if conn, found := cache.Connections.Get(prefixConnection + tenant); found {
		log.Println("Found connection in cache for tenant ", tenant)
		return conn.(Connection), nil
	}

	catalog, err := GetTenant(context.Background(), tenant)
	if err != nil {
		return Connection{}, err
	}

	uri := fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=disable", catalog.Driver, catalog.UserName, catalog.Password, catalog.Server, catalog.DatabaseName)
	dbCon, err := sql.Open("postgres", uri)
	if err != nil {
		return Connection{}, err
	}

	log.Println("Connection create for tenant ", tenant)
	// Configura o search_path para usar o tenant
	_, err = dbCon.Exec(fmt.Sprintf("SET search_path TO %s", tenant))
	if err != nil {
		return Connection{}, err
	}

	// Salva a conexão no cache
	connection := Connection{DB: dbCon, SearchPath: tenant}
	cache.Connections.SetWithTTL(prefixConnection+tenant, connection, 1, 1*time.Hour)

	return connection, nil
}
