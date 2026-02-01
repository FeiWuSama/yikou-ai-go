package saver

import (
	"fmt"
	"github.com/sony/sonyflake"
	"os"
	"path/filepath"
	"strconv"
	ai "workspace-yikou-ai-go/biz/ai/model"
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
func buildUniqueDir(typeStr enum.CodeGenType) (string, error) {
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
func writeToFile(dirPath string, fileName string, content string) error {
	filePath := filepath.Join(dirPath, fileName)
	err := os.WriteFile(filePath, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// SaveHtmlCode 保存 HTML 代码文件
func SaveHtmlCode(response ai.HtmlCodeResponse) (string, error) {
	dirPath, err := buildUniqueDir(enum.HtmlCodeGen)
	if err != nil {
		return "", err
	}
	fileName := "index.html"
	return dirPath, writeToFile(dirPath, fileName, response.HtmlCode)
}

// SaveMutiFileCode 保存多文件代码文件
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
