package ai

type HtmlCodeResponse struct {
	HtmlCode    string `json:"html_code"`
	Description string `json:"description"`
}

type MultiFileCodeResponse struct {
	HtmlCodeResponse
	JsCode  string `json:"js_code"`
	CssCode string `json:"css_code"`
}
