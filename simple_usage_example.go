package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

// ===== SETUP INICIAL (FAZER UMA VEZ ONLY) =====

// Este exemplo usa o mesmo Queries do usage_example.go
// No seu projeto real, você importaria: "your-project/db" ou onde seu SQLC está

// ===== CONFIGURAÇÃO INICIAL NO MAIN() OU INIT() =====
func SetupGlobalFactory() {
	// Configure o factory global UMA VEZ no início da aplicação
	factory := func(db *sql.DB) *Queries {
		return NewQueries(db)
	}

	SetGlobalFactory(factory)
	log.Println("Global factory configured - now you can use simplified functions!")
}

// ===== AGORA VOCÊ PODE USAR DE FORMA MUITO MAIS SIMPLES =====

// OPÇÃO 1: Super simples (só precisa do tenant)
func ExampleSimpleUsage(tenant string, customerId string) error {
	// Sem factory! Usa o global que foi configurado
	queries, dbConn, err := GetConnectionSimple[*Queries](context.Background(), tenant)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer dbConn.Close()

	// Usar normalmente
	customer, err := queries.GetCustomerByID(context.Background(), customerId)
	if err != nil {
		return fmt.Errorf("error getting customer: %w", err)
	}

	log.Printf("Customer found: %+v", customer)
	return nil
}

// OPÇÃO 2: Ultra simples (tenant vem do contexto automaticamente)
func ExampleUltraSimple(ctx context.Context, customerId string) error {
	// Sem factory! Sem tenant! Tudo automático
	queries, dbConn, err := GetConnectionFromContextSimple[*Queries](ctx)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer dbConn.Close()

	customer, err := queries.GetCustomerByID(ctx, customerId)
	if err != nil {
		return fmt.Errorf("error getting customer: %w", err)
	}

	log.Printf("Customer found: %+v", customer)
	return nil
}

// OPÇÃO 3: Para casos especiais (ainda pode usar a forma completa)
func ExampleWithCustomFactory(tenant string, customerId string) error {
	// Factory customizado para casos especiais
	customFactory := func(db *sql.DB) *Queries {
		// Alguma configuração especial aqui
		return NewQueries(db)
	}

	queries, dbConn, err := GetConnection(context.Background(), tenant, customFactory)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer dbConn.Close()

	customer, err := queries.GetCustomerByID(context.Background(), customerId)
	if err != nil {
		return fmt.Errorf("error getting customer: %w", err)
	}

	log.Printf("Customer found: %+v", customer)
	return nil
}

// ===== EXEMPLO DE UM HANDLER COMPLETO =====

func YourRealHandler(w http.ResponseWriter, r *http.Request) {
	// Pega tenant do header (exemplo)
	tenant := r.Header.Get("X-Tenant-ID")

	// USO SUPER SIMPLES:
	queries, dbConn, err := GetConnectionSimple[*Queries](r.Context(), tenant)
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	// Exemplo de uso das queries
	customer, err := queries.GetCustomerByID(r.Context(), "123")
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Customer: %+v", customer)
}

// Ou ainda mais simples se você usar middleware de tenant:
func YourUltraSimpleHandler(w http.ResponseWriter, r *http.Request) {
	// Tenant já está no contexto (middleware)
	queries, dbConn, err := GetConnectionFromContextSimple[*Queries](r.Context())
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()

	// Exemplo de uso das queries
	customer, err := queries.GetCustomerByID(r.Context(), "123")
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Customer: %+v", customer)
}
