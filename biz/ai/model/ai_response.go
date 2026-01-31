package ai

import (
	"regexp"
	"strings"
)

type HtmlCodeResponse struct {
	HtmlCode    string `json:"html_code"`
	Description string `json:"description"`
}

type MultiFileCodeResponse struct {
	HtmlCodeResponse
	JsCode  string `json:"js_code"`
	CssCode string `json:"css_code"`
}

var codeBlockRegex = regexp.MustCompile("```(\\w*)\\n([\\s\\S]*?)```")

type ResponseBlock struct {
	Tag  string
	Code string
}

func parseToResponseBlock(content string) []ResponseBlock {
	matches := codeBlockRegex.FindAllStringSubmatch(content, -1)
	blocks := make([]ResponseBlock, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 3 {
			language := strings.ToLower(strings.TrimSpace(match[1]))
			code := strings.TrimSpace(match[2])
			blocks = append(blocks, ResponseBlock{
				Tag:  language,
				Code: code,
			})
		}
	}
	return blocks
}

func extractDescription(content string) string {
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

func ParseHtmlCodeResponse(content string) (*HtmlCodeResponse, error) {
	blocks := parseToResponseBlock(content)
	response := &HtmlCodeResponse{}
	for _, block := range blocks {
		switch block.Tag {
		case "html":
			response.HtmlCode = block.Code
		case "description":
			response.Description = block.Code
		}
	}
	if response.Description == "" {
		response.Description = extractDescription(content)
	}
	return response, nil
}

func ParseMultiFileCodeResponse(content string) (*MultiFileCodeResponse, error) {
	blocks := parseToResponseBlock(content)
	response := &MultiFileCodeResponse{}
	for _, block := range blocks {
		switch block.Tag {
		case "html":
			response.HtmlCode = block.Code
		case "css":
			response.CssCode = block.Code
		case "javascript", "js":
			response.JsCode = block.Code
		case "description":
			response.Description = block.Code
		}
	}
	if response.Description == "" {
		response.Description = extractDescription(content)
	}
	return response, nil
}
