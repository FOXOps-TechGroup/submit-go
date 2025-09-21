package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/FOXOps-TechGroup/submit-go/pkg/config"
	"github.com/FOXOps-TechGroup/submit-go/pkg/submit"
)

var (
	cfgFile        string
	contestFlag    string
	languageFlag   string
	problemFlag    string
	entryPointFlag string
	ojFlag         string
	urlFlag        string
	tokenFlag      string
	assumeYesFlag  bool
	verboseFlag    bool
	quietFlag      bool

	// 当前使用的提交器
	activeSubmitter submit.Submitter
)

// rootCmd 表示没有调用子命令时的基本命令
var rootCmd = &cobra.Command{
	Use:   "submit [flags] [filepath...]",
	Short: "A universal submission tool for programming contests",
	Long: `Submit is a command-line tool for submitting solutions to various online judges.
It supports multiple OJ systems and provides a unified interface for submissions.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 如果没有参数，显示帮助
		if len(args) == 0 {
			return cmd.Help()
		}

		// 否则执行提交
		return submitFiles(cmd.Context(), args)
	},
}

// Execute 添加所有子命令到根命令并设置标志。
// 这由 main.main() 调用。它只需要对 rootCmd 执行一次。
func Execute() error {
	ctx := context.Background()
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/submit/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&contestFlag, "contest", "c", "", "specify contest ID or shortname")
	rootCmd.PersistentFlags().StringVarP(&languageFlag, "language", "l", "", "specify language ID or extension")
	rootCmd.PersistentFlags().StringVarP(&problemFlag, "problem", "p", "", "specify problem ID or label")
	rootCmd.PersistentFlags().StringVarP(&entryPointFlag, "entry", "e", "", "specify code entry point (e.g. Java main class)")
	rootCmd.PersistentFlags().StringVarP(&ojFlag, "oj", "o", "", "specify OJ system to use")
	rootCmd.PersistentFlags().StringVarP(&urlFlag, "url", "u", "", "specify OJ system URL")
	rootCmd.PersistentFlags().StringVarP(&tokenFlag, "token", "t", "", "use specified access token")
	rootCmd.PersistentFlags().BoolVarP(&assumeYesFlag, "yes", "y", false, "auto confirm, don't ask")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "quiet mode, only output necessary information")

	// 将标志绑定到viper
	viper.BindPFlag("contest", rootCmd.PersistentFlags().Lookup("contest"))
	viper.BindPFlag("language", rootCmd.PersistentFlags().Lookup("language"))
	viper.BindPFlag("problem", rootCmd.PersistentFlags().Lookup("problem"))
	viper.BindPFlag("entry_point", rootCmd.PersistentFlags().Lookup("entry"))
	viper.BindPFlag("oj", rootCmd.PersistentFlags().Lookup("oj"))
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("assume_yes", rootCmd.PersistentFlags().Lookup("yes"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

	// 添加子命令
	// 注意：这些命令将在各自的init函数中定义，这里只是引用
	// rootCmd.AddCommand(configCmd)
	// rootCmd.AddCommand(ojCmd)
	// rootCmd.AddCommand(contestCmd)
	// rootCmd.AddCommand(infoCmd)
	// rootCmd.AddCommand(printCmd)
}

// initConfig 读取配置文件和ENV变量
func initConfig() {
	config.InitConfig(cfgFile)
}

// submitFiles 处理文件提交
func submitFiles(ctx context.Context, filePaths []string) error {
	// 初始化提交器
	if err := initSubmitter(ctx); err != nil {
		return err
	}

	// 验证文件
	validFiles, err := validateFiles(filePaths)
	if err != nil {
		return err
	}

	if len(validFiles) == 0 {
		return fmt.Errorf("no valid files to submit")
	}

	// 准备提交选项
	options, err := prepareSubmissionOptions(ctx, validFiles)
	if err != nil {
		return err
	}

	// 显示提交信息
	if !quietFlag {
		displaySubmissionInfo(options, validFiles)
	}

	// 确认提交
	if !assumeYesFlag && !confirmSubmission() {
		return fmt.Errorf("submission aborted by user")
	}

	// 准备文件内容
	files, err := prepareFileContents(validFiles)
	if err != nil {
		return err
	}

	// 执行提交
	result, err := activeSubmitter.Submit(ctx, options, files)
	if err != nil {
		return err
	}

	// 显示结果
	fmt.Println(result.Message)
	if result.URL != "" {
		fmt.Printf("Check %s for the result.\n", result.URL)
	}

	return nil
}

// initSubmitter 初始化提交器
func initSubmitter(ctx context.Context) error {
	// 确定使用的OJ系统
	ojSystem := viper.GetString("oj")
	if ojFlag != "" {
		ojSystem = ojFlag
	}

	if ojSystem == "" {
		return fmt.Errorf("no OJ system specified, use --oj flag or set in config")
	}

	// 创建提交器实例
	var err error
	activeSubmitter, err = createSubmitter(ojSystem)
	if err != nil {
		return err
	}

	// 获取URL
	baseURL := viper.GetString("url")
	if urlFlag != "" {
		baseURL = urlFlag
	}

	if baseURL == "" {
		return fmt.Errorf("no URL specified for %s, use --url flag or set in config", ojSystem)
	}

	// 初始化提交器
	if err := activeSubmitter.Initialize(ctx, baseURL); err != nil {
		return fmt.Errorf("failed to initialize %s: %w", ojSystem, err)
	}

	return nil
}

// createSubmitter 创建指定OJ系统的提交器
func createSubmitter(ojSystem string) (submit.Submitter, error) {
	switch ojSystem {
	//case "domjudge":
	//	// 从配置中获取API版本
	//	apiVersion := viper.GetString("domjudge.api_version")
	//	return domjudge.NewDOMjudgeSubmitter(apiVersion), nil
	// 可以添加其他OJ系统
	default:
		return nil, fmt.Errorf("unsupported OJ system: %s", ojSystem)
	}
}

// validateFiles 验证提交文件
func validateFiles(filePaths []string) ([]string, error) {
	// 实现文件验证逻辑
	validFiles := make([]string, 0, len(filePaths))

	for _, path := range filePaths {
		// 检查文件是否存在
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("file '%s' not found or not accessible: %w", path, err)
		}

		// 检查是否是常规文件
		if !fileInfo.Mode().IsRegular() {
			if !quietFlag {
				fmt.Printf("WARNING: '%s' is not a regular file!\n", path)
			}
		}

		// 添加到有效文件列表
		validFiles = append(validFiles, path)
	}

	return validFiles, nil
}

// prepareSubmissionOptions 准备提交选项
func prepareSubmissionOptions(ctx context.Context, files []string) (*submit.SubmissionOptions, error) {
	options := &submit.SubmissionOptions{
		Files:     files,
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
			return nil, fmt.Errorf("failed to get contests: %w", err)
		}

		if len(contests) == 1 {
			contestID = contests[0].ID
		} else if len(contests) > 1 {
			return nil, fmt.Errorf("multiple contests available, please specify one using --contest flag")
		} else {
			return nil, fmt.Errorf("no contests available")
		}
	}

	options.ContestID = contestID

	// 获取问题ID
	if problemFlag != "" {
		options.ProblemID = problemFlag
	} else if len(files) > 0 {
		// 尝试从第一个文件名推断问题
		problemID, _, err := activeSubmitter.InferProblemAndLanguage(ctx, contestID, files[0])
		if err != nil {
			return nil, fmt.Errorf("failed to infer problem: %w", err)
		}
		options.ProblemID = problemID
	}

	// 获取语言ID
	if languageFlag != "" {
		options.LanguageID = languageFlag
	} else if len(files) > 0 {
		// 尝试从第一个文件名推断语言
		_, languageID, err := activeSubmitter.InferProblemAndLanguage(ctx, contestID, files[0])
		if err != nil {
			return nil, fmt.Errorf("failed to infer language: %w", err)
		}
		options.LanguageID = languageID
	}

	// 获取入口点
	if entryPointFlag != "" {
		options.EntryPoint = entryPointFlag
	} else if options.LanguageID != "" {
		// 获取语言信息
		languages, err := activeSubmitter.GetLanguages(ctx, contestID)
		if err == nil && len(languages) > 0 {
			// 查找匹配的语言
			for _, lang := range languages {
				if strings.EqualFold(lang.ID, options.LanguageID) {
					if lang.EntryPointRequired && len(files) > 0 {
						// 尝试推断入口点
						entryPoint, err := activeSubmitter.InferEntryPoint(ctx, &lang, files[0])
						if err == nil {
							options.EntryPoint = entryPoint
						}
					}
					break
				}
			}
		}
	}

	return options, nil
}

// displaySubmissionInfo 显示提交信息
func displaySubmissionInfo(options *submit.SubmissionOptions, files []string) {
	fmt.Println("Submission information:")

	if len(files) == 1 {
		fmt.Printf("  filename:    %s\n", files[0])
	} else {
		fmt.Printf("  filenames:   %s\n", strings.Join(files, " "))
	}

	fmt.Printf("  contest:     %s\n", options.ContestID)
	fmt.Printf("  problem:     %s\n", options.ProblemID)
	fmt.Printf("  language:    %s\n", options.LanguageID)

	if options.EntryPoint != "" {
		fmt.Printf("  entry point: %s\n", options.EntryPoint)
	}

	fmt.Printf("  url:         %s\n", viper.GetString("url"))
}

// confirmSubmission 请求用户确认提交
func confirmSubmission() bool {
	fmt.Print("Do you want to continue? (y/n) ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

// prepareFileContents 准备文件内容
func prepareFileContents(files []string) (map[string]io.Reader, error) {
	result := make(map[string]io.Reader)

	for _, path := range files {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", path, err)
		}
		result[path] = file
	}

	return result, nil
}
