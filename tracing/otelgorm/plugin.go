package otelgorm

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type plugin struct {
	dbName              string
	logQueries          bool
	recordRowsAffected  bool
	recordSQLParameters bool
	excludeTables       map[string]bool
}

// Option configures the GORM tracing plugin.
type Option func(*plugin)

// WithDBName sets the database name for trace attribution.
// This appears in the span as the db.name attribute.
//
// Parameters:
//   - name: The database name.
//
// Example:
//
//	db.Use(otelgorm.NewPlugin(otelgorm.WithDBName("mydb")))
func WithDBName(name string) Option {
	return func(p *plugin) {
		p.dbName = name
	}
}

// WithLogQueries enables or disables logging of SQL queries in spans.
// When enabled, the full SQL statement is recorded as the db.statement attribute.
// WARNING: Be careful with sensitive data in production - queries may contain
// sensitive information like passwords or PII.
//
// Parameters:
//   - enabled: true to log queries, false to omit them.
//
// Example:
//
//	// Enable for development
//	db.Use(otelgorm.NewPlugin(otelgorm.WithLogQueries(true)))
//
//	// Disable for production
//	db.Use(otelgorm.NewPlugin(otelgorm.WithLogQueries(false)))
func WithLogQueries(enabled bool) Option {
	return func(p *plugin) {
		p.logQueries = enabled
	}
}

// WithRecordRowsAffected enables or disables recording the number of rows affected.
// When enabled, db.rows_affected is added as a span attribute.
//
// Parameters:
//   - enabled: true to record rows affected, false to omit.
//
// Example:
//
//	db.Use(otelgorm.NewPlugin(otelgorm.WithRecordRowsAffected(true)))
func WithRecordRowsAffected(enabled bool) Option {
	return func(p *plugin) {
		p.recordRowsAffected = enabled
	}
}

// WithRecordSQLParameters enables or disables recording of SQL parameters.
// WARNING: This can expose sensitive data. Use only in development environments.
//
// Parameters:
//   - enabled: true to record SQL parameters, false to omit.
//
// Example:
//
//	// Only enable in development
//	if os.Getenv("ENVIRONMENT") == "local" {
//	    db.Use(otelgorm.NewPlugin(otelgorm.WithRecordSQLParameters(true)))
//	}
func WithRecordSQLParameters(enabled bool) Option {
	return func(p *plugin) {
		p.recordSQLParameters = enabled
	}
}

// WithExcludeTables excludes specific tables from tracing.
// Useful for high-frequency tables that don't need tracing (e.g., session tables).
//
// Parameters:
//   - tables: Table names to exclude from tracing.
//
// Example:
//
//	db.Use(otelgorm.NewPlugin(
//	    otelgorm.WithExcludeTables("sessions", "audit_log"),
//	))
func WithExcludeTables(tables ...string) Option {
	return func(p *plugin) {
		for _, t := range tables {
			p.excludeTables[t] = true
		}
	}
}

var _ gorm.Plugin = (*plugin)(nil)

// NewPlugin creates a GORM plugin that automatically traces all database operations.
//
// The plugin traces the following operations:
//   - Create (INSERT)
//   - Query (SELECT)
//   - Update (UPDATE)
//   - Delete (DELETE)
//   - Row (Row queries)
//   - Raw (Raw SQL)
//
// Each operation produces a client span with standard DB semantic conventions:
//   - db.system: "sql"
//   - db.operation: The operation type (Create, Query, etc.)
//   - db.sql.table: The table name (if available)
//   - db.name: The database name (if configured)
//   - db.statement: The SQL query (if WithLogQueries is enabled)
//
// Parameters:
//   - opts: Optional configuration options.
//
// Returns:
//   - gorm.Plugin: A GORM plugin that can be registered with db.Use().
//
// Example:
//
//	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = db.Use(otelgorm.NewPlugin(
//	    otelgorm.WithDBName("mydb"),
//	    otelgorm.WithLogQueries(os.Getenv("ENVIRONMENT") == "local"),
//	    otelgorm.WithRecordRowsAffected(true),
//	))
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewPlugin(opts ...Option) gorm.Plugin {
	p := &plugin{
		excludeTables: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *plugin) Name() string {
	return "otelgorm"
}

func (p *plugin) Initialize(db *gorm.DB) error {
	callbacks := db.Callback()

	if err := callbacks.Create().Before("gorm:create").Register("otel:before_create", p.beforeCallback("Create")); err != nil {
		return err
	}
	if err := callbacks.Create().After("gorm:create").Register("otel:after_create", p.afterCallback()); err != nil {
		return err
	}

	if err := callbacks.Query().Before("gorm:query").Register("otel:before_query", p.beforeCallback("Query")); err != nil {
		return err
	}
	if err := callbacks.Query().After("gorm:query").Register("otel:after_query", p.afterCallback()); err != nil {
		return err
	}

	if err := callbacks.Update().Before("gorm:update").Register("otel:before_update", p.beforeCallback("Update")); err != nil {
		return err
	}
	if err := callbacks.Update().After("gorm:update").Register("otel:after_update", p.afterCallback()); err != nil {
		return err
	}

	if err := callbacks.Delete().Before("gorm:delete").Register("otel:before_delete", p.beforeCallback("Delete")); err != nil {
		return err
	}
	if err := callbacks.Delete().After("gorm:delete").Register("otel:after_delete", p.afterCallback()); err != nil {
		return err
	}

	if err := callbacks.Row().Before("gorm:row").Register("otel:before_row", p.beforeCallback("Row")); err != nil {
		return err
	}
	if err := callbacks.Row().After("gorm:row").Register("otel:after_row", p.afterCallback()); err != nil {
		return err
	}

	if err := callbacks.Raw().Before("gorm:raw").Register("otel:before_raw", p.beforeCallback("Raw")); err != nil {
		return err
	}
	if err := callbacks.Raw().After("gorm:raw").Register("otel:after_raw", p.afterCallback()); err != nil {
		return err
	}

	return nil
}

func (p *plugin) beforeCallback(operation string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context

		tableName := db.Statement.Table
		if _, excluded := p.excludeTables[tableName]; excluded {
			return
		}

		attrs := []attribute.KeyValue{
			attribute.String("db.system", "sql"),
			attribute.String("db.operation", operation),
		}

		if p.dbName != "" {
			attrs = append(attrs, attribute.String("db.name", p.dbName))
		}

		if tableName != "" {
			attrs = append(attrs, attribute.String("db.sql.table", tableName))
		}

		if p.logQueries {
			attrs = append(attrs, attribute.String("db.statement", db.Statement.SQL.String()))
		}

		spanName := operation
		if tableName != "" {
			spanName = tableName + "." + operation
		}

		ctx, span := tracer(ctx).Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attrs...),
		)

		db.InstanceSet("otel:span", span)
		db.InstanceSet("otel:start_time", time.Now())
		db.Statement.Context = ctx
	}
}

func (p *plugin) afterCallback() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		spanInterface, ok := db.InstanceGet("otel:span")
		if !ok {
			return
		}

		span, ok := spanInterface.(trace.Span)
		if !ok {
			return
		}

		if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
			span.RecordError(db.Error)
			span.SetStatus(codes.Error, db.Error.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		if p.recordRowsAffected {
			span.SetAttributes(attribute.Int64("db.rows_affected", db.RowsAffected))
		}

		span.End()
	}
}

func tracer(ctx context.Context) trace.Tracer {
	return trace.SpanFromContext(ctx).TracerProvider().Tracer("qbp-lib")
}
