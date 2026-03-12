package service

import (
	"context"
	"log"
	"sync"

	"github.com/Akshay2904-bit/repomind/internal/model"
)

const NumWorkers = 10 // number of parallel goroutines

// StartWorkers launches NumWorkers goroutines that process chunks from the queue.
// IMPORTANT: Call this BEFORE ScanRepo, then close(scanner.Queue()) after ScanRepo returns.
// The workers range over s.queue — they exit automatically when the channel is closed.
func (s *ScannerService) StartWorkers(ctx context.Context, repoID int) {
	var wg sync.WaitGroup

	for i := 0; i < NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range s.queue { // blocks until a chunk arrives; exits when channel is closed
				embedding, err := s.embedder.Embed(ctx, chunk.Content)
				if err != nil {
					log.Printf("embed error for %s: %v", chunk.FilePath, err)
					continue
				}
				err = s.repo.InsertChunk(ctx, &model.CodeChunk{
					RepoID:    repoID,
					FilePath:  chunk.FilePath,
					Content:   chunk.Content,
					StartLine: chunk.StartLine,
					EndLine:   chunk.EndLine,
					Embedding: embedding,
				})
				if err != nil {
					log.Printf("db insert error: %v", err)
				}
			}
		}()
	}

	// NOTE: wg.Wait() is intentionally NOT called here.
	// StartWorkers is called from a goroutine in IndexRepo.
	// The goroutine closes the channel after ScanRepo finishes,
	// which causes all workers to exit their range loops naturally.
	// If you need to block until done, add wg.Wait() and call close() before this.
}