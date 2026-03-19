package myfile

import (
	"fmt"
	"io"
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

func CopyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("源路径不是目录: %s", src)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
