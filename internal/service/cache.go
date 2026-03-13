package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Akshay2904-bit/repomind/internal/ai"
	"github.com/Akshay2904-bit/repomind/internal/repository"
)

type AskService struct {
	redis     *redis.Client
	embedder  *ai.Embedder
	chunkRepo *repository.ChunkRepository
	llm       *ai.LLM
}

func NewAskService(
	redis *redis.Client,
	embedder *ai.Embedder,
	chunkRepo *repository.ChunkRepository,
	llm *ai.LLM,
) *AskService {
	return &AskService{
		redis:     redis,
		embedder:  embedder,
		chunkRepo: chunkRepo,
		llm:       llm,
	}
}

func (s *AskService) AskWithCache(ctx context.Context,
	repo, question string) (string, error) {

	// Create a unique cache key from repo + question
	key := "ask:" + repo + ":" + hashQuestion(question)

	// Try cache first
	cached, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		fmt.Println("cache hit!")
		return cached, nil
	} else if err != redis.Nil {
		fmt.Printf("redis error: %v\n", err)
	}

	// Cache miss: run the full RAG pipeline
	answer, err := s.ragAnswer(ctx, repo, question)
	if err != nil {
		return "", err
	}

	// Store in cache for 1 hour
	s.redis.Set(ctx, key, answer, time.Hour)

	return answer, nil
}

// ragAnswer runs the full embed → search → LLM pipeline
func (s *AskService) ragAnswer(ctx context.Context,
	repo, question string) (string, error) {

	// Step 1: get the integer repo ID
	repoID, err := s.chunkRepo.GetRepoByName(ctx, repo)
	if err != nil {
		return "", err
	}

	// Step 2: embed the question
	qEmbed, err := s.embedder.Embed(ctx, question)
	if err != nil {
		return "", err
	}

	// Step 3: find the 5 most similar chunks
	chunks, err := s.chunkRepo.SimilarChunks(ctx, repoID, qEmbed, 5)
	if err != nil {
		return "", err
	}

	// Step 4: generate answer with Ollama
	return s.llm.Answer(ctx, question, chunks)
}

// hashQuestion creates a short hex hash of the question for use as a cache key
func hashQuestion(q string) string {
	h := sha256.Sum256([]byte(q))
	return fmt.Sprintf("%x", h[:8]) // ✅ %x not bare %
}