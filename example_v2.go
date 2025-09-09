package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Exemplo de uso da nova versão v2 da conexão com tenant

func ExampleBasicUsage() {
	ctx := context.Background()

	// Exemplo 1: Conexão básica v2
	basicExample(ctx)

	// Exemplo 2: Conexão com configurações avançadas
	advancedExample(ctx)

	// Exemplo 3: Usando com factory do sqlc
	sqlcExample(ctx)

	// Exemplo 4: Gerenciamento manual de conexões
	manualManagementExample(ctx)
}

func basicExample(ctx context.Context) {
	fmt.Println("=== Exemplo 1: Conexão básica v2 ===")

	// Configuração básica - usa valores padrão
	opts := TenantConnectOptions{
		Tenant: "meu_tenant",
	}

	tenantConn, err := GetTenantConnectionV2(ctx, opts)
	if err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}
	defer tenantConn.Close() // Importante: sempre feche a conexão

	// Verifica se a conexão está saudável
	if tenantConn.IsHealthy(ctx) {
		fmt.Printf("Conexão criada com sucesso para tenant: %s\n", tenantConn.SearchPath)
		fmt.Printf("Idade da conexão: %v\n", tenantConn.GetAge())
	}

	// Executa uma query com log
	result, err := tenantConn.ExecWithLog(ctx, "SELECT 1")
	if err != nil {
		log.Printf("Erro na query: %v", err)
		return
	}

	fmt.Printf("Query executada com sucesso: %v\n", result)
}

func advancedExample(ctx context.Context) {
	fmt.Println("\n=== Exemplo 2: Conexão com configurações avançadas ===")

	// Logger customizado
	customLogger := func(ctx context.Context, query string, args ...any) {
		fmt.Printf("[CUSTOM_LOG] Tenant Query: %s with args: %v\n", query, args)
	}

	// Configuração avançada
	opts := TenantConnectOptions{
		Tenant:          "tenant_avancado",
		MaxOpenConns:    50,
		MaxIdleConns:    25,
		ConnMaxIdle:     30 * time.Minute,
		ConnMaxLifetime: 2 * time.Hour,
		ForceUTC:        true,
		QueryLogger:     customLogger,
		CacheEnabled:    true,
		CacheTTL:        1 * time.Hour,
	}

	tenantConn, err := GetTenantConnectionV2(ctx, opts)
	if err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}
	defer tenantConn.Close()

	fmt.Printf("Conexão avançada criada para tenant: %s\n", tenantConn.SearchPath)

	// Executa query com logging customizado
	rows, err := tenantConn.QueryWithLog(ctx, "SELECT version()")
	if err != nil {
		log.Printf("Erro na query: %v", err)
		return
	}
	defer rows.Close()

	var version string
	if rows.Next() {
		rows.Scan(&version)
		fmt.Printf("Versão do PostgreSQL: %s\n", version)
	}
}

// Exemplo de estrutura para usar com sqlc (mockado)
type MockQueries struct {
	db *sql.DB
}

func NewMockQueries(db *sql.DB) *MockQueries {
	return &MockQueries{db: db}
}

func (q *MockQueries) GetUser(ctx context.Context, id int) error {
	// Simulação de uma query do sqlc
	row := q.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", id)
	var name string
	return row.Scan(&name)
}

func sqlcExample(ctx context.Context) {
	fmt.Println("\n=== Exemplo 3: Usando com factory do sqlc ===")

	opts := TenantConnectOptions{
		Tenant:       "tenant_sqlc",
		QueryLogger:  DefaultTenantLogger,
		CacheEnabled: true,
	}

	// Factory para criar a instância do "sqlc"
	factory := func(db *sql.DB) *MockQueries {
		return NewMockQueries(db)
	}

	queries, tenantConn, err := NewSqlcWithTenantConnection(ctx, factory, opts)
	if err != nil {
		log.Printf("Erro ao criar conexão com sqlc: %v", err)
		return
	}
	defer tenantConn.Close()

	fmt.Printf("Sqlc criado com sucesso para tenant: %s\n", tenantConn.SearchPath)

	// Usar o sqlc normalmente
	err = queries.GetUser(ctx, 123)
	if err != nil {
		log.Printf("Erro ao buscar usuário: %v", err)
	}
}

func manualManagementExample(ctx context.Context) {
	fmt.Println("\n=== Exemplo 4: Gerenciamento manual de conexões ===")

	opts := TenantConnectOptions{
		Tenant:       "tenant_manual",
		CacheEnabled: false, // Desabilita cache para controle manual
	}

	// Criar múltiplas conexões
	connections := make([]*TenantConnectionV2, 0, 3)

	for i := 0; i < 3; i++ {
		tenantConn, err := GetTenantConnectionV2(ctx, opts)
		if err != nil {
			log.Printf("Erro ao criar conexão %d: %v", i, err)
			continue
		}
		connections = append(connections, tenantConn)
		fmt.Printf("Conexão %d criada (idade: %v)\n", i+1, tenantConn.GetAge())
	}

	// Verifica saúde das conexões
	for i, conn := range connections {
		if conn.IsHealthy(ctx) {
			fmt.Printf("Conexão %d está saudável\n", i+1)
		} else {
			fmt.Printf("Conexão %d não está saudável\n", i+1)
		}
	}

	// Fecha todas as conexões
	for i, conn := range connections {
		if err := conn.Close(); err != nil {
			log.Printf("Erro ao fechar conexão %d: %v", i+1, err)
		} else {
			fmt.Printf("Conexão %d fechada com sucesso\n", i+1)
		}
	}
}
