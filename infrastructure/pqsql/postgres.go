package pqsql

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type database struct {
	db      *sql.DB
	wrapper *WrapperTx
}

func (d *database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

type Database interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Transaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) (any, error)) (any, error)
}

func (d *database) Transaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) (any, error)) (any, error) {
	return d.wrapper.WrapTx(ctx, fn)
}

type client struct {
	db *sql.DB
}

type Client interface {
	Database() Database
	PingContext(ctx context.Context) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Begin() (*sql.Tx, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	Close() error
	Ping() error
}

func (c *client) Database() Database {
	return &database{db: c.db, wrapper: NewWrapper(c.db)}
}

func (c *client) PingContext(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, opts)
}

func (c *client) Begin() (*sql.Tx, error) {
	return c.db.Begin()
}

func (c *client) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.db.PrepareContext(ctx, query)
}

func (c *client) Prepare(query string) (*sql.Stmt, error) {
	return c.db.Prepare(query)
}

func (c *client) Close() error {
	return c.db.Close()
}

func (c *client) Ping() error {
	return c.db.Ping()
}

func NewClient(connection string) (Client, error) {
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	client := &client{db: db}

	return client, nil
}
