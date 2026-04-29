package store

import (
	"context"
	"time"
)

// Chunk represents a piece of OEM knowledge with its embedding.
type Chunk struct {
	ID              string
	DocumentID      string
	ControllerMake  string
	ControllerModel string
	Section         string
	Content         string
	PageNumber      int
	ChunkIndex      int
	Embedding       []float32
	CreatedAt       time.Time
}

// Document represents an ingested OEM manual.
type Document struct {
	ID              string
	Filename        string
	ControllerMake  string
	ControllerModel string
	PageCount       int
	ChunkCount      int
	IngestedAt      time.Time
	Checksum        string
}

// SearchResult is a chunk with a similarity score.
type SearchResult struct {
	Chunk
	DocumentName string
	Similarity   float64
}

// Store defines the data access layer for the knowledge base.
type Store interface {
	// Search finds the top-k most similar chunks to the given embedding vector.
	// Optional make/model filters narrow the search space.
	Search(ctx context.Context, embedding []float32, make, model string, topK int) ([]SearchResult, error)

	// InsertDocument creates a document record and returns its ID.
	InsertDocument(ctx context.Context, doc Document) (string, error)

	// InsertChunks bulk-inserts knowledge chunks.
	InsertChunks(ctx context.Context, chunks []Chunk) error

	// ListDocuments returns all ingested documents with chunk counts.
	ListDocuments(ctx context.Context) ([]Document, error)

	// ChunkCount returns the total number of chunks in the knowledge base.
	ChunkCount(ctx context.Context) (int, error)
}
