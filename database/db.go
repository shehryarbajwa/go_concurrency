package database

import (
	"concurrent-downloader/models"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func NewDB(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	fmt.Println("✓ Database connected successfully")
	return &DB{conn: conn}, nil
}

func (db *DB) Insert(todo models.Todo) error {
	query, args, err := sq.Insert("todos").
		Columns("id", "title", "completed").
		Values(todo.ID, todo.Title, todo.Completed).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = db.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert todo: %w", err)
	}
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
