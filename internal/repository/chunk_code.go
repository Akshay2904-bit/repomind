package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
	"github.com/Akshay2904-bit/repomind/internal/model"
)

type ChunkRepository struct {
	db *pgxpool.Pool
}

func NewChunkRepository(db *pgxpool.Pool) *ChunkRepository {
	return &ChunkRepository{db: db}
}

// GetOrCreateRepo inserts the repo name if it doesn't exist, then returns its integer ID.
// Called by IndexRepo before scanning so every chunk has a valid repo_id foreign key.
func (r *ChunkRepository) GetOrCreateRepo(ctx context.Context, name string) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO repositories (name) VALUES ($1)
		 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
		name).Scan(&id)
	return id, err
}

// GetRepoByName returns the integer ID for a repo that was previously indexed.
// Called by AskQuestion to convert the string "myproject" into a repo_id for SimilarChunks.
func (r *ChunkRepository) GetRepoByName(ctx context.Context, name string) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`SELECT id FROM repositories WHERE name = $1`, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("repo not found: %s (did you run /index first?)", name)
	}
	return id, nil
}

// InsertChunk saves a code chunk and its embedding to the database.
// This method signature must exactly match the RepositoryStore interface in service/scanner.go.
func (r *ChunkRepository) InsertChunk(ctx context.Context, c *model.CodeChunk) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO code_chunks
		(repo_id, file_path, content, start_line, end_line, embedding)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		c.RepoID, c.FilePath, c.Content,
		c.StartLine, c.EndLine, pgvector.NewVector(c.Embedding))
	return err
}

// SimilarChunks finds the N most semantically similar chunks to a query embedding.
func (r *ChunkRepository) SimilarChunks(ctx context.Context,
	repoID int, embedding []float32, limit int) ([]*model.CodeChunk, error) {
	rows, err := r.db.Query(ctx,
		`SELECT file_path, content, start_line, end_line
		FROM code_chunks
		WHERE repo_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3`,
		repoID, pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []*model.CodeChunk
	for rows.Next() {
		c := &model.CodeChunk{}
		rows.Scan(&c.FilePath, &c.Content, &c.StartLine, &c.EndLine)
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}