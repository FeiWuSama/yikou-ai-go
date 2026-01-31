package core

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

func BuildUniqueDir(typeStr enum.CodeGenType) (string, error) {
	var sf = sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) { return 1, nil },
	})
	id, err := sf.NextID()
	if err != nil {
		return "", err
	}
	uniqueDirName := fmt.Sprintf("%s_%s", typeStr, strconv.FormatUint(id, 20))
	dirPath := filepath.Join(FileSaveDir, uniqueDirName)
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return dirPath, nil
}

func WriteToFile(dirPath string, fileName string, content string) error {
	filePath := filepath.Join(dirPath, fileName)
	err := os.WriteFile(filePath, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func SaveHtmlCode(response ai.HtmlCodeResponse) error {
	dirPath, err := BuildUniqueDir(enum.HtmlCodeGen)
	if err != nil {
		return err
	}
	fileName := "index.html"
	return WriteToFile(dirPath, fileName, response.HtmlCode)
}

func SaveMutiFileCode(response ai.MultiFileCodeResponse) error {
	dirPath, err := BuildUniqueDir(enum.MultiFileGen)
	if err != nil {
		return err
	}
	// 保存 HTML 文件
	err = WriteToFile(dirPath, "index.html", response.HtmlCode)
	if err != nil {
		return err
	}
	// 保存 JS 文件
	err = WriteToFile(dirPath, "script.js", response.JsCode)
	if err != nil {
		return err
	}
	// 保存 CSS 文件
	err = WriteToFile(dirPath, "style.css", response.CssCode)
	if err != nil {
		return err
	}
	return nil
}
