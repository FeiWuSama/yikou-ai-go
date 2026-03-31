package myprompt

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	path "workspace-yikou-ai-go/pkg/myfile"
)

var (
	htmlPrompt                string
	multiFilePrompt           string
	vuePrompt                 string
	routingPrompt             string
	imageCollectionPrompt     string
	imageCollectionPlanPrompt string
	codeQualityCheckPrompt    string
	promptOnce                sync.Once
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
			panic(err)
		}

		multiFilePrompt, err = loadPromptFile("codegen-multi-file-system-prompt.txt")
		if err != nil {
			panic(err)
		}

		vuePrompt, err = loadPromptFile("codegen-vue-project-system-prompt.txt")
		if err != nil {
			panic(err)
		}

		routingPrompt, err = loadPromptFile("codegen-routing-system-prompt.txt")
		if err != nil {
			panic(err)
		}

		imageCollectionPrompt, err = loadPromptFile("image-collection-system-prompt.txt")
		if err != nil {
			panic(err)
		}

		imageCollectionPlanPrompt, err = loadPromptFile("image-collection-plan-system-prompt.txt")
		if err != nil {
			panic(err)
		}

		codeQualityCheckPrompt, err = loadPromptFile("code-quality-check-system-prompt.txt")
		if err != nil {
			panic(err)
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

func GetRoutingPrompt() string {
	return routingPrompt
}

func GetImageCollectionPrompt() string {
	return imageCollectionPrompt
}

func GetImageCollectionPlanPrompt() string {
	return imageCollectionPlanPrompt
}

func GetCodeQualityCheckPrompt() string {
	return codeQualityCheckPrompt
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

func NewRoutingChatTemplate() (prompt.ChatTemplate, error) {
	return newChatTemplate(GetRoutingPrompt()), nil
}

func NewImageCollectionChatTemplate() (prompt.ChatTemplate, error) {
	return newChatTemplate(GetImageCollectionPrompt()), nil
}

func NewImageCollectionPlanChatTemplate() (prompt.ChatTemplate, error) {
	return newChatTemplate(GetImageCollectionPlanPrompt()), nil
}

func NewCodeQualityCheckChatTemplate() (prompt.ChatTemplate, error) {
	return newChatTemplate(GetCodeQualityCheckPrompt()), nil
}

func newChatTemplate(systemPrompt string) prompt.ChatTemplate {
	ctp := prompt.FromMessages(schema.GoTemplate, []schema.MessagesTemplate{
		schema.SystemMessage(systemPrompt),
		schema.MessagesPlaceholder("history", false),
		schema.UserMessage("{{.content}}"),
	}...)
	return ctp
}
