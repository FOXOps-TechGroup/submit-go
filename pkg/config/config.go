package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 表示配置管理器
type Config struct {
	viper *viper.Viper
	file  string
}

var (
	globalConfig  *Config
	projectConfig *Config
)

// InitConfig 初始化配置
func InitConfig(cfgFile string) {
	// 如果提供了配置文件，直接使用
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找项目配置
		findProjectConfig()

		// 查找全局配置
		findGlobalConfig()
	}

	// 读取环境变量
	viper.SetEnvPrefix("SUBMIT")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	// 读取配置文件
	if err := viper.ReadInConfig(); err == nil {
		// 成功读取配置文件
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// 设置默认值
	viper.SetDefault("oj", "domjudge")
}

// findProjectConfig 查找项目配置文件
func findProjectConfig() {
	// 在当前目录及其父目录中查找.submit.yaml
	dir, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		configFile := filepath.Join(dir, ".submit.yaml")
		if _, err := os.Stat(configFile); err == nil {
			viper.SetConfigFile(configFile)
			return
		}

		// 移动到父目录
		parent := filepath.Dir(dir)
		if parent == dir {
			// 已到达根目录
			break
		}
		dir = parent
	}
}

// findGlobalConfig 查找全局配置文件
func findGlobalConfig() {
	// 设置配置文件搜索路径
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// 添加全局配置目录
	configDir := filepath.Join(home, ".config", "submit")
	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
}

// GetGlobalConfig 获取全局配置管理器
func GetGlobalConfig() *Config {
	if globalConfig == nil {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("Failed to get home directory: %v", err))
		}

		configDir := filepath.Join(home, ".config", "submit")
		configFile := filepath.Join(configDir, "config.yaml")

		// 确保目录存在
		if err := os.MkdirAll(configDir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create config directory: %v", err))
		}

		// 创建新的viper实例
		v := viper.New()
		v.SetConfigFile(configFile)
		v.SetConfigType("yaml")

		// 读取配置文件
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				// 配置文件存在但无法读取
				fmt.Fprintf(os.Stderr, "Warning: Failed to read global config file: %v\n", err)
			}
			// 如果配置文件不存在，将在保存时创建
		}

		globalConfig = &Config{
			viper: v,
			file:  configFile,
		}
	}

	return globalConfig
}

// GetProjectConfig 获取项目配置管理器
func GetProjectConfig() *Config {
	if projectConfig == nil {
		// 获取当前工作目录
		dir, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Failed to get current directory: %v", err))
		}

		configFile := filepath.Join(dir, ".submit.yaml")

		// 创建新的viper实例
		v := viper.New()
		v.SetConfigFile(configFile)
		v.SetConfigType("yaml")

		// 读取配置文件
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				// 配置文件存在但无法读取
				fmt.Fprintf(os.Stderr, "Warning: Failed to read project config file: %v\n", err)
			}
			// 如果配置文件不存在，将在保存时创建
		}

		projectConfig = &Config{
			viper: v,
			file:  configFile,
		}
	}

	return projectConfig
}

// GetConfigFile 获取配置文件路径
func (c *Config) GetConfigFile() string {
	return c.file
}

// Get 获取配置值
func (c *Config) Get(key string) interface{} {
	return c.viper.Get(key)
}

// GetString 获取字符串配置值
func (c *Config) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetBool 获取布尔配置值
func (c *Config) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

// GetInt 获取整数配置值
func (c *Config) GetInt(key string) int {
	return c.viper.GetInt(key)
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) error {
	c.viper.Set(key, value)
	return nil
}

// Save 保存配置到文件
func (c *Config) Save() error {
	// 确保目录存在
	dir := filepath.Dir(c.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 保存配置
	return c.viper.WriteConfig()
}

// AllSettings 获取所有配置
func (c *Config) AllSettings() map[string]interface{} {
	return c.viper.AllSettings()
}
