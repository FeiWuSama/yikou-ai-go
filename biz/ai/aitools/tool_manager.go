package aitools

import (
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/eino/components/tool"
)

type ToolInterface interface {
	GetToolName() string
	GetDisplayName() string
	GenerateToolRequestResponse() string
	GenerateToolExecutedResult(arguments string) string
}

type ToolManager struct {
	fileWriteTool   *FileWriteTool
	fileDeleteTool  *FileDeleteTool
	fileReadTool    *FileReadTool
	fileDirReadTool *FileDirReadTool
	fileModifyTool  *FileModifyTool
	toolMap         map[string]ToolInterface
}

func NewToolManager() (*ToolManager, error) {
	fileWriteTool, err := CreateFileWriteTool()
	if err != nil {
		return nil, err
	}

	fileDeleteTool, err := CreateFileDeleteTool()
	if err != nil {
		return nil, err
	}

	fileReadTool, err := CreateFileReadTool()
	if err != nil {
		return nil, err
	}

	fileDirReadTool, err := CreateFileDirReadTool()
	if err != nil {
		return nil, err
	}

	fileModifyTool, err := CreateFileModifyTool()
	if err != nil {
		return nil, err
	}

	toolMap := make(map[string]ToolInterface)

	tools := []ToolInterface{
		fileWriteTool,
		fileDeleteTool,
		fileReadTool,
		fileDirReadTool,
		fileModifyTool,
	}

	for _, t := range tools {
		toolMap[t.GetToolName()] = t
		logger.Infof("注册工具: %s -> %s\n", t.GetToolName(), t.GetDisplayName())
	}
	logger.Infof("工具管理器初始化完成，共注册 %d 个工具\n", len(toolMap))

	return &ToolManager{
		fileWriteTool:   fileWriteTool,
		fileDeleteTool:  fileDeleteTool,
		fileReadTool:    fileReadTool,
		fileDirReadTool: fileDirReadTool,
		fileModifyTool:  fileModifyTool,
		toolMap:         toolMap,
	}, nil
}

func (m *ToolManager) GetTool(toolName string) ToolInterface {
	return m.toolMap[toolName]
}

func (m *ToolManager) GetAllTools() []tool.BaseTool {
	return []tool.BaseTool{
		m.fileWriteTool.BaseTool,
		m.fileDeleteTool.BaseTool,
		m.fileReadTool.BaseTool,
		m.fileDirReadTool.BaseTool,
		m.fileModifyTool.BaseTool,
	}
}
