package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("spark not found")

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.Init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func OpenInMemory() (*Store, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &Store{db: db}
	if err := store.Init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS sparks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	parent_id INTEGER NULL REFERENCES sparks(id) ON DELETE SET NULL,
	important BOOLEAN NOT NULL DEFAULT 0,
	done BOOLEAN NOT NULL DEFAULT 0,
	deleted BOOLEAN NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	completed_at DATETIME NULL,
	deleted_at DATETIME NULL
);
CREATE INDEX IF NOT EXISTS idx_sparks_active ON sparks(deleted, done, important, created_at);
CREATE INDEX IF NOT EXISTS idx_sparks_parent ON sparks(parent_id);
CREATE INDEX IF NOT EXISTS idx_sparks_title ON sparks(title);
`)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	return nil
}

func (s *Store) Add(ctx context.Context, title string) (model.Spark, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Spark{}, errors.New("spark title cannot be empty")
	}

	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `
INSERT INTO sparks (title, created_at, updated_at)
VALUES (?, ?, ?)`, title, now, now)
	if err != nil {
		return model.Spark{}, fmt.Errorf("add spark: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return model.Spark{}, fmt.Errorf("read new spark id: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *Store) Get(ctx context.Context, id int64) (model.Spark, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, title, parent_id, important, done, deleted, created_at, updated_at, completed_at, deleted_at
FROM sparks
WHERE id = ? AND deleted = 0`, id)
	return scanSpark(row)
}

func (s *Store) List(ctx context.Context, opts model.ListOptions) ([]model.Spark, error) {
	query := `
SELECT id, title, parent_id, important, done, deleted, created_at, updated_at, completed_at, deleted_at
FROM sparks
WHERE deleted = 0`
	if !opts.IncludeAll && !opts.IncludeDone {
		query += " AND done = 0"
	}
	query += " ORDER BY important DESC, done ASC, created_at ASC, id ASC"

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list sparks: %w", err)
	}
	defer rows.Close()

	return scanSparks(rows)
}

func (s *Store) Search(ctx context.Context, query string, opts model.ListOptions) ([]model.Spark, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	sqlQuery := `
SELECT id, title, parent_id, important, done, deleted, created_at, updated_at, completed_at, deleted_at
FROM sparks
WHERE deleted = 0 AND lower(title) LIKE lower(?)`
	if !opts.IncludeAll && !opts.IncludeDone {
		sqlQuery += " AND done = 0"
	}
	sqlQuery += " ORDER BY important DESC, done ASC, created_at ASC, id ASC"

	rows, err := s.db.QueryContext(ctx, sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("search sparks: %w", err)
	}
	defer rows.Close()

	return scanSparks(rows)
}

func (s *Store) MarkDone(ctx context.Context, id int64) (model.Spark, error) {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `
UPDATE sparks
SET done = 1, completed_at = COALESCE(completed_at, ?), updated_at = ?
WHERE id = ? AND deleted = 0`, now, now, id)
	if err != nil {
		return model.Spark{}, fmt.Errorf("mark spark done: %w", err)
	}
	if err := ensureChanged(res); err != nil {
		return model.Spark{}, err
	}
	return s.Get(ctx, id)
}

func (s *Store) ToggleImportant(ctx context.Context, id int64) (model.Spark, error) {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `
UPDATE sparks
SET important = CASE important WHEN 1 THEN 0 ELSE 1 END, updated_at = ?
WHERE id = ? AND deleted = 0`, now, id)
	if err != nil {
		return model.Spark{}, fmt.Errorf("toggle important: %w", err)
	}
	if err := ensureChanged(res); err != nil {
		return model.Spark{}, err
	}
	return s.Get(ctx, id)
}

func (s *Store) Remove(ctx context.Context, id int64) error {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `
UPDATE sparks
SET deleted = 1, deleted_at = ?, updated_at = ?
WHERE id = ? AND deleted = 0`, now, now, id)
	if err != nil {
		return fmt.Errorf("remove spark: %w", err)
	}
	return ensureChanged(res)
}

func (s *Store) ClearCompleted(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `
UPDATE sparks
SET deleted = 1, deleted_at = ?, updated_at = ?
WHERE done = 1 AND deleted = 0`, now, now)
	if err != nil {
		return 0, fmt.Errorf("clear completed sparks: %w", err)
	}
	return res.RowsAffected()
}

func (s *Store) ClearAll(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx, `
UPDATE sparks
SET deleted = 1, deleted_at = ?, updated_at = ?
WHERE deleted = 0`, now, now)
	if err != nil {
		return 0, fmt.Errorf("clear sparks: %w", err)
	}
	return res.RowsAffected()
}

func scanSparks(rows *sql.Rows) ([]model.Spark, error) {
	var sparks []model.Spark
	for rows.Next() {
		var spark model.Spark
		if err := scanInto(rows.Scan, &spark); err != nil {
			return nil, err
		}
		sparks = append(sparks, spark)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read sparks: %w", err)
	}
	return sparks, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSpark(row rowScanner) (model.Spark, error) {
	var spark model.Spark
	if err := scanInto(row.Scan, &spark); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Spark{}, ErrNotFound
		}
		return model.Spark{}, err
	}
	return spark, nil
}

func scanInto(scan func(dest ...any) error, spark *model.Spark) error {
	var parentID sql.NullInt64
	var completedAt sql.NullTime
	var deletedAt sql.NullTime
	if err := scan(
		&spark.ID,
		&spark.Title,
		&parentID,
		&spark.Important,
		&spark.Done,
		&spark.Deleted,
		&spark.CreatedAt,
		&spark.UpdatedAt,
		&completedAt,
		&deletedAt,
	); err != nil {
		return fmt.Errorf("scan spark: %w", err)
	}
	if parentID.Valid {
		spark.ParentID = &parentID.Int64
	}
	if completedAt.Valid {
		spark.CompletedAt = &completedAt.Time
	}
	if deletedAt.Valid {
		spark.DeletedAt = &deletedAt.Time
	}
	return nil
}

func ensureChanged(res sql.Result) error {
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}
	if changed == 0 {
		return ErrNotFound
	}
	return nil
}
