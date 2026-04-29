package llm

import "context"

// Generator produces natural-language answers from a question and context chunks.
type Generator interface {
	// Generate synthesizes an answer using the given context passages.
	Generate(ctx context.Context, question string, chunks []string) (string, error)
}
