package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Akshay2904-bit/repomind/internal/ai"
	"github.com/Akshay2904-bit/repomind/internal/repository"
	"github.com/Akshay2904-bit/repomind/internal/service"
	"github.com/go-chi/chi/v5"
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
	URL  string `json:"url"`
}

func (h *Handler) IndexRepo(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")

	var req IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.URL == "" && req.Path == "" {
		http.Error(w, `provide either "url" (GitHub URL) or "path" (local path)`,
			http.StatusBadRequest)
		return
	}

	repoID, err := h.chunkRepo.GetOrCreateRepo(r.Context(), repoName)
	if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		ctx := context.Background() // REQUIRED — r.Context() cancels when HTTP returns (Bug 3/4 fix)
		scanPath := req.Path

		// If a GitHub URL was supplied, clone it first
		if service.IsGitHubURL(req.URL) {
			cloneResult, err := service.CloneGitHubRepo(req.URL)
			if err != nil {
				log.Printf("clone error for %s: %v", req.URL, err)
				return
			}
			defer cloneResult.Cleanup() // delete temp dir when done
			scanPath = cloneResult.Dir
		}

		scanner := service.NewScannerService(h.embedder, h.chunkRepo)
		scanner.StartWorkers(ctx, repoID)
		if err := scanner.ScanRepo(ctx, scanPath); err != nil {
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
