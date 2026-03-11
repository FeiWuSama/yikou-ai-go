package parser

import (
	"fmt"
	"regexp"
	"strings"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/model/enum"
)

type Parser[T any] interface {
	Parse(content string) (T, error)
}

var (
	htmlCodeRegex    = regexp.MustCompile("```html\\n([\\s\\S]*?)```")
	cssCodeRegex     = regexp.MustCompile("```css\\n([\\s\\S]*?)```")
	jsCodeRegex      = regexp.MustCompile("```(javascript|js)\\n([\\s\\S]*?)```")
	descriptionRegex = regexp.MustCompile("```description\\n([\\s\\S]*?)```")
	codeBlockRegex   = regexp.MustCompile("```\\w*\\n[\\s\\S]*?```")
)

type HtmlCodeParser struct{}

func NewHtmlCodeParser() *HtmlCodeParser {
	return &HtmlCodeParser{}
}

func (p *HtmlCodeParser) Parse(content string) (*ai.HtmlCodeResponse, error) {
	result := &ai.HtmlCodeResponse{}

	matches := htmlCodeRegex.FindStringSubmatch(content)
	if len(matches) >= 2 {
		result.HtmlCode = strings.TrimSpace(matches[1])
	}

	description := p.extractDescription(content)
	result.Description = description

	return result, nil
}

func (p *HtmlCodeParser) extractDescription(content string) string {
	matches := descriptionRegex.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}

	cleaned := codeBlockRegex.ReplaceAllString(content, "")
	lines := strings.Split(cleaned, "\n")
	var descriptionLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			descriptionLines = append(descriptionLines, trimmed)
		}
	}
	return strings.Join(descriptionLines, "\n")
}

type MultiFileCodeParser struct{}

func NewMultiFileCodeParser() *MultiFileCodeParser {
	return &MultiFileCodeParser{}
}

func (p *MultiFileCodeParser) Parse(content string) (*ai.MultiFileCodeResponse, error) {
	result := &ai.MultiFileCodeResponse{}

	htmlMatches := htmlCodeRegex.FindStringSubmatch(content)
	if len(htmlMatches) >= 2 {
		result.HtmlCode = strings.TrimSpace(htmlMatches[1])
	}

	cssMatches := cssCodeRegex.FindStringSubmatch(content)
	if len(cssMatches) >= 2 {
		result.CssCode = strings.TrimSpace(cssMatches[1])
	}

	jsMatches := jsCodeRegex.FindStringSubmatch(content)
	if len(jsMatches) >= 3 {
		result.JsCode = strings.TrimSpace(jsMatches[2])
	}

	description := p.extractDescription(content)
	result.Description = description

	return result, nil
}

func (p *MultiFileCodeParser) extractDescription(content string) string {
	matches := descriptionRegex.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}

	cleaned := codeBlockRegex.ReplaceAllString(content, "")
	lines := strings.Split(cleaned, "\n")
	var descriptionLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			descriptionLines = append(descriptionLines, trimmed)
		}
	}
	return strings.Join(descriptionLines, "\n")
}

type CodeParserExecutor struct {
	htmlCodeParser      *HtmlCodeParser
	multiFileCodeParser *MultiFileCodeParser
}

func NewCodeParserExecutor() *CodeParserExecutor {
	return &CodeParserExecutor{
		htmlCodeParser:      NewHtmlCodeParser(),
		multiFileCodeParser: NewMultiFileCodeParser(),
	}
}

func (e *CodeParserExecutor) ExecuteParser(content string, parserType enum.CodeGenTypeEnum) (interface{}, error) {
	switch parserType {
	case enum.HtmlCodeGen:
		return e.htmlCodeParser.Parse(content)
	case enum.MultiFileGen:
		return e.multiFileCodeParser.Parse(content)
	default:
		return nil, fmt.Errorf("不支持的解析类型: %s", parserType)
	}
}
