package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"
)

// PostgresRepo encapsulates PostgreSQL connection and operations
type PostgresRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewPostgresRepo creates and initializes a PostgreSQL repository
func NewPostgresRepo(dsn string, logger *zap.Logger) (*PostgresRepo, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database connection test failed: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database")

	return &PostgresRepo{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (r *PostgresRepo) Close() error {
	return r.db.Close()
}

// User represents the user model
type User struct {
	ID        int64  `db:"id" json:"id"`
	Username  string `db:"username" json:"username"`
	Email     string `db:"email" json:"email"`
	CreatedAt string `db:"created_at" json:"created_at"`
	UpdatedAt string `db:"updated_at" json:"updated_at"`
}

// GetUser retrieves a user by ID
func (r *PostgresRepo) GetUser(id int64) (*User, error) {
	var user User
	query := `SELECT id, username, email, created_at, updated_at FROM users WHERE id = $1`

	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	return &user, nil
}

// ListUsers retrieves all users with pagination support
func (r *PostgresRepo) ListUsers(limit, offset int) ([]*User, error) {
	users := []*User{}
	query := `
		SELECT id, username, email, created_at, updated_at 
		FROM users 
		ORDER BY id 
		LIMIT $1 OFFSET $2
	`

	err := r.db.Select(&users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user list: %w", err)
	}

	return users, nil
}

// CreateUser creates a new user
func (r *PostgresRepo) CreateUser(username, email string) (*User, error) {
	query := `
		INSERT INTO users (username, email) 
		VALUES ($1, $2) 
		RETURNING id, username, email, created_at, updated_at
	`

	var user User
	err := r.db.QueryRowx(query, username, email).StructScan(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (r *PostgresRepo) UpdateUser(id int64, username, email string) (*User, error) {
	query := `
		UPDATE users 
		SET username = $2, email = $3, updated_at = NOW() 
		WHERE id = $1 
		RETURNING id, username, email, created_at, updated_at
	`

	var user User
	err := r.db.QueryRowx(query, id, username, email).StructScan(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// DeleteUser deletes a user
func (r *PostgresRepo) DeleteUser(id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("no user found with ID %d", id)
	}

	return nil
}
