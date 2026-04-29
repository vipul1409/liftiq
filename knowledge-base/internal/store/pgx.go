package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var migrationStmts = []string{
	`CREATE EXTENSION IF NOT EXISTS vector`,

	`CREATE TABLE IF NOT EXISTS oem_documents (
		id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		filename         VARCHAR(255) NOT NULL,
		controller_make  VARCHAR(100),
		controller_model VARCHAR(100),
		page_count       INT,
		ingested_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		checksum         VARCHAR(64)
	)`,

	`CREATE TABLE IF NOT EXISTS oem_knowledge (
		id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		document_id      UUID REFERENCES oem_documents(id),
		controller_make  VARCHAR(100) NOT NULL,
		controller_model VARCHAR(100),
		section          TEXT NOT NULL,
		content          TEXT NOT NULL,
		page_number      INT,
		chunk_index      INT,
		embedding        vector(768),
		created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	// HNSW index for cosine similarity — no training step required.
	`DO $$ BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM pg_indexes WHERE indexname = 'idx_oem_knowledge_embedding'
		) THEN
			CREATE INDEX idx_oem_knowledge_embedding
				ON oem_knowledge USING hnsw (embedding vector_cosine_ops);
		END IF;
	END $$`,
}

// Connect creates a pgxpool connection and pings the database.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

// Migrate applies all DDL statements (idempotent).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	for i, stmt := range migrationStmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			preview := stmt
			if len(preview) > 60 {
				preview = preview[:60] + "…"
			}
			return fmt.Errorf("migration step %d (%s): %w", i+1, preview, err)
		}
	}
	return nil
}

// PGXStore implements Store using pgxpool + pgvector.
type PGXStore struct {
	pool *pgxpool.Pool
}

// NewPGXStore creates a new PGXStore.
func NewPGXStore(pool *pgxpool.Pool) *PGXStore {
	return &PGXStore{pool: pool}
}

func vectorLiteral(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (s *PGXStore) Search(ctx context.Context, embedding []float32, make, model string, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = 3
	}

	vec := vectorLiteral(embedding)

	query := `
		SELECT k.id, k.document_id, k.controller_make, k.controller_model,
		       k.section, k.content, k.page_number, k.chunk_index, k.created_at,
		       COALESCE(d.filename, '') AS document_name,
		       1 - (k.embedding <=> $1::vector) AS similarity
		FROM oem_knowledge k
		LEFT JOIN oem_documents d ON d.id = k.document_id
		WHERE 1=1`

	args := []interface{}{vec}
	argIdx := 2

	if make != "" {
		query += fmt.Sprintf(" AND k.controller_make = $%d", argIdx)
		args = append(args, make)
		argIdx++
	}
	if model != "" {
		query += fmt.Sprintf(" AND k.controller_model = $%d", argIdx)
		args = append(args, model)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY k.embedding <=> $1::vector LIMIT $%d", argIdx)
	args = append(args, topK)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(
			&r.ID, &r.DocumentID, &r.ControllerMake, &r.ControllerModel,
			&r.Section, &r.Content, &r.PageNumber, &r.ChunkIndex, &r.CreatedAt,
			&r.DocumentName, &r.Similarity,
		); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *PGXStore) InsertDocument(ctx context.Context, doc Document) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO oem_documents (filename, controller_make, controller_model, page_count, checksum)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		doc.Filename, doc.ControllerMake, doc.ControllerModel, doc.PageCount, doc.Checksum,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert document: %w", err)
	}
	return id, nil
}

func (s *PGXStore) InsertChunks(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Build a multi-row INSERT for simplicity (chunk count is small for seed data).
	query := `INSERT INTO oem_knowledge (document_id, controller_make, controller_model, section, content, page_number, chunk_index, embedding) VALUES `
	args := make([]interface{}, 0, len(chunks)*8)
	placeholders := make([]string, 0, len(chunks))

	for i, c := range chunks {
		base := i*8 + 1
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d::vector)",
			base, base+1, base+2, base+3, base+4, base+5, base+6, base+7,
		))
		var docID interface{} = nil
		if c.DocumentID != "" {
			docID = c.DocumentID
		}
		args = append(args, docID, c.ControllerMake, c.ControllerModel, c.Section, c.Content, c.PageNumber, c.ChunkIndex, vectorLiteral(c.Embedding))
	}

	query += strings.Join(placeholders, ", ")
	if _, err := s.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert chunks: %w", err)
	}
	return nil
}

func (s *PGXStore) ListDocuments(ctx context.Context) ([]Document, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.filename, d.controller_make, d.controller_model, d.page_count,
		       d.ingested_at, d.checksum,
		       (SELECT COUNT(*) FROM oem_knowledge k WHERE k.document_id = d.id) AS chunk_count
		FROM oem_documents d
		ORDER BY d.ingested_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.Filename, &d.ControllerMake, &d.ControllerModel,
			&d.PageCount, &d.IngestedAt, &d.Checksum, &d.ChunkCount); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *PGXStore) ChunkCount(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM oem_knowledge`).Scan(&count)
	return count, err
}
