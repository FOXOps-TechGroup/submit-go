package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FOXOps-TechGroup/submit-go/pkg/config"
)

var (
	globalConfigFlag  bool
	showAllConfigFlag bool
)

// configCmd 表示配置命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage submit tool configuration",
	Long:  `Manage configuration for the submit tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认显示当前配置
		displayConfig()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Long:  `Set a configuration value for the submit tool.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setConfig(args[0], args[1], globalConfigFlag)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Long:  `Get a configuration value from the submit tool.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		getConfig(args[0])
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `List all configuration values for the submit tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		listConfig(showAllConfigFlag)
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)

	configSetCmd.Flags().BoolVar(&globalConfigFlag, "global", false, "set global configuration instead of project configuration")
	configListCmd.Flags().BoolVar(&showAllConfigFlag, "all", false, "show all configurations, including default values")
}

// displayConfig 显示当前配置
func displayConfig() {
	fmt.Println("Current configuration:")

	// 显示常用配置项
	fmt.Printf("OJ system:      %s\n", viper.GetString("oj"))
	fmt.Printf("URL:            %s\n", viper.GetString("url"))
	fmt.Printf("Default contest: %s\n", viper.GetString("default_contest"))

	// 显示认证信息（隐藏敏感信息）
	username := viper.GetString("credentials.username")
	if username != "" {
		fmt.Printf("Username:       %s\n", username)
	}

	token := viper.GetString("credentials.token")
	if token != "" {
		fmt.Printf("Token:          %s\n", maskString(token))
	}

	// 显示配置文件路径
	fmt.Printf("\nConfiguration file: %s\n", viper.ConfigFileUsed())
}

// setConfig 设置配置项
func setConfig(key, value string, global bool) error {

	// 根据是否全局配置选择配置管理器
	var cfg *config.Config
	if global {
		cfg = config.GetGlobalConfig()
	} else {
		cfg = config.GetProjectConfig()
	}

	// 设置配置
	if err := cfg.Set(key, value); err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	// 保存配置
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Set %s = %s in %s\n", key, value, cfg.GetConfigFile())
	return nil
}

// getConfig 获取配置项
func getConfig(key string) {
	value := viper.Get(key)
	if value == nil {
		fmt.Printf("%s is not set\n", key)
	} else {
		fmt.Printf("%s = %v\n", key, value)
	}
}

// listConfig 列出所有配置
func listConfig(showAll bool) {
	allSettings := viper.AllSettings()
	fmt.Println("Configuration values:")

	// 以递归方式打印所有设置
	printSettings("", allSettings, showAll)

	fmt.Printf("\nConfiguration file: %s\n", viper.ConfigFileUsed())
}

// printSettings 递归打印设置
func printSettings(prefix string, settings map[string]interface{}, showAll bool) {
	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// 如果是嵌套map，递归处理
		if subSettings, ok := value.(map[string]interface{}); ok {
			printSettings(fullKey, subSettings, showAll)
		} else {
			// 处理敏感信息
			if strings.Contains(fullKey, "password") || strings.Contains(fullKey, "token") {
				if value != nil && value != "" {
					fmt.Printf("%-30s = %s\n", fullKey, maskString(fmt.Sprintf("%v", value)))
				} else if showAll {
					fmt.Printf("%-30s = %v\n", fullKey, value)
				}
			} else if value != nil && value != "" {
				fmt.Printf("%-30s = %v\n", fullKey, value)
			} else if showAll {
				fmt.Printf("%-30s = %v\n", fullKey, value)
			}
		}
	}
}

// maskString 隐藏敏感信息
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}
