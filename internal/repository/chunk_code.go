package repository

import (
	"context"
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

// InsterChunk saves a code chunk and its embedded to the database
func (r *ChunkRepository) InsertChunk(ctx context.Context, chunk model.CodeChunk) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO code_chunk (repo_id, file_path, content, stat_line, end_line, embedding)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		chunk.RepoID, chunk.FilePath, chunk.Content, chunk.StartLine, chunk.EndLine, pgvector.NewVector(chunk.Embedding))
	return err
}

// SimilarChunks finds the N most semantically similar chunks to a query embedding
func (r *ChunkRepository) SimilarChunks(ctx context.Context,
    repoID int, embedding []float32, limit int) ([]*model.CodeChunk, error) {
    rows, err := r.db.Query(ctx,
        `SELECT file_path, content, start_line, end_line
         FROM code_chunks
         WHERE repo_id = $1
         ORDER BY embedding <=> $2   -- <=> is cosine distance
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

