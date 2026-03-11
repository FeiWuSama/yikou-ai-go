package saver

import (
	"fmt"
	"github.com/sony/sonyflake"
	"os"
	"path/filepath"
	"strconv"
	ai "workspace-yikou-ai-go/biz/ai/aimodel"
	"workspace-yikou-ai-go/biz/model/enum"
	pkg "workspace-yikou-ai-go/pkg/file"
)

var (
	FileSaveDir string
)

func init() {
	rootDir, err := pkg.GetProjectRoot()
	if err != nil {
		panic(err)
	}
	// 拼接文件保存目录
	FileSaveDir = filepath.Join(rootDir, "tmp", "code_output")
}

// buildUniqueDir 构建唯一的目录名
// 目录名格式: {代码生成类型}_{唯一ID}
// Deprecated: 该函数已被废弃
func buildUniqueDir(typeStr enum.CodeGenTypeEnum) (string, error) {
	// 生成雪花id
	var sf = sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) { return 1, nil },
	})
	id, err := sf.NextID()
	if err != nil {
		return "", err
	}
	// 构建唯一目录名
	uniqueDirName := fmt.Sprintf("%s_%s", typeStr, strconv.FormatUint(id, 20))
	dirPath := filepath.Join(FileSaveDir, uniqueDirName)
	// 创建目录
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return dirPath, nil
}

