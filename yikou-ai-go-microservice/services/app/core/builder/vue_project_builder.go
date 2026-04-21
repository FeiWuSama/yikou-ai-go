package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bytedance/gopkg/util/logger"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func BuildCommand(baseCommand string) string {
	if IsWindows() {
		return baseCommand + ".cmd"
	}
	return baseCommand
}

func ExecuteCommand(workingDir string, command string, timeoutSeconds int) bool {
	logger.Infof("在目录 %s 中执行命令: %s", workingDir, command)

	var cmd *exec.Cmd
	if IsWindows() {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = workingDir

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.Errorf("执行命令失败: %s, 错误信息: %v", command, err)
			return false
		}
		if cmd.ProcessState.ExitCode() == 0 {
			logger.Infof("命令执行成功: %s", command)
			return true
		} else {
			logger.Errorf("命令执行失败，退出码: %d", cmd.ProcessState.ExitCode())
			return false
		}
	case <-time.After(time.Duration(timeoutSeconds) * time.Second):
		logger.Errorf("命令执行超时（%d秒），强制终止进程", timeoutSeconds)
		_ = cmd.Process.Kill()
		return false
	}
}

func ExecuteNpmInstall(projectDir string) bool {
	logger.Info("执行 npm install...")
	command := fmt.Sprintf("%s install", BuildCommand("npm"))
	return ExecuteCommand(projectDir, command, 300)
}

func ExecuteNpmBuild(projectDir string) bool {
	logger.Info("执行 npm run build...")
	command := fmt.Sprintf("%s run build", BuildCommand("npm"))
	return ExecuteCommand(projectDir, command, 180)
}

func BuildProject(projectPath string) bool {
	projectDir, err := os.Open(projectPath)
	if err != nil {
		logger.Errorf("项目目录不存在: %s", projectPath)
		return false
	}
	defer projectDir.Close()

	projectInfo, err := projectDir.Stat()
	if err != nil || !projectInfo.IsDir() {
		logger.Errorf("项目目录不存在: %s", projectPath)
		return false
	}

	packageJson := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJson); os.IsNotExist(err) {
		logger.Errorf("package.json 文件不存在: %s", packageJson)
		return false
	}

	logger.Infof("开始构建 Vue 项目: %s", projectPath)

	if !ExecuteNpmInstall(projectPath) {
		logger.Error("npm install 执行失败")
		return false
	}

	if !ExecuteNpmBuild(projectPath) {
		logger.Error("npm run build 执行失败")
		return false
	}

	distDir := filepath.Join(projectPath, "dist")
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		logger.Errorf("构建完成但 dist 目录未生成: %s", distDir)
		return false
	}

	logger.Infof("Vue 项目构建成功，dist 目录: %s", distDir)
	return true
}
