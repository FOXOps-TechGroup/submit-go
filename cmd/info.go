package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var showAllInfoFlag bool

// infoCmd 表示信息查询命令
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Query information",
	Long:  `Query information about problems, languages, etc.`,
}

var infoProblemCmd = &cobra.Command{
	Use:   "problem",
	Short: "List problems for current contest",
	Long:  `List all problems for the current contest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listProblems(cmd.Context())
	},
}

var infoLanguageCmd = &cobra.Command{
	Use:   "language",
	Short: "List supported languages for current contest",
	Long:  `List all supported languages for the current contest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listLanguages(cmd.Context())
	},
}

func init() {
	infoCmd.AddCommand(infoProblemCmd)
	infoCmd.AddCommand(infoLanguageCmd)

	infoProblemCmd.Flags().BoolVar(&showAllInfoFlag, "all", false, "show detailed information")
	infoLanguageCmd.Flags().BoolVar(&showAllInfoFlag, "all", false, "show detailed information, including extensions")
}

// listProblems 列出当前竞赛的问题
func listProblems(ctx context.Context) error {
	// 初始化提交器
	if err := initSubmitter(ctx); err != nil {
		return err
	}

	// 获取当前竞赛
	contestID := viper.GetString("default_contest")
	if contestFlag != "" {
		contestID = contestFlag
	}

	if contestID == "" {
		return fmt.Errorf("no contest selected, use --contest flag or 'submit contest use <id>'")
	}

	// 获取问题列表
	problems, err := activeSubmitter.GetProblems(ctx, contestID)
	if err != nil {
		return fmt.Errorf("failed to get problems: %w", err)
	}

	if len(problems) == 0 {
		fmt.Printf("No problems available for contest %s\n", contestID)
		return nil
	}

	// 找出最长的标签以便对齐
	maxLabelLength := 0
	for _, problem := range problems {
		if len(problem.Label) > maxLabelLength {
			maxLabelLength = len(problem.Label)
		}
	}

	fmt.Printf("Problems for contest %s:\n", contestID)
	for _, problem := range problems {
		fmt.Printf("  %-*s - %s\n", maxLabelLength+3, problem.Label, problem.Name)

		if showAllInfoFlag {
			fmt.Printf("    ID: %s\n", problem.ID)
			// 如果API提供更多信息，可以在这里显示
		}
	}

	return nil
}

// listLanguages 列出当前竞赛支持的语言
func listLanguages(ctx context.Context) error {
	// 初始化提交器
	if err := initSubmitter(ctx); err != nil {
		return err
	}

	// 获取当前竞赛
	contestID := viper.GetString("default_contest")
	if contestFlag != "" {
		contestID = contestFlag
	}

	if contestID == "" {
		return fmt.Errorf("no contest selected, use --contest flag or 'submit contest use <id>'")
	}

	// 获取语言列表
	languages, err := activeSubmitter.GetLanguages(ctx, contestID)
	if err != nil {
		return fmt.Errorf("failed to get languages: %w", err)
	}

	if len(languages) == 0 {
		fmt.Printf("No languages available for contest %s\n", contestID)
		return nil
	}

	// 找出最长的名称以便对齐
	maxNameLength := 0
	for _, lang := range languages {
		if len(lang.Name) > maxNameLength {
			maxNameLength = len(lang.Name)
		}
	}

	fmt.Printf("Languages for contest %s:\n", contestID)
	for _, lang := range languages {
		if showAllInfoFlag {
			fmt.Printf("  %-*s - ID: %s\n", maxNameLength+3, lang.Name, lang.ID)

			if len(lang.Extensions) > 0 {
				sortedExts := make([]string, 0, len(lang.Extensions))
				for _, ext := range lang.Extensions {
					sortedExts = append(sortedExts, ext)
				}
				fmt.Printf("    Extensions: %s\n", strings.Join(sortedExts, ", "))
			}

			if lang.EntryPointRequired {
				fmt.Printf("    Entry point required: yes\n")
			}
		} else {
			fmt.Printf("  %-*s - %s\n", maxNameLength+3, lang.Name, lang.ID)
		}
	}

	return nil
}
