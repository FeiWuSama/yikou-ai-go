package commonenum

type CodeGenTypeEnum string

const (
	HtmlCodeGen  CodeGenTypeEnum = "html"
	MultiFileGen CodeGenTypeEnum = "multi_file"
	VueCodeGen   CodeGenTypeEnum = "vue_project"
)

var CodeGenTypeTextMap = map[CodeGenTypeEnum]string{
	HtmlCodeGen:  "原生 HTML 模式",
	MultiFileGen: "原生多文件模式",
	VueCodeGen:   "Vue工厂模式",
}
