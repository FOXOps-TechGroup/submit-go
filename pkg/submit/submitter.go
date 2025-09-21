package submit

import (
	"context"
	"io"
)

// SubmissionResult 表示提交结果
type SubmissionResult struct {
	Success      bool   // 提交是否成功
	SubmissionID string // 提交ID
	Message      string // 提交结果消息
	URL          string // 查看结果的URL
}

// ContestInfo 表示竞赛信息
type ContestInfo struct {
	ID        string
	ShortName string
	Name      string
}

// ProblemInfo 表示问题信息
type ProblemInfo struct {
	ID    string
	Label string
	Name  string
}

// LanguageInfo 表示编程语言信息
type LanguageInfo struct {
	ID                  string
	Name                string
	Extensions          []string
	EntryPointRequired  bool
	EntryPointExtension string // 如果需要入口点，这里指定扩展名
}

// SubmissionOptions 表示提交选项
type SubmissionOptions struct {
	ContestID   string
	ProblemID   string
	LanguageID  string
	EntryPoint  string
	Files       []string
	PrintMode   bool
	AssumeYes   bool
	Credentials Credentials
}

// Credentials 表示认证信息
type Credentials struct {
	Username string
	Password string
	Token    string
}

// Submitter 接口定义了提交代码到OJ系统的基本操作
type Submitter interface {
	// Name 返回提交器的名称
	Name() string

	// Initialize 初始化提交器，可能需要连接服务器获取配置等
	Initialize(ctx context.Context, baseURL string) error

	// GetContests 获取当前可用的竞赛列表
	GetContests(ctx context.Context) ([]ContestInfo, error)

	// GetProblems 获取指定竞赛的问题列表
	GetProblems(ctx context.Context, contestID string) ([]ProblemInfo, error)

	// GetLanguages 获取指定竞赛支持的语言列表
	GetLanguages(ctx context.Context, contestID string) ([]LanguageInfo, error)

	// ValidateSubmission 验证提交选项
	ValidateSubmission(ctx context.Context, options *SubmissionOptions) error

	// Submit 提交代码或打印文件
	Submit(ctx context.Context, options *SubmissionOptions, files map[string]io.Reader) (*SubmissionResult, error)

	// InferProblemAndLanguage 从文件名推断问题和语言
	InferProblemAndLanguage(ctx context.Context, contestID string, filename string) (problemID string, languageID string, err error)

	// InferEntryPoint 推断入口点
	InferEntryPoint(ctx context.Context, languageInfo *LanguageInfo, filename string) (string, error)

	// ValidateCredentials 验证凭据，这个方法不应该做出除鉴权以外的任何举动
	ValidateCredentials(ctx context.Context, credentials Credentials) error
}
