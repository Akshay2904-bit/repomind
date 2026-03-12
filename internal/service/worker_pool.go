package service

import (
	"context"
	"log"
	"sync"
	"github.com/Akshay2904-bit/repomind/internal/model"
)

const NumWorkers = 10 // number of PARALLER GO ROUTINES

// StartWorkers laumches 10 goroutines that process chunks from the queue
func (s *ScannerService) StartWorkers(ctx context.Context, repoID int) {
	var wg sync.WaitGroup

	for i := 0; i < NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range s.queue {
				embedding, err := s.embedder.Embed(ctx, chunk.Content)
				if err != nil {
					log.Printf("Failed to embed chunk %d: %v", chunk.FilePath, err)
					continue
				}
				err = s.repo.InsertChunk(ctx, &model.CodeChunk{
					RepoID:   repoID,
					FilePath: chunk.FilePath,
					Content:  chunk.Content,
					StartLine: chunk.StartLine,
					EndLine:   chunk.EndLine,
					Embedding: embedding,
				})
				if err != nil {
					log.Printf("Failed to insert chunk %d into DB: %v", chunk.FilePath, err)
				}
			}
		}()
	}
	wg.Wait() // wait for all workers to finish
}