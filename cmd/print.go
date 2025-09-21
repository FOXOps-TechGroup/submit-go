package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FOXOps-TechGroup/submit-go/pkg/submit"
)

// printCmd 表示打印命令
var printCmd = &cobra.Command{
	Use:   "print [filepath]",
	Short: "Submit a file for printing",
	Long:  `Submit a file for printing instead of judging.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return printFile(cmd.Context(), args[0])
	},
}

// printFile 提交文件进行打印
func printFile(ctx context.Context, filePath string) error {
	// 初始化提交器
	if err := initSubmitter(ctx); err != nil {
		return err
	}

	// 验证文件
	validFiles, err := validateFiles([]string{filePath})
	if err != nil {
		return err
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("no valid file to print")
	}

	// 准备提交选项
	options := &submit.SubmissionOptions{
		Files:     validFiles,
		PrintMode: true,
		AssumeYes: assumeYesFlag,
	}

	// 设置凭据
	options.Credentials = submit.Credentials{
		Username: viper.GetString("credentials.username"),
		Password: viper.GetString("credentials.password"),
		Token:    viper.GetString("credentials.token"),
	}

	// 如果指定了token参数，覆盖配置
	if tokenFlag != "" {
		options.Credentials.Token = tokenFlag
	}

	// 获取竞赛ID
	contestID := viper.GetString("default_contest")
	if contestFlag != "" {
		contestID = contestFlag
	}

	if contestID == "" {
		// 尝试自动选择唯一的竞赛
		contests, err := activeSubmitter.GetContests(ctx)
		if err != nil {
			return fmt.Errorf("failed to get contests: %w", err)
		}

		if len(contests) == 1 {
			contestID = contests[0].ID
		} else if len(contests) > 1 {
			return fmt.Errorf("multiple contests available, please specify one using --contest flag")
		} else {
			return fmt.Errorf("no contests available")
		}
	}

	options.ContestID = contestID

	// 如果指定了语言，尝试设置
	if languageFlag != "" {
		options.LanguageID = languageFlag
	} else {
		// 尝试从文件名推断语言
		_, languageID, err := activeSubmitter.InferProblemAndLanguage(ctx, contestID, filePath)
		if err == nil && languageID != "" {
			options.LanguageID = languageID
		}
	}

	// 显示打印信息
	if !quietFlag {
		fmt.Println("Print job information:")
		fmt.Printf("  filename:    %s\n", filePath)
		fmt.Printf("  contest:     %s\n", contestID)

		if options.LanguageID != "" {
			fmt.Printf("  language:    %s\n", options.LanguageID)
		}

		fmt.Printf("  url:         %s\n", viper.GetString("url"))
	}

	// 确认打印
	if !assumeYesFlag {
		fmt.Print("Do you want to continue? (y/n) ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			return fmt.Errorf("print job aborted by user")
		}
	}

	// 准备文件内容
	files, err := prepareFileContents(validFiles)
	if err != nil {
		return err
	}

	// 执行打印
	result, err := activeSubmitter.Submit(ctx, options, files)
	if err != nil {
		return err
	}

	// 显示结果
	fmt.Println(result.Message)

	return nil
}
