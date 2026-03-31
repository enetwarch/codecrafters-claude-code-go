package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

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

	model := "anthropic/claude-haiku-4.5"
	messages := []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)}
	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Read",
			Description: openai.String("Read and return the contents of a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]string{
						"type":        "string",
						"description": "The path of the file to read",
					},
				},
				"required": []string{"file_path"},
			},
		}),
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Write",
			Description: openai.String("Write content to a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]string{
						"type":        "string",
						"description": "The path of the file to write to",
					},
					"content": map[string]string{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				"required": []string{"file_path", "content"},
			},
		}),
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Bash",
			Description: openai.String("Execute a shell command"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]string{
						"type":        "string",
						"description": "The command to execute",
					},
				},
				"required": []string{"command"},
			},
		}),
	}

	var resp *openai.ChatCompletion
	var err error
	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL))
	resp, err = client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{Model: model, Messages: messages, Tools: tools})
	if err != nil {
		panic(err)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	for {
		assistantMsg := resp.Choices[0].Message
		messages = append(messages, assistantMsg.ToParam())

		if len(assistantMsg.ToolCalls) > 0 {
			toolCall := assistantMsg.ToolCalls[0]
			switch toolCall.Function.Name {
			case "Read":
				var args struct {
					FilePath string `json:"file_path"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				content, err := os.ReadFile(args.FilePath)
				if err != nil {
					panic(err)
				}
				messages = append(messages, openai.ToolMessage(string(content), toolCall.ID))

			case "Write":
				var args struct {
					FilePath string `json:"file_path"`
					Content  string `json:"content"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				err = os.WriteFile(args.FilePath, []byte(args.Content), 0666)
				if err != nil {
					panic(err)
				}
				messages = append(messages, openai.ToolMessage("Writing successful!", toolCall.ID))

			case "Bash":
				var args struct {
					Command string `json:"command"`
				}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				cmd := exec.Command("sh", "-c", args.Command)
				output, err := cmd.CombinedOutput()
				if err != nil {
					panic(err)
				}
				messages = append(messages, openai.ToolMessage(string(output), toolCall.ID))
			}

			resp, err = client.Chat.Completions.New(ctx,
				openai.ChatCompletionNewParams{Model: model, Messages: messages, Tools: tools})
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Print(assistantMsg.Content)
			break
		}
	}
}
