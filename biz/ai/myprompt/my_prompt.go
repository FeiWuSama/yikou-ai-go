package myprompt

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"os"
	"path/filepath"
	path "workspace-yikou-ai-go/pkg/file"
)

func NewMultiFileChatTemplate() (prompt.ChatTemplate, error) {
	projectRoot, err := path.GetProjectRoot()
	if err != nil {
		return nil, err
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-multi-file-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, err
	}
	return newChatTemplate(string(systemPrompt)), nil
}

func NewHtmlChatTemplate() (prompt.ChatTemplate, error) {
	projectRoot, err := path.GetProjectRoot()
	if err != nil {
		return nil, err
	}
	promptPath := filepath.Join(projectRoot, "prompt/codegen-html-system-prompt.txt")
	systemPrompt, err := os.ReadFile(promptPath)
	if err != nil {
		return nil, err
	}
	return newChatTemplate(string(systemPrompt)), nil
}

func newChatTemplate(systemPrompt string) prompt.ChatTemplate {
	ctp := prompt.FromMessages(schema.FString, []schema.MessagesTemplate{
		schema.SystemMessage(systemPrompt),
		schema.MessagesPlaceholder("history", false),
		schema.UserMessage("{content}"),
	}...)
	return ctp
}
