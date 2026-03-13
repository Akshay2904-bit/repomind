package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/Akshay2904-bit/repomind/internal/ai"
	"github.com/Akshay2904-bit/repomind/internal/repository"
	"github.com/Akshay2904-bit/repomind/internal/service"
)

// Handler holds all the dependencies needed by HTTP endpoints.
type Handler struct {
	embedder  *ai.Embedder
	llm       *ai.LLM
	chunkRepo *repository.ChunkRepository
	askSvc    *service.AskService
}

// NewHandler wires all dependencies together.
func NewHandler(
	embedder *ai.Embedder,
	llm *ai.LLM,
	chunkRepo *repository.ChunkRepository,
	askSvc *service.AskService,
) *Handler {
	return &Handler{
		embedder:  embedder,
		llm:       llm,
		chunkRepo: chunkRepo,
		askSvc:    askSvc,
	}
}

// IndexRequest is the JSON body for POST /repos/{repo}/index
type IndexRequest struct {
	Path string `json:"path"`
}

// IndexRepo handles POST /repos/{repo}/index
func (h *Handler) IndexRepo(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")

	var req IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		http.Error(w, `missing "path" field in request body`, http.StatusBadRequest)
		return
	}

	// GetOrCreateRepo — NOT GetRepoIDByName — is the correct method name.
	// It upserts the repo row and returns the integer primary key.
	repoID, err := h.chunkRepo.GetOrCreateRepo(r.Context(), repoName)
	if err != nil {
		http.Error(w, "db error creating repo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Run indexing in the background so HTTP responds immediately.
	// chunkRepo satisfies service.RepositoryStore because it has InsertChunk —
	// Go interfaces are implicit, no cast or wrapper needed.
	go func() {
    ctx := context.Background() // ← use this, not r.Context()
    scanner := service.NewScannerService(h.embedder, h.chunkRepo)
    scanner.StartWorkers(ctx, repoID)
    if err := scanner.ScanRepo(ctx, req.Path); err != nil {
        log.Printf("scan error for %s: %v", repoName, err)
    }
    close(scanner.Queue())
    log.Printf("indexing complete: %s", repoName)
}()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"indexing started"}`))
}

// ListFiles handles GET /repos/{repo}/files
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"files":[]}`))
}