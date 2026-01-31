package enum

type CodeGenType string

const (
	HtmlCodeGen  CodeGenType = "html"
	MultiFileGen CodeGenType = "multi_file"
)

var CodeGenTypeTextMap = map[CodeGenType]string{
	HtmlCodeGen:  "原生 HTML 模式",
	MultiFileGen: "原生多文件模式",
}
