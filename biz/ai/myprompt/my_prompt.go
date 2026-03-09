package myprompt

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	path "workspace-yikou-ai-go/pkg/file"
)

var (
	htmlPrompt      string
	multiFilePrompt string
	vuePrompt       string
	promptOnce      sync.Once
)

func loadPromptFile(fileName string) (string, error) {
	projectRoot, err := path.GetProjectRoot()
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(projectRoot, "prompt", fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func LoadPrompts() error {
	var err error
	promptOnce.Do(func() {
		htmlPrompt, err = loadPromptFile("codegen-html-system-prompt.txt")
		if err != nil {
			return
		}

		multiFilePrompt, err = loadPromptFile("codegen-multi-file-system-prompt.txt")
		if err != nil {
			return
		}

		vuePrompt, err = loadPromptFile("codegen-vue-project-system-prompt.txt")
		if err != nil {
			vuePrompt = htmlPrompt
			err = nil
		}
	})
	return err
}

func GetHtmlPrompt() string {
	return htmlPrompt
}

func GetMultiFilePrompt() string {
	return multiFilePrompt
}

func GetVuePrompt() string {
	return vuePrompt
}

func NewMultiFileChatTemplate() (prompt.ChatTemplate, error) {
	return newChatTemplate(GetMultiFilePrompt()), nil
}

func NewHtmlChatTemplate() (prompt.ChatTemplate, error) {
	return newChatTemplate(GetHtmlPrompt()), nil
}

func NewVueProjectPrompt() (prompt.ChatTemplate, error) {
	return newChatTemplate(GetVuePrompt()), nil
}

func newChatTemplate(systemPrompt string) prompt.ChatTemplate {
	ctp := prompt.FromMessages(schema.GoTemplate, []schema.MessagesTemplate{
		schema.SystemMessage(systemPrompt),
		schema.MessagesPlaceholder("history", false),
		schema.UserMessage("{{.content}}"),
	}...)
	return ctp
}
