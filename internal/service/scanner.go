package service

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"github.com/akshay_2029-bit/repomind/pkg/chunker"
)

type ScannerService struct {
    queue    chan chunker.Chunk   // channel that feeds the worker pool
    embedder *EmbedderService
    repo     RepositoryStore
}
