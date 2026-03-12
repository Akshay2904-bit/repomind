package ai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	ollama "github.com/ollama/ollama/api"
	"github.com/Akshay2904-bit/repomind/internal/model"
)


type LLM struct {
	client *ollama.Client
	modelName string
}

func NewLLM(ollamaURL, modelName string) *LLM {
	return &LLM{
		client: ollama.NewClient(mustParseURL(ollamaURL), &http.Client{}),
		modelName: modelName,
	}
}

func (l *LLM) Answer(ctx context.Context, question string, chunks []*model.CodeChunk) (string, error) {

	//build context block from the retrieved code chunks
	var ctxBlock strings.Builder
	for _, chunk := range chunks {
		fmt.Fprintf(
			&ctxBlock,
            "### File: %s (lines %d-%d)\n%s\n\n",
            chunk.FilePath, chunk.StartLine, chunk.EndLine, chunk.Content)
	}

	prompt := fmt.Sprintf(`You are an expert code assistant.
	Answer ONLY using the code context below.
	Reference specific file paths and line numbers in your answer.
	If the answer is not in the context, say 'I could not find that in the indexed code.'

	Context:
	%s
	Question: %s`, ctxBlock.String(), question)

	//stream=false means we wait for the complete response 
	var answer strings.Builder
	streamFalse := false
	err := l.client.Generate(ctx, &ollama.GenerateRequest{
		Model: l.modelName,
		Prompt: prompt,
		Stream: &streamFalse,
	}, func(resp ollama.GenerateResponse) error {
		answer.WriteString(resp.Response)
		return nil
	})
	if err != nil {
		return "", err
	}
	return answer.String(), nil
}