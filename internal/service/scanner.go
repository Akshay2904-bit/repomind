package service

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"github.com/Akshay2904-bit/repomind/internal/model"
	"github.com/Akshay2904-bit/repomind/internal/ai"
	"github.com/Akshay2904-bit/repomind/pkg/chunker"
)

// RepositoryStore is an interface — any type that has InsertChunk works here.
// Your *repository.ChunkRepository already satisfies this automatically
// because it has an InsertChunk method with this exact signature.
type RepositoryStore interface {
    InsertChunk(ctx context.Context, c *model.CodeChunk) error
}

type ScannerService struct {
    queue    chan chunker.Chunk  // channel that feeds the worker pool
    embedder *ai.Embedder       // the Ollama embedder from internal/ai/embedder.go
    repo     RepositoryStore    // accepts *repository.ChunkRepository
}

func NewScannerService(embedder *ai.Embedder, repo RepositoryStore) *ScannerService {
	return &ScannerService{
		queue:    make(chan chunker.Chunk, 1000), //buffered channel can be used if needed
		embedder: embedder,
		repo:     repo,
	}
}

// Scan Repo walk every file in the repo directory and ques chunks for embedding
func (s *ScannerService) ScanRepo(ctx context.Context, repoPath string) error {
	return filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isSupportedFile(path) {
			return nil // skip directories
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		chunks := chunker.ChunkFile(path, string(content))
		for _, ch := range chunks {
			s.queue <- ch // send chunk to the worker pool
		}
		return nil
	})
}

// isSupportedFile returns true if the file is of a supported type we want to index
func isSupportedFile(path string) bool{
	    supported := map[string]bool{
        ".go": true, ".py": true, ".ts": true,
        ".js": true, ".java": true, ".rs": true,
        ".md": true, ".sql": true, ".sh": true,
    }

	return supported[filepath.Ext(path)]
}