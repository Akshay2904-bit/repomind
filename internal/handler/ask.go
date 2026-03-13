package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Akshay2904-bit/repomind/internal/model"
)

type AskRequest struct {
	Repo     string `json:"repo"`
	Question string `json:"question"`
}

type AskResponse struct {
	Answer string   `json:"answer"`
	Files  []string `json:"files"`
}

// AskQuestion handles POST /ask
func (h *Handler) AskQuestion(w http.ResponseWriter, r *http.Request) {
	var req AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert repo name string → integer ID.
	// If /index was never called (or is still running), this returns a clear error.
	repoID, err := h.chunkRepo.GetRepoByName(r.Context(), req.Repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	qEmbed, err := h.embedder.Embed(r.Context(), req.Question)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chunks, err := h.chunkRepo.SimilarChunks(r.Context(), repoID, qEmbed, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	answer, err := h.askSvc.AskWithCache(r.Context(), req.Repo, req.Question)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AskResponse{
		Answer: answer,
		Files:  uniqueFiles(chunks),
	})
}

func uniqueFiles(chunks []*model.CodeChunk) []string {
	seen := map[string]bool{}
	var files []string
	for _, c := range chunks {
		if !seen[c.FilePath] {
			seen[c.FilePath] = true
			files = append(files, c.FilePath)
		}
	}
	return files
}