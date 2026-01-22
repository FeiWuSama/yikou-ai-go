package config

import (
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"testing"
)

func TestInitConfig(t *testing.T) {
	err := InitConfig()
	if err != nil {
		t.Fatalf("初始化配置文件失败: %v", err)
	}
	assert.NotNil(t, GlobalConfig)
}
