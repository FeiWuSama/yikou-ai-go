package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

var GlobalConfig *Config

type ServerConfig struct {
	Name           string `yaml:"name"`
	ConfigValidate string `yaml:"config-validate"`
	Port           int    `yaml:"port"`
	ContextPath    string `yaml:"context-path"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
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

// InitConfig 初始化配置文件
func InitConfig() error {
	configPath := "config/config.yml"
	// 获取配置文件的绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}
	// 读取配置文件内容
	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}
	// 解析配置文件内容到 Config 结构体
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	// 如果配置文件中指定了配置验证模式，合并验证模式下的配置
	if config.Server.ConfigValidate != "" {
		overrideConfigPath := fmt.Sprintf("config/config-%s.yml", config.Server.ConfigValidate)
		overrideAbsPath, err := filepath.Abs(overrideConfigPath)
		if err != nil {
			return err
		}
		overrideData, err := os.ReadFile(overrideAbsPath)
		if err == nil {
			var overrideConfig Config
			if err := yaml.Unmarshal(overrideData, &overrideConfig); err == nil {
				// 合并验证模式下的配置到基础配置
				mergeServerConfig(&config.Server, &overrideConfig.Server)
			}
		}
	}

	GlobalConfig = &config
	return nil
}