// writeToFile 将内容写入文件并保存
// Deprecated: 该函数已被废弃
func writeToFile(dirPath string, fileName string, content string) error {
	filePath := filepath.Join(dirPath, fileName)
	err := os.WriteFile(filePath, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// SaveHtmlCode 保存 HTML 代码文件
// Deprecated: 该函数已被废弃
func SaveHtmlCode(response ai.HtmlCodeResponse) (string, error) {
	dirPath, err := buildUniqueDir(enum.HtmlCodeGen)
	if err != nil {
		return "", err
	}
	fileName := "index.html"
	return dirPath, writeToFile(dirPath, fileName, response.HtmlCode)
}

// SaveMutiFileCode 保存多文件代码文件
// Deprecated: 该函数已被废弃
func SaveMutiFileCode(response ai.MultiFileCodeResponse) (string, error) {
	dirPath, err := buildUniqueDir(enum.MultiFileGen)
	if err != nil {
		return "", err
	}
	// 保存 HTML 文件
	err = writeToFile(dirPath, "index.html", response.HtmlCode)
	if err != nil {
		return "", err
	}
	// 保存 JS 文件
	err = writeToFile(dirPath, "script.js", response.JsCode)
	if err != nil {
		return "", err
	}
	// 保存 CSS 文件
	err = writeToFile(dirPath, "style.css", response.CssCode)
	if err != nil {
		return "", err
	}
	return dirPath, nil
}

type CodeFileSaverTemplate[T any] interface {
	getCodeType() enum.CodeGenTypeEnum
	saveFiles(response T, baseDir string) error
	validateInput(response T) error
}

type DefaultCodeFileSaver[T any] struct {
}

func (d *DefaultCodeFileSaver[T]) saveCode(
	response T,
	appId int64,
	getCodeType func() enum.CodeGenTypeEnum,
	saveFiles func(response T, baseDir string) error,
	validateInput func(response T) error,
) (string, error) {
	err := validateInput(response)
	if err != nil {
		return "", err
	}
	dirPath, err := d.buildUniqueDir(getCodeType, appId)
	if err != nil {
		return "", err
	}
	return dirPath, saveFiles(response, dirPath)
}

// buildUniqueDir 构建唯一的目录名
// 目录名格式: {代码生成类型}_{唯一ID}
func (d *DefaultCodeFileSaver[T]) buildUniqueDir(getCodeType func() enum.CodeGenTypeEnum, appId int64) (string, error) {
	//// 生成雪花id
	//var sf = sonyflake.NewSonyflake(sonyflake.Settings{
	//	MachineID: func() (uint16, error) { return 1, nil },
	//})
	//id, err := sf.NextID()
	//if err != nil {
	//	return "", err
	//}
	// 构建唯一目录名
	//uniqueDirName := fmt.Sprintf("%s_%s", getCodeType(), strconv.FormatUint(id, 20))
	uniqueDirName := fmt.Sprintf("%s_%s", getCodeType(), strconv.Itoa(int(appId)))
	dirPath := filepath.Join(FileSaveDir, uniqueDirName)
	// 创建目录
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return dirPath, nil
}

// writeToFile 将内容写入文件并保存
func (d *DefaultCodeFileSaver[T]) writeToFile(dirPath string, fileName string, content string) error {
	filePath := filepath.Join(dirPath, fileName)
	err := os.WriteFile(filePath, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

type HtmlCodeFileSaverTemplate struct {
	DefaultCodeFileSaver[*ai.HtmlCodeResponse]
}

func (h *HtmlCodeFileSaverTemplate) getCodeType() enum.CodeGenTypeEnum {
	return enum.HtmlCodeGen
}

func (h *HtmlCodeFileSaverTemplate) saveFiles(response *ai.HtmlCodeResponse, baseDir string) error {
	fileName := "index.html"
	return h.writeToFile(baseDir, fileName, response.HtmlCode)
}

func (h *HtmlCodeFileSaverTemplate) validateInput(response *ai.HtmlCodeResponse) error {
	if response == nil {
		return fmt.Errorf("代码结果为空")
	}
	if response.HtmlCode == "" {
		return fmt.Errorf("HTML 代码为空")
	}
	return nil
}

type MultiFileCodeFileSaverTemplate struct {
	DefaultCodeFileSaver[*ai.MultiFileCodeResponse]
}

func (m *MultiFileCodeFileSaverTemplate) getCodeType() enum.CodeGenTypeEnum {
	return enum.MultiFileGen
}

func (m *MultiFileCodeFileSaverTemplate) saveFiles(response *ai.MultiFileCodeResponse, baseDir string) error {
	// 保存 HTML 文件
	err := m.writeToFile(baseDir, "index.html", response.HtmlCode)
	if err != nil {
		return err
	}
	// 保存 JS 文件
	err = m.writeToFile(baseDir, "script.js", response.JsCode)
	if err != nil {
		return err
	}
	// 保存 CSS 文件
	err = m.writeToFile(baseDir, "style.css", response.CssCode)
	if err != nil {
		return err
	}
	return nil
}

func (m *MultiFileCodeFileSaverTemplate) validateInput(response *ai.MultiFileCodeResponse) error {
	if response == nil {
		return fmt.Errorf("代码结果为空")
	}
	if response.HtmlCode == "" {
		return fmt.Errorf("HTML 代码为空")
	}
	if response.JsCode == "" {
		return fmt.Errorf("JS 代码为空")
	}
	if response.CssCode == "" {
		return fmt.Errorf("CSS 代码为空")
	}
	return nil
}

type CodeFileSaverExecutor struct {
	htmlCodeFileSaver      *HtmlCodeFileSaverTemplate
	multiFileCodeFileSaver *MultiFileCodeFileSaverTemplate
}

func NewCodeFileSaverExecutor() *CodeFileSaverExecutor {
	return &CodeFileSaverExecutor{
		htmlCodeFileSaver:      &HtmlCodeFileSaverTemplate{},
		multiFileCodeFileSaver: &MultiFileCodeFileSaverTemplate{},
	}
}

func (e *CodeFileSaverExecutor) ExecuteSaver(content interface{}, saveType enum.CodeGenTypeEnum, appId int64) (string, error) {
	switch saveType {
	case enum.HtmlCodeGen:
		return e.htmlCodeFileSaver.saveCode(
			content.(*ai.HtmlCodeResponse),
			appId,
			e.htmlCodeFileSaver.getCodeType,
			e.htmlCodeFileSaver.saveFiles,
			e.htmlCodeFileSaver.validateInput,
		)
	case enum.MultiFileGen:
		return e.multiFileCodeFileSaver.saveCode(
			content.(*ai.MultiFileCodeResponse),
			appId,
			e.multiFileCodeFileSaver.getCodeType,
			e.multiFileCodeFileSaver.saveFiles,
			e.multiFileCodeFileSaver.validateInput,
		)
	default:
		return "", fmt.Errorf("不支持的代码文件类型: %s", saveType)
	}
}
