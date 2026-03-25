package chain

import (
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"testing"
)

func TestWorkflow(t *testing.T) {
	fmt.Println("=== 简化版网站生成工作流 ===")
	if err := RunSimpleWorkflow(); err != nil {
		logger.Errorf("工作流执行失败: %v", err)
	}
}
