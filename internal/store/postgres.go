package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hollgett/shortener.git/internal/logger"
	"github.com/hollgett/shortener.git/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQLStore struct {
	logger             *logger.Logger
	DB                 *sql.DB
	insertStmt         *sql.Stmt
	selectShortStmt    *sql.Stmt
	selectOriginalStmt *sql.Stmt
	selectUserURLsStmt *sql.Stmt
}

// try open connection and ping server
func newConn(DSN string) (*sql.DB, error) {
	db, err := sql.Open("pgx", DSN)
	if err != nil {
		return nil, fmt.Errorf("failed open connection to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed ping database error: %w", err)
	}
	return db, nil
}

func getPGError(err error) *pgconn.PgError {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError
	}
	return nil
}

// NewPostgreSQLStore create new connection to PostgreSQL and return error if newConn have problem with open connection and ping database.
func NewPostgreSQLStore(logger *logger.Logger, DSN string) (*PostgreSQLStore, error) {
	postgreSQLStore := PostgreSQLStore{
		logger: logger,
	}

	db, err := newConn(DSN)
	if err != nil {
		return nil, fmt.Errorf("new conn error: %w", err)
	}
	postgreSQLStore.DB = db

	if err := postgreSQLStore.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed run migrations: %w", err)
	}

	STMTs := []struct {
		name     string
		query    string
		addrStmt **sql.Stmt
	}{
		{"insert url", InsertReq, &postgreSQLStore.insertStmt},
		{"select original", SelectOriginalReq, &postgreSQLStore.selectOriginalStmt},
		{"select short", selectShortReq, &postgreSQLStore.selectShortStmt},
		{"select user URLs", SelectUserURLsReq, &postgreSQLStore.selectUserURLsStmt},
	}
	for _, stmt := range STMTs {
		prep, err := postgreSQLStore.DB.Prepare(stmt.query)
		if err != nil {
			err := postgreSQLStore.Close()
			return nil, errors.Join(fmt.Errorf("failed create stmt %s: %w", stmt.name, err), err)
		}
		*stmt.addrStmt = prep
	}

	return &postgreSQLStore, nil
}

func (p *PostgreSQLStore) runMigrations() error {
	driver, err := postgres.WithInstance(p.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed create driver migrations: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "yandex", driver)
	if err != nil {
		return fmt.Errorf("failed create migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed migration up: %w", err)
	}

	return nil
}

func (p *PostgreSQLStore) closeStmt() error {
	var errs []error
	if err := p.insertStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed close %s stmt: %w", "insertReq", err))
	}

	if err := p.selectOriginalStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed close %s stmt: %w", "selectOriginalReq", err))
	}

	if err := p.selectShortStmt.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed close %s stmt: %w", "selectShortStmt", err))
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (p *PostgreSQLStore) Ping() error {
	return p.DB.Ping()
}

func (p *PostgreSQLStore) SaveShortURL(URL models.ShortenerURL) (string, error) {
	_, err := p.insertStmt.Query(URL.OriginalURL, URL.ShortURL, URL.UserID)
	if err == nil {
		return "", nil
	} else if pgErr := getPGError(err); pgErr != nil && pgErr.Code == pgerrcode.UniqueViolation {
		var shortExists string
		if err := p.selectShortStmt.QueryRow(URL.OriginalURL).Scan(&shortExists); err != nil {
			return "", fmt.Errorf("failed select short link: %w", err)
		}
		return shortExists, ErrShortExists
	}
	return "", fmt.Errorf("failed insert exec: %w", err)
}

func (p *PostgreSQLStore) SaveShortURLs(URLs []models.ShortenerURL) ([]models.ShortenerURL, error) {
	// create transaction
	tx, err := p.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed begin transaction: %w", err)
	}
	defer tx.Rollback()

	// set prepare to transaction
	insertTx, err := tx.Prepare(InsertReq)
	if err != nil {
		return nil, fmt.Errorf("failed set prepare insert to transaction: %w", err)
	}

	// begin transaction request
	for _, v := range URLs {
		// //create save point transaction
		// if _, err := tx.Exec(fmt.Sprintf("SAVEPOINT sp%s", i)); err != nil {
		// 	tx.Rollback()
		// 	return nil, fmt.Errorf("failed create save point: %w", err)
		// }

		// insert to database
		if _, err := insertTx.Exec(v.OriginalURL, v.ShortURL, v.UserID); err != nil {
			// tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT sp%s", i))
			return nil, fmt.Errorf("failed insert original: %s, short: %s: %w", v.OriginalURL, v.ShortURL, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed commit transaction: %w", err)
	}

	return URLs, nil
}

func (p *PostgreSQLStore) GetOriginalURL(ShortLink string) (string, error) {
	row := p.selectOriginalStmt.QueryRow(ShortLink)

	var originalURL string
	var is_deleted bool
	if err := row.Scan(&originalURL, &is_deleted); err == sql.ErrNoRows {
		return "", ErrIsNotExists
	} else if err != nil {
		return "", fmt.Errorf("failed scan row: %w", err)
	}

	if is_deleted {
		return "", ErrURLDeleted
	}

	return originalURL, nil
}

func (p *PostgreSQLStore) GetUserURLs(userID string) ([]models.URLResponse, error) {
	rows, err := p.selectUserURLsStmt.Query(userID)
	if err != nil {
		return nil, fmt.Errorf("failed query: %w", err)
	}
	defer rows.Close()
	userURLs := make([]models.URLResponse, 0)
	for rows.Next() {
		var userURL models.URLResponse
		err := rows.Scan(&userURL.ShortURL, &userURL.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("failed scan rows: %w", err)
		}
		userURLs = append(userURLs, userURL)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if len(userURLs) == 0 {
		return nil, ErrUserURLsNotExists
	}

	return userURLs, nil
}

func (p *PostgreSQLStore) DeleteURLs(URLs []models.DeleteURL) error {
	var query strings.Builder
	args := make([]interface{}, 0)
	query.WriteString(`
	UPDATE shortener_urls AS s
	SET is_deleted = TRUE
	FROM (VALUES`)

	for i, v := range URLs {
		query.WriteString(fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		if i != len(URLs)-1 {
			query.WriteString(",")
		}
		args = append(args, v.UserID, v.ShortURL)
	}

	query.WriteString(`) AS tmp(user_id, short)
	WHERE s.user_id = tmp.user_id
  	AND s.short = tmp.short;`)

	if _, err := p.DB.Exec(query.String(), args...); err != nil {
		return fmt.Errorf("failed batch delete urls: %w", err)
	}
	return nil
}

func (p *PostgreSQLStore) Close() error {
	errStmt := p.closeStmt()
	errDB := p.DB.Close()
	return errors.Join(errStmt, errDB)
}
