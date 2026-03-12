package handler

import (
	"encoding/json"
	"net/http"
	"github.com/Akshay2904-bit/repomind/internal/model"
)

type AskRequest struct {
	Repo string `json:"repo"`
	Question string `json:"question"`
}

type AskResponse struct {
	Answer string `json:"answer"`
	Files []string `json:"files"`
}

func (h *Handler) AskQuestion(w http.ResponseWriter, r *http.Request) {
    var req AskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	//step 1: embed the user's quetion into a vector
	qEmbed, err := h.embedder.Embed(r.Context(), req.Question)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//step 2: search for the 5 most aimialr code chunks in the database
	chunks, err := h.chunkrepo.SimilarChunks(r.Context(), req.Repo, qEmbed, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 3: send chunks + question to local Ollama LLM for an answer
    answer, err := h.llm.Answer(r.Context(), req.Question, chunks)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AskResponse{
		Answer: answer,
		Files: uniqueFiles(chunks),
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