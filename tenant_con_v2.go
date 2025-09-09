package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// TenantConnectOptions configurações para conexão v2 com tenant
type TenantConnectOptions struct {
	Tenant          string        // Nome do tenant (obrigatório)
	MaxOpenConns    int           // Número máximo de conexões abertas
	MaxIdleConns    int           // Número máximo de conexões idle
	ConnMaxIdle     time.Duration // Tempo máximo que uma conexão pode ficar idle
	ConnMaxLifetime time.Duration // Tempo máximo de vida de uma conexão
	ForceUTC        bool          // Força timezone UTC
	QueryLogger     QueryLogger   // Logger personalizado para queries
	CacheEnabled    bool          // Se deve usar cache (padrão: true)
	CacheTTL        time.Duration // TTL do cache (padrão: 55min)
}

// TenantConnectionV2 representa uma conexão v2 com tenant
type TenantConnectionV2 struct {
	DB         *sql.DB
	SearchPath string
	Options    TenantConnectOptions
	createdAt  time.Time
	mu         sync.RWMutex // Protege contra race conditions
	closed     bool         // Flag para saber se foi fechada
}

// QueryLogger função para log de queries
type QueryLogger func(ctx context.Context, query string, args ...any)

// SqlcFactory factory para criar instâncias do sqlc
type SqlcFactory[T any] func(db *sql.DB) T

var defaultLogger QueryLogger = DefaultTenantLogger

// DefaultTenantLogger logger padrão
func DefaultTenantLogger(ctx context.Context, query string, args ...any) {
	fmt.Printf("[TenantQuery] %s | args: %v\n", query, args)
}

// GetTenantConnectionV2 obtém uma conexão v2 para o tenant com opções avançadas
func GetTenantConnectionV2(ctx context.Context, opts TenantConnectOptions) (*TenantConnectionV2, error) {
	if opts.Tenant == "" {
		return nil, fmt.Errorf("tenant name is required")
	}

	// Define valores padrão
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

	// Verifica cache se habilitado (padrão é true)
	if opts.CacheEnabled || (!opts.CacheEnabled && opts.CacheTTL > 0) {
		opts.CacheEnabled = true
	}

	if opts.CacheEnabled {
		Mutex.Lock()
		cacheKey := prefixConnection + "v2-" + opts.Tenant
		if conn, found := Connections.Get(cacheKey); found {
			Mutex.Unlock()
			tenantConn := conn.(*TenantConnectionV2)
			// Verifica se a conexão ainda está válida
			if err := tenantConn.DB.PingContext(ctx); err == nil {
				return tenantConn, nil
			}
			// Remove conexão inválida do cache
			Connections.Del(cacheKey)
		}
		Mutex.Unlock()
	}

	// Busca informações do tenant no catálogo
	catalog, err := GetTenant(opts.Tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant info: %w", err)
	}

	// Constrói a DSN
	dsn := fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=disable",
		catalog.Driver, catalog.UserName, catalog.Password,
		catalog.Server, catalog.DatabaseName)

	// Abre a conexão
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection for tenant %s: %w", opts.Tenant, err)
	}

	// Configura parâmetros da conexão
	db.SetMaxOpenConns(opts.MaxOpenConns)
	db.SetMaxIdleConns(opts.MaxIdleConns)
	db.SetConnMaxIdleTime(opts.ConnMaxIdle)
	db.SetConnMaxLifetime(opts.ConnMaxLifetime)

	// Usa context independente para configuração inicial
	// Isso evita que context timeout do usuário cancele a configuração da conexão
	setupCtx, setupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer setupCancel()

	// Testa a conexão
	if err := db.PingContext(setupCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database for tenant %s: %w", opts.Tenant, err)
	}

	// Configura o search_path para o tenant
	if _, err := db.ExecContext(setupCtx, fmt.Sprintf("SET search_path TO %s", opts.Tenant)); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set search_path for tenant %s: %w", opts.Tenant, err)
	}

	// Força UTC se solicitado
	if opts.ForceUTC {
		if _, err := db.ExecContext(setupCtx, "SET TIMEZONE='UTC'"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set timezone for tenant %s: %w", opts.Tenant, err)
		}
	}

	// Cria a conexão do tenant
	tenantConn := &TenantConnectionV2{
		DB:         db,
		SearchPath: opts.Tenant,
		Options:    opts,
		createdAt:  time.Now(),
	}

	// Valida uma última vez se a conexão está realmente funcional
	if err := tenantConn.DB.PingContext(setupCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("final connection validation failed for tenant %s: %w", opts.Tenant, err)
	}

	// Salva no cache se habilitado
	if opts.CacheEnabled {
		Mutex.Lock()
		cacheKey := prefixConnection + "v2-" + opts.Tenant
		Connections.SetWithTTL(cacheKey, tenantConn, 1, opts.CacheTTL)
		Mutex.Unlock()
	}

	log.Printf("TenantConnectionV2 created for tenant: %s", opts.Tenant)
	return tenantConn, nil
}

