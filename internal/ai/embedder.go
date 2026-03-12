package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	ollama "github.com/ollama/ollama/api"
)

type Embedder struct {
	client *ollama.Client
	modelName string 
}

// NewEmbedder initializes the Ollama client and returns an Embedder instance
func NewEmbedder(ollamaURL, modelName string) *Embedder {
	return &Embedder{
		client: ollama.NewClient(mustParseURL(ollamaURL), &http.Client{}),
		modelName: modelName,
	}
}

// Embed converts text into a vector embedding using the Ollama API
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := e.client.Embeddings(ctx, &ollama.EmbeddingRequest{
		Model: e.modelName,
		Prompt: text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}
	//ollama returns float64 -  convert to float 32 for pgvector
	out := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		out[i] = float32(v)
	}
	return out, nil
}

func mustParseURL(u string) *url.URL {
    parsed, err := url.Parse(u)
    if err != nil { panic(fmt.Sprintf("bad Ollama URL: %v", err)) }
    return parsed
}
