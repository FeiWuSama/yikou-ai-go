package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
	"workspace-yikou-ai-go/pkg/file"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	AI       AIConfig       `yaml:"ai"`
}

var GlobalConfig *Config

type ServerConfig struct {
	ConfigActive string `yaml:"config-active"`
	Port         int    `yaml:"port"`
	ContextPath  string `yaml:"context-path"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type AIConfig struct {
	ChatModel ChatModelConfig `yaml:"chat-model"`
}

type ChatModelConfig struct {
	BaseURL   string `yaml:"base-url"`
	APIKey    string `yaml:"api-key"`
	ModelName string `yaml:"model-name"`
}

func init() {
	if err := InitConfig(); err != nil {
		log.Fatalf("初始化配置文件失败: %v", err)
	}
}

func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName)
}

// mergeServerConfig 合并 ServerConfig 结构体的配置
func mergeServerConfig(base, override *ServerConfig) {
	baseValue := reflect.ValueOf(base).Elem()
	overrideValue := reflect.ValueOf(override).Elem()
	overrideType := overrideValue.Type()
	// 遍历 override 结构体的字段，合并非空值到 base 结构体
	for i := 0; i < overrideValue.NumField(); i++ {
		fieldName := overrideType.Field(i).Name
		overrideField := overrideValue.Field(i)
		baseField := baseValue.FieldByName(fieldName)
		// 根据字段类型合并非空值
		switch overrideField.Kind() {
		case reflect.String:
			if overrideField.String() != "" {
				baseField.SetString(overrideField.String())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if overrideField.Int() != 0 {
				baseField.SetInt(overrideField.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if overrideField.Uint() != 0 {
				baseField.SetUint(overrideField.Uint())
			}
		case reflect.Float32, reflect.Float64:
			if overrideField.Float() != 0 {
				baseField.SetFloat(overrideField.Float())
			}
		case reflect.Bool:
			if overrideField.Bool() {
				baseField.SetBool(overrideField.Bool())
			}
		}
	}
}

// mergeAIConfig 合并 AIConfig 结构体的配置
func mergeAIConfig(base, override *AIConfig) {
	mergeChatModelConfig(&base.ChatModel, &override.ChatModel)
}

// mergeChatModelConfig 合并 ChatModelConfig 结构体的配置
func mergeChatModelConfig(base, override *ChatModelConfig) {
	baseValue := reflect.ValueOf(base).Elem()
	overrideValue := reflect.ValueOf(override).Elem()
	overrideType := overrideValue.Type()
	for i := 0; i < overrideValue.NumField(); i++ {
		fieldName := overrideType.Field(i).Name
		overrideField := overrideValue.Field(i)
		baseField := baseValue.FieldByName(fieldName)
		switch overrideField.Kind() {
		case reflect.String:
			if overrideField.String() != "" {
				baseField.SetString(overrideField.String())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if overrideField.Int() != 0 {
				baseField.SetInt(overrideField.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if overrideField.Uint() != 0 {
				baseField.SetUint(overrideField.Uint())
			}
		case reflect.Float32, reflect.Float64:
			if overrideField.Float() != 0 {
				baseField.SetFloat(overrideField.Float())
			}
		case reflect.Bool:
			if overrideField.Bool() {
				baseField.SetBool(overrideField.Bool())
			}
		}
	}
}

// InitConfig 初始化配置文件
func InitConfig() error {
	projectRoot, err := file.GetProjectRoot()
	if err != nil {
		return fmt.Errorf("获取项目根目录失败: %w", err)
	}
	configPath := filepath.Join(projectRoot, "config/config.yml")
	// 读取配置文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	// 解析配置文件内容到 Config 结构体
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}
	// 如果配置文件中指定了配置验证模式，合并验证模式下的配置
	if config.Server.ConfigActive != "" {
		overrideConfigPath := filepath.Join(projectRoot, fmt.Sprintf("config/config-%s.yml", config.Server.ConfigActive))
		overrideData, err := os.ReadFile(overrideConfigPath)
		if err == nil {
			var overrideConfig Config
			if err := yaml.Unmarshal(overrideData, &overrideConfig); err == nil {
				mergeServerConfig(&config.Server, &overrideConfig.Server)
				mergeAIConfig(&config.AI, &overrideConfig.AI)
			}
		}
	}

	GlobalConfig = &config
	return nil
}
