package enum

type CodeGenTypeEnum string

const (
	HtmlCodeGen  CodeGenTypeEnum = "html"
	MultiFileGen CodeGenTypeEnum = "multi_file"
)

var CodeGenTypeTextMap = map[CodeGenTypeEnum]string{
	HtmlCodeGen:  "原生 HTML 模式",
	MultiFileGen: "原生多文件模式",
}
