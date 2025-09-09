package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// Exemplo prático de como usar igual ao seu outro projeto

// Suponha que você tenha SQLC gerado assim:
type Queries struct {
	db *sql.DB
}

func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}

// Métodos do SQLC (exemplo)
func (q *Queries) GetCustomerByID(ctx context.Context, id string) (*Customer, error) {
	query := `SELECT id, name, email FROM customers WHERE id = $1`
	row := q.db.QueryRowContext(ctx, query, id)

	var customer Customer
	err := row.Scan(&customer.ID, &customer.Name, &customer.Email)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

type Customer struct {
	ID    string
	Name  string
	Email string
}

// Agora você pode usar EXATAMENTE como no seu outro projeto:
func ExampleUsagePatternLikeYourProject(tenant string, customerId string) error {
	// Factory para o SQLC
	factory := func(db *sql.DB) *Queries {
		return NewQueries(db)
	}

	// USO IGUAL AO SEU OUTRO PROJETO:
	queries, dbConn, err := GetConnection(context.Background(), tenant, factory)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer dbConn.Close() // ← Igual ao seu padrão!

	// Usa o SQLC normalmente
	customerExists, err := queries.GetCustomerByID(context.Background(), customerId)
	if err != nil {
		return fmt.Errorf("error getting customer: %w", err)
	}

	log.Printf("Customer found: %+v", customerExists)
	return nil
}

// Usando com tenant no contexto (mais elegante para APIs):
func ExampleUsageWithTenantInContext(customerId string) error {
	// Simula um contexto com tenant (normalmente vem do middleware)
	ctx := context.WithValue(context.Background(), TenantContextKey, "meu_tenant")

	factory := func(db *sql.DB) *Queries {
		return NewQueries(db)
	}

	// USO AINDA MAIS SIMPLES (sem precisar passar tenant):
	queries, dbConn, err := GetConnectionFromContext(ctx, factory)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer dbConn.Close()

	customerExists, err := queries.GetCustomerByID(ctx, customerId)
	if err != nil {
		return fmt.Errorf("error getting customer: %w", err)
	}

	log.Printf("Customer found: %+v", customerExists)
	return nil
}

// Se você quiser usar com configurações customizadas:
func ExampleUsageWithCustomOptions(tenant string, customerId string) error {
	factory := func(db *sql.DB) *Queries {
		return NewQueries(db)
	}

	// Com opções customizadas
	opts := TenantConnectOptions{
		Tenant:       tenant,
		MaxOpenConns: 50,
		ForceUTC:     true,
		CacheEnabled: true,
	}

	queries, dbConn, err := GetConnectionWithOptions(context.Background(), factory, opts)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer dbConn.Close()

	customerExists, err := queries.GetCustomerByID(context.Background(), customerId)
	if err != nil {
		return fmt.Errorf("error getting customer: %w", err)
	}

	log.Printf("Customer found: %+v", customerExists)
	return nil
}
