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
	Server ServerConfig `yaml:"server" mapstructure:"server"`
	Nacos  NacosConfig  `yaml:"nacos" mapstructure:"nacos"`
	Proxy  ProxyConfig  `yaml:"proxy" mapstructure:"proxy"`
}

type ServerConfig struct {
	ConfigActive string `yaml:"config-active" mapstructure:"config-active"`
	Port         int    `yaml:"port" mapstructure:"port"`
	ContextPath  string `yaml:"context-path" mapstructure:"context-path"`
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

type ProxyConfig struct {
	Routes []RouteConfig `yaml:"routes" mapstructure:"routes"`
}

type RouteConfig struct {
	Path        string `yaml:"path" mapstructure:"path"`
	Service     string `yaml:"service" mapstructure:"service"`
	StripPath   bool   `yaml:"strip-path" mapstructure:"strip-path"`
	StripPrefix string `yaml:"strip-prefix" mapstructure:"strip-prefix"`
	RewritePath string `yaml:"rewrite-path" mapstructure:"rewrite-path"`
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
				mergeConfig(&config.Nacos, &overrideConfig.Nacos)
				mergeConfig(&config.Proxy, &overrideConfig.Proxy)
			}
		}
	}

	return &config
}
