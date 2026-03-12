package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/url" // FIX: this import was missing from the guide's version

	ollama "github.com/ollama/ollama/api"
)

type Embedder struct {
	client    *ollama.Client
	modelName string // e.g. nomic-embed-text
}

// NewEmbedder creates an Embedder that talks to your local Ollama server
func NewEmbedder(ollamaURL, modelName string) *Embedder {
	return &Embedder{
		client:    ollama.NewClient(mustParseURL(ollamaURL), &http.Client{}),
		modelName: modelName,
	}
}

// Embed converts text into a 768-dimensional vector using nomic-embed-text
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := e.client.Embeddings(ctx, &ollama.EmbeddingRequest{
		Model:  e.modelName,
		Prompt: text,
	})
	if err != nil {
		return nil, err
	}

	// Ollama returns float64 — convert to float32 for pgvector
	out := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		out[i] = float32(v)
	}
	return out, nil
}

// mustParseURL is shared by Embedder and LLM
func mustParseURL(u string) *url.URL {
	parsed, err := url.Parse(u)
	if err != nil {
		panic(fmt.Sprintf("bad Ollama URL: %v", err))
	}
	return parsed
}