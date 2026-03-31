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
	htmlPrompt             string
	multiFilePrompt        string
	vuePrompt              string
	routingPrompt          string
	imageCollectionPrompt  string
	codeQualityCheckPrompt string
	promptOnce             sync.Once
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

		routingPrompt, err = loadPromptFile("codegen-routing-system-prompt.txt")
		if err != nil {
			routingPrompt = htmlPrompt
			err = nil
		}

		imageCollectionPrompt, err = loadPromptFile("image-collection-system-prompt.txt")
		if err != nil {
			imageCollectionPrompt = "你是一个图片收集助手，帮助用户收集和搜索各类图片资源。你可以使用以下工具：\n1. imageSearch - 搜索内容相关的图片\n2. undrawIllustration - 搜索插画图片\n3. logoGenerator - 生成Logo图片\n\n请根据用户的需求选择合适的工具来收集图片，并以JSON数组格式返回结果。"
			err = nil
		}

		codeQualityCheckPrompt, err = loadPromptFile("code-quality-check-system-prompt.txt")
		if err != nil {
			codeQualityCheckPrompt = getDefaultCodeQualityCheckPrompt()
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

func GetRoutingPrompt() string {
	return routingPrompt
}

func GetImageCollectionPrompt() string {
	return imageCollectionPrompt
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

func getDefaultCodeQualityCheckPrompt() string {
	return `你是一个专业的代码质量检查专家。你的任务是分析代码并评估其质量。

请检查以下方面：
1. 代码结构和组织
2. 命名规范
3. 潜在的bug和错误
4. 性能问题
5. 安全隐患
6. 最佳实践遵循情况

请以JSON格式返回检查结果，格式如下：
{
  "is_valid": true/false,
  "errors": ["错误1", "错误2"],
  "suggestions": ["建议1", "建议2"]
}

注意：
- is_valid 表示代码是否通过质量检查
- errors 列出发现的问题
- suggestions 列出改进建议
- 如果代码质量良好，is_valid 为 true，errors 可以为空
- 如果发现严重问题，is_valid 为 false，errors 列出具体问题`
}
