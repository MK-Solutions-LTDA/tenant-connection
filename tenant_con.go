package connection

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const prefixConnection = "con-"

type Connection struct {
	DB         *sql.DB
	SearchPath string
}

func GetTenantConnection(tenant string) (Connection, error) {
	Mutex.Lock()
	defer Mutex.Unlock()

	// Verifica se já existe uma conexão no cache para o tenant
	if conn, found := Connections.Get(prefixConnection + tenant); found {
		return conn.(Connection), nil
	}

	catalog, err := GetTenant(tenant)
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
	Connections.SetWithTTL(prefixConnection+tenant, connection, 1, 1*time.Hour)
	connection.DB.SetConnMaxLifetime(1 * time.Hour)
	connection.DB.SetConnMaxIdleTime(1 * time.Hour)

	return connection, nil
}
