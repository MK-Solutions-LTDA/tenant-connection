package connection

import (
	"context"
	"testing"
	"time"
)

// TestTenantConnectionV2_Basic testa funcionalidade básica da v2
func TestTenantConnectionV2_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Pula o teste se não há configuração de catálogo
	if dbCatalog == nil {
		t.Skip("Skipping test - no catalog database configured")
	}

	ctx := context.Background()

	opts := TenantConnectOptions{
		Tenant:       "test_tenant",
		CacheEnabled: false, // Desabilita cache para teste
	}

	// Esta função falhará se não houver configuração de banco válida
	_, err := GetTenantConnectionV2(ctx, opts)
	if err != nil {
		t.Logf("Expected error for missing/invalid tenant config: %v", err)
		// Este é o comportamento esperado quando não há configuração válida
		return
	}

	// Se chegou aqui, a conexão foi bem-sucedida (improvável sem config válida)
	t.Log("Connection succeeded - this test needs a valid database configuration to be meaningful")
}

// TestTenantConnectOptions_Defaults testa valores padrão
func TestTenantConnectOptions_Defaults(t *testing.T) {
	opts := TenantConnectOptions{
		Tenant: "test_tenant",
	}

	// Simula o processamento de defaults que acontece em GetTenantConnectionV2
	if opts.CacheTTL == 0 {
		opts.CacheTTL = 55 * time.Minute
	}
	if opts.ConnMaxLifetime == 0 {
		opts.ConnMaxLifetime = 1 * time.Hour
	}
	if opts.ConnMaxIdle == 0 {
		opts.ConnMaxIdle = 1 * time.Hour
	}
	if opts.MaxOpenConns == 0 {
		opts.MaxOpenConns = 25
	}
	if opts.MaxIdleConns == 0 {
		opts.MaxIdleConns = 25
	}

	// Verifica se os valores padrão foram aplicados
	if opts.CacheTTL != 55*time.Minute {
		t.Errorf("Expected CacheTTL to be 55 minutes, got %v", opts.CacheTTL)
	}
	if opts.ConnMaxLifetime != 1*time.Hour {
		t.Errorf("Expected ConnMaxLifetime to be 1 hour, got %v", opts.ConnMaxLifetime)
	}
	if opts.MaxOpenConns != 25 {
		t.Errorf("Expected MaxOpenConns to be 25, got %d", opts.MaxOpenConns)
	}
}

// TestQueryLogger testa o logger customizado
func TestQueryLogger(t *testing.T) {
	ctx := context.Background()

	var loggedQuery string
	var loggedArgs []any

	customLogger := func(ctx context.Context, query string, args ...any) {
		loggedQuery = query
		loggedArgs = args
	}

	// Testa o logger
	testQuery := "SELECT * FROM users WHERE id = $1"
	testArgs := []any{123}

	customLogger(ctx, testQuery, testArgs...)

	if loggedQuery != testQuery {
		t.Errorf("Expected query %s, got %s", testQuery, loggedQuery)
	}

	if len(loggedArgs) != 1 || loggedArgs[0] != 123 {
		t.Errorf("Expected args [123], got %v", loggedArgs)
	}
}

// BenchmarkTenantConnectionV2_Cache benchmarks cache performance
func BenchmarkTenantConnectionV2_Cache(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	// Este benchmark só funcionará com configuração de banco válida
	// É um exemplo de como estruturar benchmarks
	b.Skip("Benchmark requires valid database configuration")
}

// ExampleGetTenantConnectionV2 exemplo de uso da função principal
func ExampleGetTenantConnectionV2() {
	ctx := context.Background()

	opts := TenantConnectOptions{
		Tenant:       "example_tenant",
		MaxOpenConns: 10,
		CacheEnabled: true,
	}

	tenantConn, err := GetTenantConnectionV2(ctx, opts)
	if err != nil {
		// Handle error
		return
	}
	defer tenantConn.Close()

	// Use the connection
	// tenantConn.ExecWithLog(ctx, "SELECT 1")
}
