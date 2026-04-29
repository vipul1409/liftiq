package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/liftiq/knowledge-base/internal/embed"
	"github.com/liftiq/knowledge-base/internal/store"
)

// SeedFile is the top-level structure of the JSON seed file.
type SeedFile struct {
	Document SeedDocument `json:"document"`
	Chunks   []SeedChunk  `json:"chunks"`
}

// SeedDocument describes the source document.
type SeedDocument struct {
	Filename        string `json:"filename"`
	ControllerMake  string `json:"controller_make"`
	ControllerModel string `json:"controller_model"`
	PageCount       int    `json:"page_count"`
}

// SeedChunk is a single knowledge chunk in the seed file.
type SeedChunk struct {
	Section    string `json:"section"`
	Content    string `json:"content"`
	PageNumber int    `json:"page_number"`
}

func main() {
	seedPath := flag.String("seed", "", "path to JSON seed file")
	dbURL := flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "Ollama API base URL")
	ollamaModel := flag.String("ollama-model", "nomic-embed-text", "Ollama embedding model")
	flag.Parse()

	if *seedPath == "" {
		fmt.Fprintln(os.Stderr, "usage: ingest --seed <path.json> [--database-url <dsn>]")
		os.Exit(1)
	}
	if *dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required (env or --database-url flag)")
		os.Exit(1)
	}

	ctx := context.Background()

	// Read seed file.
	data, err := os.ReadFile(*seedPath)
	if err != nil {
		log.Fatalf("read seed file: %v", err)
	}

	var seed SeedFile
	if err := json.Unmarshal(data, &seed); err != nil {
		log.Fatalf("parse seed file: %v", err)
	}

	fmt.Printf("Loaded %d chunks from %s\n", len(seed.Chunks), *seedPath)

	// Connect to database.
	pool, err := store.Connect(ctx, *dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	if err := store.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	st := store.NewPGXStore(pool)

	// Check if data already exists.
	count, err := st.ChunkCount(ctx)
	if err != nil {
		log.Fatalf("check chunk count: %v", err)
	}
	if count > 0 {
		fmt.Printf("Knowledge base already has %d chunks. Skipping seed.\n", count)
		return
	}

	// Create document record.
	docID, err := st.InsertDocument(ctx, store.Document{
		Filename:        seed.Document.Filename,
		ControllerMake:  seed.Document.ControllerMake,
		ControllerModel: seed.Document.ControllerModel,
		PageCount:       seed.Document.PageCount,
	})
	if err != nil {
		log.Fatalf("insert document: %v", err)
	}
	fmt.Printf("Created document %s\n", docID)

	// Embed and insert chunks.
	embedder := embed.NewOllamaEmbedder(*ollamaURL, *ollamaModel)

	chunks := make([]store.Chunk, len(seed.Chunks))
	for i, sc := range seed.Chunks {
		fmt.Printf("Embedding chunk %d/%d: %s\n", i+1, len(seed.Chunks), sc.Section)

		vec, err := embedder.Embed(ctx, sc.Content)
		if err != nil {
			log.Fatalf("embed chunk %d: %v", i, err)
		}

		chunks[i] = store.Chunk{
			DocumentID:      docID,
			ControllerMake:  seed.Document.ControllerMake,
			ControllerModel: seed.Document.ControllerModel,
			Section:         sc.Section,
			Content:         sc.Content,
			PageNumber:      sc.PageNumber,
			ChunkIndex:      i,
			Embedding:       vec,
		}
	}

	if err := st.InsertChunks(ctx, chunks); err != nil {
		log.Fatalf("insert chunks: %v", err)
	}

	fmt.Printf("Successfully seeded %d chunks into knowledge base.\n", len(chunks))
}
