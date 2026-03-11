package model

import "time"

type repository struct {
	ID         int       `db:"id"`
	Name       string    `db:"name"`
	URL        string    `db:"url"`
	CreatedAt  time.Time `db:"created_at"`
}

type code_chunk struct {
	ID         int       `db:"id"`
	RepoID     int       `db:"repo_id"`
	Content    string    `db:"content"`
	StartLine  int       `db:"start_line"`
	EndLine    int       `db:"end_line"`
	Embedding  []float32 `db:"embedding"`
	CreatedAt  time.Time `db:"created_at"`
}