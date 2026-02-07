package pkg

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetCodeDeployRoot() (string, error) {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return "", fmt.Errorf("获取项目根目录失败: %w", err)
	}
	return filepath.Join(projectRoot, "tmp/code_deploy"), nil
}

func GetCodeOutputRoot() (string, error) {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return "", fmt.Errorf("获取项目根目录失败: %w", err)
	}
	return filepath.Join(projectRoot, "tmp/code_output"), nil
}

func GetProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err == nil {
		if root := findGoModDir(cwd); root != "" {
			return root, nil
		}
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	execDir := filepath.Dir(execPath)
	if root := findGoModDir(execDir); root != "" {
		return root, nil
	}

	cwd, err = os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	return cwd, nil
}

func findGoModDir(startDir string) string {
	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return ""
		}
		dir = parentDir
	}
}
