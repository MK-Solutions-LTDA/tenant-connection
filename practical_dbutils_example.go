package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// ===== EXEMPLO PRÁTICO: Como atualizar seu dbutils =====

// Simula suas estruturas (substitua pelos imports reais)
type DbQueries struct {
	db *sql.DB
}

func NewDbQueries(db *sql.DB) *DbQueries {
	return &DbQueries{db: db}
}

func (q *DbQueries) GetCustomerByID(ctx context.Context, id string) error {
	// Simula uma query
	_, err := q.db.QueryContext(ctx, "SELECT * FROM customers WHERE id = $1", id)
	return err
}

// ===== VERSÃO ATUALIZADA DO SEU DBUTILS =====

// Nova assinatura com Close() - igual ao que você quer
func GetTenantConnectionUpdated(tenant string) (*DbQueries, *TenantConnectionV2, error) {
	ctx := context.Background()

	opts := TenantConnectOptions{
		Tenant:       tenant,
		CacheEnabled: true,
		MaxOpenConns: 25,
		MaxIdleConns: 25,
		ForceUTC:     true,
	}

	tenantConn, err := GetTenantConnectionV2(ctx, opts)
	if err != nil {
		log.Printf("::::::GetTenantConnection - Erro ao conectar ao tenant::::::: %s %v", tenant, err)
		// utils.SendDiscordMessage("GetTenantConnection - Erro: " + tenant + " - " + err.Error())
		return nil, nil, err
	}

	dbSqlc := NewDbQueries(tenantConn.DB)

	return dbSqlc, tenantConn, nil // ← Retorna conexão para Close()
}

// ===== EXEMPLOS DE USO =====

// Exemplo 1: Handler básico
func ExampleHandler(tenant string, customerId string) error {
	// Nova assinatura: agora retorna dbConn também
	queries, dbConn, err := GetTenantConnectionUpdated(tenant)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	defer dbConn.Close() // ← IMPORTANTE: sempre fechar

	// Usar queries normalmente
	err = queries.GetCustomerByID(context.Background(), customerId)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	log.Printf("Customer %s found for tenant %s", customerId, tenant)
	return nil
}

// Exemplo 2: Com health check
func ExampleWithHealthCheck(tenant string) error {
	_, dbConn, err := GetTenantConnectionUpdated(tenant)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	// Verifica saúde da conexão
	if !dbConn.IsHealthy(context.Background()) {
		return fmt.Errorf("database connection unhealthy for tenant: %s", tenant)
	}

	log.Printf("Connection healthy for tenant %s (age: %v)", tenant, dbConn.GetAge())
	return nil
}

// Exemplo 3: Multiple operations
func ExampleMultipleOperations(tenant string) error {
	queries, dbConn, err := GetTenantConnectionUpdated(tenant)
	if err != nil {
		return err
	}
	defer dbConn.Close() // ← Uma vez só, no final fecha tudo

	ctx := context.Background()

	// Múltiplas operações com a mesma conexão
	err = queries.GetCustomerByID(ctx, "customer1")
	if err != nil {
		return err
	}

	err = queries.GetCustomerByID(ctx, "customer2")
	if err != nil {
		return err
	}

	err = queries.GetCustomerByID(ctx, "customer3")
	if err != nil {
		return err
	}

	log.Printf("Multiple operations completed for tenant: %s", tenant)
	return nil
}

// Exemplo 4: Com transação
func ExampleWithTransaction(tenant string) error {
	_, dbConn, err := GetTenantConnectionUpdated(tenant)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	ctx := context.Background()

	// Iniciar transação
	tx, err := dbConn.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback se não committar

	// Usar transação
	// queriesWithTx := queries.WithTx(tx) // Se seu SQLC suportar

	// Operações da transação...
	_, err = tx.ExecContext(ctx, "INSERT INTO customers (id, name) VALUES ($1, $2)", "123", "Test")
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	// Se tudo ok, commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Transaction completed for tenant: %s", tenant)
	return nil
}

// ===== COMPARAÇÃO: ANTES vs DEPOIS =====

func ShowBeforeAfter() {
	tenant := "example_tenant"
	customerId := "customer123"

	// ANTES (assinatura antiga):
	/*
		queries, err := GetTenantConnection(tenant)
		if err != nil {
			return err
		}
		customer, err := queries.GetCustomerByID(context.Background(), customerId)
	*/

	// DEPOIS (nova assinatura com Close):
	queries, dbConn, err := GetTenantConnectionUpdated(tenant)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer dbConn.Close() // ← SÓ ISSO DE DIFERENÇA!

	err = queries.GetCustomerByID(context.Background(), customerId)
	if err != nil {
		log.Printf("Customer error: %v", err)
		return
	}

	log.Printf("Success! Connection age: %v", dbConn.GetAge())
}

// ===== FACTORY GLOBAL (ALTERNATIVA AINDA MAIS SIMPLES) =====

func SetupFactoryExample() {
	// Configure uma vez no início da aplicação
	factory := func(db *sql.DB) *DbQueries {
		return NewDbQueries(db)
	}

	SetGlobalFactory(factory)
	log.Println("Factory global configurado!")
}

func ExampleWithGlobalFactory(tenant string) error {
	// Super simples com factory global
	queries, dbConn, err := GetConnectionSimple[*DbQueries](context.Background(), tenant)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	err = queries.GetCustomerByID(context.Background(), "123")
	if err != nil {
		return err
	}

	log.Printf("Global factory example completed for tenant: %s", tenant)
	return nil
}
