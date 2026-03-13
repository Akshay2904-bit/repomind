package ai

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Akshay2904-bit/repomind/internal/model"
	ollama "github.com/ollama/ollama/api"
)

type LLM struct {
	client    *ollama.Client
	modelName string
}

func NewLLM(ollamaURL, modelName string) *LLM {
	return &LLM{
		client:    ollama.NewClient(mustParseURL(ollamaURL), &http.Client{}),
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
	Answer using the code context below.
	Reference specific file paths and line numbers in your answer.
	give a descriptive answer, not just a code snippet. If the question cannot be answered make some logical assumptions based on the context
	and answer accordingly.
	give reasoning for your answer based on the provided code context.
	for eg "the reason  believe this is because of the function defined in file X at line Y which does Z"
	If the question is not related to the code, say "I can only answer questions related to the provided code context.
	Answer in points if there are multiple reasons for your answer, and provide a summary at the end.
	remove unnecessary noise from the final answer any symbols etc
	"

	Context:
	%s
	Question: %s`, ctxBlock.String(), question)

	//stream=false means we wait for the complete response
	var answer strings.Builder
	streamFalse := false
	err := l.client.Generate(ctx, &ollama.GenerateRequest{
		Model:  l.modelName,
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
