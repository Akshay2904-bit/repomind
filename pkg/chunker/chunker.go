package chunker

import (
	"strings"
)

const (
	ChunkSize = 60
	ChunkOverlap = 10
)

type Chunk struct {
	FilePath  string
	Content   string
	StartLine int
	EndLine   int
}

//ChunkFile slips overlaping soruce file into multiple chunks
func ChunkFile(path, content string) []Chunk {
	lines := strings.Split(content, "\n")
	var chunks []Chunk
	
	for i := 0; i < len(lines); i += ChunkSize - ChunkOverlap {
		end := i + ChunkSize
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, Chunk{
			FilePath:  path,
			Content:   strings.Join(lines[i:end], "\n"),
			StartLine: i + 1,
			EndLine:   end,
		})
	if end == len(lines) {
			break
		}
	}
	return chunks
 }