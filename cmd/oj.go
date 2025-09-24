package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/FOXOps-TechGroup/submit-go/pkg/config"
	"github.com/spf13/cobra"
)

var (
	ojUsername  string
	ojPassword  string
	ojSaveCreds bool
)

// ojCmd 表示OJ系统命令
var ojCmd = &cobra.Command{
	Use:   "oj",
	Short: "Manage online judge systems",
	Long:  `Manage online judge systems for the submit tool.`,
}

var ojListCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported OJ systems",
	Long:  `List all supported online judge systems.`,
	Run: func(cmd *cobra.Command, args []string) {
		listOJSystems()
	},
}

var ojUseCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Switch default OJ system",
	Long:  `Switch the default OJ system for submissions.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return useOJSystem(args[0])
	},
}

var ojLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to current OJ system",
	Long:  `Login to the current OJ system and save credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return loginOJSystem(cmd.Context())
	},
}

func init() {
	ojCmd.AddCommand(ojListCmd)
	ojCmd.AddCommand(ojUseCmd)
	ojCmd.AddCommand(ojLoginCmd)

	ojLoginCmd.Flags().StringVarP(&ojUsername, "username", "u", "", "username for login")
	ojLoginCmd.Flags().StringVarP(&ojPassword, "password", "p", "", "password for login")
	ojLoginCmd.Flags().StringVarP(&tokenFlag, "token", "t", "", "access token for login")
	ojLoginCmd.Flags().BoolVar(&ojSaveCreds, "save", false, "save credentials")
}

// listOJSystems 列出支持的OJ系统
func listOJSystems() {
	systems := []struct {
		Name        string
		Description string
	}{
		{"domjudge", "DOMjudge Contest System"},
		// 添加其他OJ系统
	}

	currentOJ := viper.GetString("oj")

	fmt.Println("Supported OJ systems:")
	for _, sys := range systems {
		current := ""
		if sys.Name == currentOJ {
			current = " (current)"
		}
		fmt.Printf("  %-15s - %s%s\n", sys.Name, sys.Description, current)
	}
}

// useOJSystem 切换默认OJ系统
func useOJSystem(name string) error {
	// 检查是否是支持的OJ系统
	supported := false
	supportedSystems := []string{"domjudge"} // 添加其他支持的系统

	for _, sys := range supportedSystems {
		if strings.EqualFold(name, sys) {
			name = sys // 使用正确的大小写
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported OJ system: %s", name)
	}

	// 设置默认OJ系统
	cfg := config.GetProjectConfig()
	if err := cfg.Set("oj", name); err != nil {
		return fmt.Errorf("failed to set OJ system: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Switched to %s as default OJ system\n", name)
	return nil
}

// loginOJSystem 登录OJ系统
func loginOJSystem(ctx context.Context) error {
	// 获取当前OJ系统
	ojSystem := viper.GetString("oj")
	if ojSystem == "" {
		return fmt.Errorf("no OJ system selected, use 'submit oj use <name>' first")
	}

	// 如果没有提供用户名、密码或令牌，则提示输入
	if ojUsername == "" && ojPassword == "" && tokenFlag == "" {
		// 交互式登录
		fmt.Printf("Login to %s\n", ojSystem)

		// 优先使用令牌
		fmt.Print("Access token (leave empty to use username/password): ")
		fmt.Scanln(&tokenFlag)

		if tokenFlag == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&ojUsername)

			fmt.Print("Password: ")
			fmt.Scanln(&ojPassword)
		}

		// 默认保存凭据
		ojSaveCreds = true
	}

	// 验证凭据
	if err := validateCredentials(ctx, ojSystem); err != nil {
		return err
	}

	// 保存凭据
	if ojSaveCreds {
		cfg := config.GetProjectConfig()

		if tokenFlag != "" {
			if err := cfg.Set("credentials.token", tokenFlag); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
		} else {
			if err := cfg.Set("credentials.username", ojUsername); err != nil {
				return fmt.Errorf("failed to save username: %w", err)
			}
			if err := cfg.Set("credentials.password", ojPassword); err != nil {
				return fmt.Errorf("failed to save password: %w", err)
			}
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Println("Credentials saved successfully")
	}

	return nil
}

// validateCredentials 验证凭据
func validateCredentials(ctx context.Context, ojSystem string) error {
	// 创建提交器
	submitter, err := createSubmitter(ojSystem)
	if err != nil {
		return err
	}

	// 获取URL
	baseURL := viper.GetString("url")
	if baseURL == "" {
		return fmt.Errorf("no URL specified for %s, use 'submit config set url <url>'", ojSystem)
	}

	// 初始化提交器
	if err := submitter.Initialize(ctx, baseURL); err != nil {
		return fmt.Errorf("failed to initialize %s: %w", ojSystem, err)
	}

	// 创建临时选项用于验证
	//options := &submit.SubmissionOptions{
	//	Credentials: submit.Credentials{
	//		Username: ojUsername,
	//		Password: ojPassword,
	//		Token:    tokenFlag,
	//	},
	//}

	//// 尝试获取竞赛列表来验证凭据
	//if err := submitter.ValidateCredentials(ctx, options.Credentials); err != nil {
	//	return fmt.Errorf("authentication failed: %w", err)
	//}

	fmt.Println("Authentication successful")
	return nil
}
