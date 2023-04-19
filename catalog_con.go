package connection

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"
)

var ErrRecordNotFound = errors.New("record not found")

type Catalog struct {
	Driver       string
	UserName     string
	Password     string
	Server       string
	DatabaseName string
	SchemaName   string
}

var (
	dbCatalog *sql.DB
	once      sync.Once
)

func Connect(url string) {
	var err error
	log.Println("vai conectar em: ", url)
	dbCatalog, err = sql.Open("postgres", url+"?sslmode=disable")
	if err != nil {
		log.Println("erro aqui?")
		panic(err)
	}
}

func GetCatalogConnection(url string) *sql.DB {
	once.Do(func() { Connect(url) })
	return dbCatalog
}

func GetTenant(ctx context.Context, tenant string) (*Catalog, error) {
	query := `
        SELECT driver, user_name, password, server, database_name, schema_name
        FROM catalog
		WHERE schema_name = $1
        LIMIT 1`

	var catalog Catalog

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err := dbCatalog.QueryRowContext(ctx, query, tenant).Scan(
		&catalog.Driver,
		&catalog.UserName,
		&catalog.Password,
		&catalog.Server,
		&catalog.DatabaseName,
		&catalog.SchemaName,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &catalog, nil
}