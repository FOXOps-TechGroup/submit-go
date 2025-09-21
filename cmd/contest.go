package cmd

import (
	"context"
	"fmt"
	"github.com/FOXOps-TechGroup/submit-go/pkg/submit"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FOXOps-TechGroup/submit-go/pkg/config"
)

// contestCmd 表示竞赛命令
var contestCmd = &cobra.Command{
	Use:   "contest",
	Short: "Manage contests",
	Long:  `Manage contests for the submit tool.`,
}

var contestListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available contests",
	Long:  `List all available contests for the current OJ system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listContests(cmd.Context())
	},
}

var contestUseCmd = &cobra.Command{
	Use:   "use [id]",
	Short: "Switch current contest",
	Long:  `Switch the current contest for submissions.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return useContest(cmd.Context(), args[0])
	},
}

func init() {
	contestCmd.AddCommand(contestListCmd)
	contestCmd.AddCommand(contestUseCmd)

	// 将此命令添加到根命令
	rootCmd.AddCommand(contestCmd)
}

// listContests 列出可用竞赛
func listContests(ctx context.Context) error {
	// 初始化提交器
	if err := initSubmitter(ctx); err != nil {
		return err
	}

	// 获取竞赛列表
	contests, err := activeSubmitter.GetContests(ctx)
	if err != nil {
		return fmt.Errorf("failed to get contests: %w", err)
	}

	if len(contests) == 0 {
		fmt.Println("No contests available")
		return nil
	}

	// 获取当前竞赛
	currentContestID := viper.GetString("default_contest")

	// 找出最长的shortname以便对齐
	maxLength := 0
	for _, contest := range contests {
		if len(contest.ShortName) > maxLength {
			maxLength = len(contest.ShortName)
		}
	}

	fmt.Println("Available contests:")
	for _, contest := range contests {
		current := ""
		if contest.ID == currentContestID || contest.ShortName == currentContestID {
			current = " (current)"
		}
		fmt.Printf("  %-*s - %s%s\n", maxLength+3, contest.ShortName, contest.Name, current)
	}

	return nil
}

// useContest 切换当前竞赛
func useContest(ctx context.Context, contestID string) error {
	// 初始化提交器
	if err := initSubmitter(ctx); err != nil {
		return err
	}

	// 验证竞赛ID - 通过获取所有竞赛并查找匹配项
	contests, err := activeSubmitter.GetContests(ctx)
	if err != nil {
		return fmt.Errorf("failed to get contests: %w", err)
	}

	var targetContest *submit.ContestInfo
	lowercaseID := strings.ToLower(contestID)

	for i, contest := range contests {
		if strings.ToLower(contest.ID) == lowercaseID ||
			strings.ToLower(contest.ShortName) == lowercaseID {
			targetContest = &contests[i]
			break
		}
	}

	if targetContest == nil {
		return fmt.Errorf("contest '%s' not found", contestID)
	}

	// 设置默认竞赛
	cfg := config.GetProjectConfig()
	if err := cfg.Set("default_contest", targetContest.ID); err != nil {
		return fmt.Errorf("failed to set default contest: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Switched to contest '%s' (%s)\n", targetContest.ShortName, targetContest.Name)
	return nil
}
