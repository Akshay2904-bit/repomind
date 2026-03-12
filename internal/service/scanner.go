package service

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Akshay2904-bit/repomind/internal/ai"
	"github.com/Akshay2904-bit/repomind/internal/model"
	"github.com/Akshay2904-bit/repomind/pkg/chunker"
)

// RepositoryStore is the interface ScannerService depends on.
// *repository.ChunkRepository satisfies this automatically because it has InsertChunk.
// Go interfaces are implicit — no "implements" keyword needed.
type RepositoryStore interface {
	InsertChunk(ctx context.Context, c *model.CodeChunk) error
}

type ScannerService struct {
	queue    chan chunker.Chunk
	embedder *ai.Embedder
	repo     RepositoryStore
}

func NewScannerService(embedder *ai.Embedder, repo RepositoryStore) *ScannerService {
	return &ScannerService{
		queue:    make(chan chunker.Chunk, 1000),
		embedder: embedder,
		repo:     repo,
	}
}

// Queue exposes the channel so IndexRepo can close it after ScanRepo finishes.
// Closing the channel is what causes all workers to exit their range loops.
func (s *ScannerService) Queue() chan chunker.Chunk {
	return s.queue
}

// ScanRepo walks the directory tree and sends each file's chunks into the queue.
// Must be called AFTER StartWorkers so workers are already waiting to consume.
func (s *ScannerService) ScanRepo(ctx context.Context, repoPath string) error {
	return filepath.WalkDir(repoPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isSupportedFile(p) {
			return nil
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return nil // skip unreadable files silently
		}
		for _, ch := range chunker.ChunkFile(p, string(content)) {
			s.queue <- ch
		}
		return nil
	})
}

func isSupportedFile(p string) bool {
	supported := map[string]bool{
		".go": true, ".py": true, ".ts": true,
		".js": true, ".java": true, ".rs": true,
		".md": true, ".sql": true, ".sh": true,
	}
	return supported[filepath.Ext(p)]
}