// Close fecha a conexão do tenant
func (tc *TenantConnectionV2) Close() error {
	// Verificação robusta para evitar panics
	if tc == nil {
		return nil
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Verifica se já foi fechada
	if tc.closed || tc.DB == nil {
		return nil
	}

	// Remove do cache se estava sendo usado
	if tc.Options.CacheEnabled && Connections != nil {
		Mutex.Lock()
		cacheKey := prefixConnection + "v2-" + tc.Options.Tenant
		Connections.Del(cacheKey)
		Mutex.Unlock()
	}

	err := tc.DB.Close()
	// ⚠️ CRÍTICO: NÃO setar tc.DB = nil para evitar panic no SQLC
	// Apenas marcar como fechada
	tc.closed = true
	log.Printf("TenantConnectionV2 closed for tenant: %s", tc.Options.Tenant)
	return err
}

// IsHealthy verifica se a conexão está saudável
func (tc *TenantConnectionV2) IsHealthy(ctx context.Context) bool {
	if tc == nil {
		return false
	}

	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.closed || tc.DB == nil {
		return false
	}

	return tc.DB.PingContext(ctx) == nil
}

// GetAge retorna a idade da conexão
func (tc *TenantConnectionV2) GetAge() time.Duration {
	return time.Since(tc.createdAt)
}

// GetDB retorna o *sql.DB de forma thread-safe
func (tc *TenantConnectionV2) GetDB() *sql.DB {
	if tc == nil {
		return nil
	}

	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Se foi fechada, retorna nil mas mantém tc.DB intacto para evitar panic no SQLC
	if tc.closed {
		return nil
	}

	return tc.DB
}

// NewSqlcWithTenantConnection cria uma instância do sqlc com conexão de tenant
func NewSqlcWithTenantConnection[T any](ctx context.Context, factory SqlcFactory[T], opts TenantConnectOptions) (T, *TenantConnectionV2, error) {
	tenantConn, err := GetTenantConnectionV2(ctx, opts)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	// Verifica se a conexão é válida
	db := tenantConn.GetDB()
	if db == nil {
		var zero T
		return zero, nil, fmt.Errorf("failed to get valid database connection")
	}

	// Por enquanto, passamos a DB diretamente mas com verificações robustas
	// TODO: Implementar wrapper completo no futuro se necessário
	return factory(db), tenantConn, nil
}

// GetConnection segue o mesmo padrão do seu outro projeto
// Retorna: queries, dbConn, err (igual ao dbutils.GetConnection)
func GetConnection[T any](ctx context.Context, tenant string, factory SqlcFactory[T]) (T, *TenantConnectionV2, error) {
	opts := TenantConnectOptions{
		Tenant:       tenant,
		CacheEnabled: true, // Usa cache por padrão
	}

	return NewSqlcWithTenantConnection(ctx, factory, opts)
}

// GetConnectionWithOptions permite configurações customizadas mas mantém a mesma assinatura
func GetConnectionWithOptions[T any](ctx context.Context, factory SqlcFactory[T], opts TenantConnectOptions) (T, *TenantConnectionV2, error) {
	return NewSqlcWithTenantConnection(ctx, factory, opts)
}

// GetConnectionFromContext extrai o tenant do contexto automaticamente (se você usar essa abordagem)
// Você pode definir uma chave no contexto para o tenant
type tenantKeyType string

const TenantContextKey tenantKeyType = "tenant"

func GetConnectionFromContext[T any](ctx context.Context, factory SqlcFactory[T]) (T, *TenantConnectionV2, error) {
	tenant, ok := ctx.Value(TenantContextKey).(string)
	if !ok || tenant == "" {
		var zero T
		return zero, nil, fmt.Errorf("tenant not found in context")
	}

	return GetConnection(ctx, tenant, factory)
}

// ExecWithTenantLog executa uma query com log para tenant
func (tc *TenantConnectionV2) ExecWithLog(ctx context.Context, query string, args ...any) (sql.Result, error) {
	db := tc.GetDB()
	if db == nil {
		return nil, fmt.Errorf("connection is closed or invalid")
	}

	start := time.Now()
	logger := tc.Options.QueryLogger
	if logger == nil {
		logger = defaultLogger
	}

	if logger != nil {
		logger(ctx, query, args...)
	}

	res, err := db.ExecContext(ctx, query, args...)
	fmt.Printf("[TenantExec][%s] Took: %s | Error: %v\n", tc.SearchPath, time.Since(start), err)
	return res, err
}

// QueryWithTenantLog executa uma query com log para tenant
func (tc *TenantConnectionV2) QueryWithLog(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	db := tc.GetDB()
	if db == nil {
		return nil, fmt.Errorf("connection is closed or invalid")
	}

	start := time.Now()
	logger := tc.Options.QueryLogger
	if logger == nil {
		logger = defaultLogger
	}

	if logger != nil {
		logger(ctx, query, args...)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	fmt.Printf("[TenantQuery][%s] Took: %s | Error: %v\n", tc.SearchPath, time.Since(start), err)
	return rows, err
}

// QueryRowWithTenantLog executa uma query row com log para tenant
func (tc *TenantConnectionV2) QueryRowWithLog(ctx context.Context, query string, args ...any) *sql.Row {
	db := tc.GetDB()
	if db == nil {
		// Para QueryRow, retornamos um Row que vai dar erro no Scan()
		// Isso mantém a interface compatível
		return &sql.Row{}
	}

	logger := tc.Options.QueryLogger
	if logger == nil {
		logger = defaultLogger
	}

	if logger != nil {
		logger(ctx, query, args...)
	}

	return db.QueryRowContext(ctx, query, args...)
}

// CloseAllTenantConnections fecha todas as conexões v2 de tenants no cache
func CloseAllTenantConnections() error {
	Mutex.Lock()
	defer Mutex.Unlock()

	// Infelizmente o ristretto não tem uma forma fácil de iterar por todas as chaves
	// então esta função serve mais como placeholder para implementação futura
	// Por enquanto, as conexões serão fechadas automaticamente pelo TTL do cache

	log.Println("CloseAllTenantConnections called - connections will be closed by cache TTL")
	return nil
}

// ===== FUNÇÕES DE CONVENIÊNCIA PARA EVITAR REPETIÇÃO =====

// GlobalFactory permite definir um factory global para evitar repetição
var globalFactory any

// SetGlobalFactory define um factory global (chame uma vez no início da aplicação)
func SetGlobalFactory[T any](factory SqlcFactory[T]) {
	globalFactory = factory
}

// GetConnectionSimple usa o factory global (sem precisar passar factory toda vez)
func GetConnectionSimple[T any](ctx context.Context, tenant string) (T, *TenantConnectionV2, error) {
	if globalFactory == nil {
		var zero T
		return zero, nil, fmt.Errorf("global factory not set - use SetGlobalFactory first")
	}

	factory, ok := globalFactory.(SqlcFactory[T])
	if !ok {
		var zero T
		return zero, nil, fmt.Errorf("global factory type mismatch")
	}

	return GetConnection(ctx, tenant, factory)
}

// GetConnectionFromContextSimple combina as duas conveniências: factory global + tenant do contexto
func GetConnectionFromContextSimple[T any](ctx context.Context) (T, *TenantConnectionV2, error) {
	tenant, ok := ctx.Value(TenantContextKey).(string)
	if !ok || tenant == "" {
		var zero T
		return zero, nil, fmt.Errorf("tenant not found in context")
	}

	return GetConnectionSimple[T](ctx, tenant)
}
