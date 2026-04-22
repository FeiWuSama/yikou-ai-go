package config

import (
	"flag"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
	pkg "yikou-ai-go-microservice/pkg/myfile"
)

var env string

func init() {
	flag.StringVar(&env, "env", "", "配置文件后缀 (例如：local, prod)")
}

type Config struct {
	Server    ServerConfig    `yaml:"server" mapstructure:"server"`
	Database  DatabaseConfig  `yaml:"database" mapstructure:"database"`
	Redis     RedisConfig     `yaml:"redis" mapstructure:"redis"`
	AI        AIConfig        `yaml:"ai" mapstructure:"ai"`
	COS       COSConfig       `yaml:"cos" mapstructure:"cos"`
	Pexels    PexelsConfig    `yaml:"pexels" mapstructure:"pexels"`
	DashScope DashScopeConfig `yaml:"dashscope" mapstructure:"dashscope"`
	Nacos     NacosConfig     `yaml:"nacos" mapstructure:"nacos"`
}

type NacosConfig struct {
	Host        string `yaml:"host" mapstructure:"host"`
	Port        int    `yaml:"port" mapstructure:"port"`
	NamespaceId string `yaml:"namespace-id" mapstructure:"namespace-id"`
	Username    string `yaml:"username" mapstructure:"username"`
	Password    string `yaml:"password" mapstructure:"password"`
	LogDir      string `yaml:"log-dir" mapstructure:"log-dir"`
	CacheDir    string `yaml:"cache-dir" mapstructure:"cache-dir"`
	LogLevel    string `yaml:"log-level" mapstructure:"log-level"`
}

type ServerConfig struct {
	ConfigActive string `yaml:"config-active" mapstructure:"config-active"`
	Port         int    `yaml:"port" mapstructure:"port"`
	ContextPath  string `yaml:"context-path" mapstructure:"context-path"`
	EnableMetric bool   `yaml:"enable-metric" mapstructure:"enable-metric"`
	MetricPort   int    `yaml:"metric-port" mapstructure:"metric-port"`
	MetricPath   string `yaml:"metric-path" mapstructure:"metric-path"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	DBName   string `yaml:"dbname" mapstructure:"dbname"`
}

type RedisConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
}

type COSConfig struct {
	Host      string `yaml:"host" mapstructure:"host"`
	SecretID  string `yaml:"secret-id" mapstructure:"secret-id"`
	SecretKey string `yaml:"secret-key" mapstructure:"secret-key"`
	Region    string `yaml:"region" mapstructure:"region"`
	Bucket    string `yaml:"bucket" mapstructure:"bucket"`
}

type PexelsConfig struct {
	APIKey string `yaml:"api-key" mapstructure:"api-key"`
}

type DashScopeConfig struct {
	APIKey     string `yaml:"api-key" mapstructure:"api-key"`
	ImageModel string `yaml:"image-model" mapstructure:"image-model"`
}

type AIConfig struct {
	ChatModel          ChatModelConfig `yaml:"chat-model" mapstructure:"chat-model"`
	ReasoningChatModel ChatModelConfig `yaml:"reasoning-chat-model" mapstructure:"reasoning-chat-model"`
}

type ChatModelConfig struct {
	BaseURL     string `yaml:"base-url" mapstructure:"base-url"`
	APIKey      string `yaml:"api-key" mapstructure:"api-key"`
	ModelName   string `yaml:"model-name" mapstructure:"model-name"`
	MemoryStore string `yaml:"memory-store" mapstructure:"memory-store"`
	MemoryTTL   int    `yaml:"memory-ttl" mapstructure:"memory-ttl"`
}

func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName)
}

func mergeConfig(base, override interface{}) {
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

func mergeAIConfig(base, override *AIConfig) {
	mergeConfig(&base.ChatModel, &override.ChatModel)
	mergeConfig(&base.ReasoningChatModel, &override.ReasoningChatModel)
}

func InitConfig() *Config {
	flag.Parse()

	projectRoot, err := pkg.GetProjectRoot()
	if err != nil {
		panic(fmt.Errorf("获取项目根目录失败: %w", err))
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(projectRoot, "config"))

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("解析配置文件失败: %w", err))
	}

	configActive := config.Server.ConfigActive
	if env != "" {
		configActive = env
	}

	if configActive != "" {
		viper.SetConfigName(fmt.Sprintf("config-%s", configActive))
		if err := viper.MergeInConfig(); err == nil {
			var overrideConfig Config
			if err := viper.Unmarshal(&overrideConfig); err == nil {
				mergeConfig(&config.Server, &overrideConfig.Server)
				mergeConfig(&config.Redis, &overrideConfig.Redis)
				mergeAIConfig(&config.AI, &overrideConfig.AI)
				mergeConfig(&config.COS, &overrideConfig.COS)
				mergeConfig(&config.Pexels, &overrideConfig.Pexels)
				mergeConfig(&config.DashScope, &overrideConfig.DashScope)
				mergeConfig(&config.Nacos, &overrideConfig.Nacos)
			}
		}
	}

	return &config
}
