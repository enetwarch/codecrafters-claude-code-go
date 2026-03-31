package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()
	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}
	ctx := context.Background()

	messages := []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)}
	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Read",
			Description: openai.String("Read and return the contents of a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"file_path"},
			},
		}),
	}

	var resp *openai.ChatCompletion
	var err error
	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL))
	resp, err = client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    "anthropic/claude-haiku-4.5",
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		panic(err)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	for {
		assistantMessage := resp.Choices[0].Message
		messages = append(messages, assistantMessage.ToParam())

		if len(assistantMessage.ToolCalls) > 0 {
			toolCall := assistantMessage.ToolCalls[0]
			if toolCall.Function.Name == "Read" {
				var args struct {
					FilePath string `json:"file_path"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				content, err := os.ReadFile(args.FilePath)
				if err != nil {
					panic(err)
				}
				messages = append(messages, openai.ToolMessage(string(content), toolCall.ID))
			}

			resp, err = client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
				Model:    "anthropic/claude-haiku-4.5",
				Messages: messages,
				Tools:    tools,
			})
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Print(assistantMessage.Content)
			break
		}
	}
}
