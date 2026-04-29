package embed

import "context"

// Embedder generates vector embeddings from text.
type Embedder interface {
	// Embed returns a dense vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// Dimension returns the output vector dimension.
	Dimension() int
}
