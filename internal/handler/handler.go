package handler

import (
	"net/http"

	"github.com/Akshay2904-bit/repomind/internal/ai"
	"github.com/Akshay2904-bit/repomind/internal/repository"
)

// Handler holds all the dependencies needed by HTTP endpoints.
// Every handler method (AskQuestion, IndexRepo, ListFiles) is defined
// as methods on this struct so they share the same embedder, LLM, and DB.
type Handler struct {
    embedder  *ai.Embedder
    llm       *ai.LLM
    chunkRepo *repository.ChunkRepository
}
 
// NewHandler is the constructor — called once in main.go to wire everything together.
// main.go creates the individual dependencies (embedder, llm, chunkRepo),
// then passes them here so Handler can use them in every request.
func NewHandler(
    embedder  *ai.Embedder,
    llm       *ai.LLM,
    chunkRepo *repository.ChunkRepository,
) *Handler {
    return &Handler{
        embedder:  embedder,
        llm:       llm,
        chunkRepo: chunkRepo,
    }
}
 
// IndexRepo handles POST /repos/{repo}/index
// (stub — implement to trigger ScannerService.ScanRepo)
func (h *Handler) IndexRepo(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"status":"indexing started"}`))
}
 
// ListFiles handles GET /repos/{repo}/files
// (stub — implement to query distinct file_path values from code_chunks)
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`{"files":[]}`))
}
