package domjudge

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	submitter "github.com/FOXOps-TechGroup/submit-go/pkg/submit"
)

// DOMjudgeSubmitter 实现了Submitter接口，用于提交代码到DOMjudge
type DOMjudgeSubmitter struct {
	baseURL    string
	apiVersion string
	client     *http.Client
	contests   []submitter.ContestInfo
	problems   map[string][]submitter.ProblemInfo  // 按contestID索引
	languages  map[string][]submitter.LanguageInfo // 按contestID索引
}

// DOMjudgeSubmissionResponse 表示DOMjudge API返回的提交响应
type DOMjudgeSubmissionResponse struct {
	ID   string `json:"id"`
	Time string `json:"time"`
}

// DOMjudgePrintResponse 表示DOMjudge API返回的打印响应
type DOMjudgePrintResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
}

// NewDOMjudgeSubmitter 创建一个新的DOMjudge提交器
func NewDOMjudgeSubmitter(apiVersion string) *DOMjudgeSubmitter {
	return &DOMjudgeSubmitter{
		apiVersion: apiVersion,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		problems:  make(map[string][]submitter.ProblemInfo),
		languages: make(map[string][]submitter.LanguageInfo),
	}
}

// Name 返回提交器的名称
func (d *DOMjudgeSubmitter) Name() string {
	return "DOMjudge"
}

// Initialize 初始化提交器
func (d *DOMjudgeSubmitter) Initialize(ctx context.Context, baseURL string) error {
	// 确保URL以/结尾
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	d.baseURL = baseURL

	// 尝试获取竞赛列表以验证连接
	_, err := d.GetContests(ctx)
	return err
}

// GetContests 获取当前可用的竞赛列表
func (d *DOMjudgeSubmitter) GetContests(ctx context.Context) ([]submitter.ContestInfo, error) {
	if d.contests != nil && len(d.contests) > 0 {
		return d.contests, nil
	}

	var contests []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Shortname string `json:"shortname"`
	}

	err := d.apiGet(ctx, "contests", &contests)
	if err != nil {
		return nil, err
	}

	d.contests = make([]submitter.ContestInfo, len(contests))
	for i, contest := range contests {
		d.contests[i] = submitter.ContestInfo{
			ID:        contest.ID,
			ShortName: contest.Shortname,
			Name:      contest.Name,
		}
	}

	return d.contests, nil
}

// GetProblems 获取指定竞赛的问题列表
func (d *DOMjudgeSubmitter) GetProblems(ctx context.Context, contestID string) ([]submitter.ProblemInfo, error) {
	// 检查缓存
	if problems, ok := d.problems[contestID]; ok {
		return problems, nil
	}

	var problems []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Label string `json:"label"`
	}

	endpoint := fmt.Sprintf("contests/%s/problems", contestID)
	err := d.apiGet(ctx, endpoint, &problems)
	if err != nil {
		return nil, err
	}

	result := make([]submitter.ProblemInfo, len(problems))
	for i, problem := range problems {
		result[i] = submitter.ProblemInfo{
			ID:    problem.ID,
			Label: problem.Label,
			Name:  problem.Name,
		}
	}

	d.problems[contestID] = result
	return result, nil
}

// GetLanguages 获取指定竞赛支持的语言列表
func (d *DOMjudgeSubmitter) GetLanguages(ctx context.Context, contestID string) ([]submitter.LanguageInfo, error) {
	// 检查缓存
	if languages, ok := d.languages[contestID]; ok {
		return languages, nil
	}

	var languages []struct {
		ID                 string   `json:"id"`
		Name               string   `json:"name"`
		Extensions         []string `json:"extensions"`
		EntryPointRequired bool     `json:"entry_point_required"`
	}

	endpoint := fmt.Sprintf("contests/%s/languages", contestID)
	err := d.apiGet(ctx, endpoint, &languages)
	if err != nil {
		return nil, err
	}

	result := make([]submitter.LanguageInfo, len(languages))
	for i, lang := range languages {
		result[i] = submitter.LanguageInfo{
			ID:                 lang.ID,
			Name:               lang.Name,
			Extensions:         lang.Extensions,
			EntryPointRequired: lang.EntryPointRequired,
		}

		// 设置入口点扩展名
		switch lang.Name {
		case "Java":
			result[i].EntryPointExtension = ".java"
		case "Kotlin":
			result[i].EntryPointExtension = ".kt"
		case "Python 3":
			result[i].EntryPointExtension = ".py"
		}
	}

	d.languages[contestID] = result
	return result, nil
}

// ValidateSubmission 验证提交选项
func (d *DOMjudgeSubmitter) ValidateSubmission(ctx context.Context, options *submitter.SubmissionOptions) error {
	// 验证竞赛ID
	if options.ContestID == "" {
		return fmt.Errorf("contest ID is required")
	}

	// 验证文件
	if len(options.Files) == 0 {
		return fmt.Errorf("at least one file must be specified")
	}

	// 如果不是打印模式，需要验证问题和语言
	if !options.PrintMode {
		// 验证问题ID
		if options.ProblemID == "" {
			return fmt.Errorf("problem ID is required for submissions")
		}

		// 验证语言ID
		if options.LanguageID == "" {
			return fmt.Errorf("language ID is required for submissions")
		}

		// 检查语言是否需要入口点
		languages, err := d.GetLanguages(ctx, options.ContestID)
		if err != nil {
			return fmt.Errorf("failed to get languages: %w", err)
		}

		var selectedLanguage *submitter.LanguageInfo
		for _, lang := range languages {
			if strings.EqualFold(lang.ID, options.LanguageID) {
				selectedLanguage = &lang
				break
			}
		}

		if selectedLanguage == nil {
			return fmt.Errorf("language '%s' not found", options.LanguageID)
		}

		if selectedLanguage.EntryPointRequired && options.EntryPoint == "" {
			return fmt.Errorf("entry point is required for %s submissions", selectedLanguage.Name)
		}
	}

	return nil
}

// Submit 提交代码或打印文件
func (d *DOMjudgeSubmitter) Submit(ctx context.Context, options *submitter.SubmissionOptions, files map[string]io.Reader) (*submitter.SubmissionResult, error) {
	// 验证选项
	err := d.ValidateSubmission(ctx, options)
	if err != nil {
		return nil, err
	}

	// 根据模式选择提交或打印
	if options.PrintMode {
		return d.submitForPrinting(ctx, options, files)
	} else {
		return d.submitForJudging(ctx, options, files)
	}
}

// submitForJudging 提交代码进行评测
func (d *DOMjudgeSubmitter) submitForJudging(ctx context.Context, options *submitter.SubmissionOptions, files map[string]io.Reader) (*submitter.SubmissionResult, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加表单字段
	_ = writer.WriteField("problem", options.ProblemID)
	_ = writer.WriteField("language", options.LanguageID)
	if options.EntryPoint != "" {
		_ = writer.WriteField("entry_point", options.EntryPoint)
	}

	// 添加文件
	for name, reader := range files {
		fileWriter, err := writer.CreateFormFile("code[]", filepath.Base(name))
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}
		if _, err := io.Copy(fileWriter, reader); err != nil {
			return nil, fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	// 关闭multipart writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// 准备请求
	endpoint := fmt.Sprintf("contests/%s/submissions", options.ContestID)
	req, err := d.createRequest(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 设置认证
	if err := d.setAuthentication(req, options.Credentials); err != nil {
		return nil, err
	}

	// 发送请求
	var response DOMjudgeSubmissionResponse
	err = d.doRequest(req, &response)
	if err != nil {
		return nil, err
	}

	// 解析提交时间
	submitTime, err := time.Parse(time.RFC3339, response.Time)
	if err != nil {
		submitTime = time.Now()
	}

	// 构建结果URL
	resultURL := fmt.Sprintf("%steam/submission/%s", d.baseURL, response.ID)

	return &submitter.SubmissionResult{
		Success:      true,
		SubmissionID: response.ID,
		Message:      fmt.Sprintf("Submission received: id = s%s, time = %s", response.ID, submitTime.Format("15:04:05")),
		URL:          resultURL,
	}, nil
}

// submitForPrinting 提交文件进行打印
func (d *DOMjudgeSubmitter) submitForPrinting(ctx context.Context, options *submitter.SubmissionOptions, files map[string]io.Reader) (*submitter.SubmissionResult, error) {
	// DOMjudge只允许打印单个文件
	if len(options.Files) != 1 {
		return nil, fmt.Errorf("only one file can be printed at a time")
	}

	// 获取文件名和内容
	var filename string
	var reader io.Reader
	for name, r := range files {
		filename = name
		reader = r
		break
	}

	// 读取文件内容
	fileContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 准备请求数据
	data := map[string]interface{}{
		"original_name": filename,
		"file_contents": base64.StdEncoding.EncodeToString(fileContent),
	}

	// 添加可选字段
	languages, err := d.GetLanguages(ctx, options.ContestID)
	if err == nil && options.LanguageID != "" {
		for _, lang := range languages {
			if strings.EqualFold(lang.ID, options.LanguageID) {
				data["language"] = lang.Name
				break
			}
		}
	}

	if options.EntryPoint != "" {
		data["entry_point"] = options.EntryPoint
	}

	// 序列化请求体
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal print data: %w", err)
	}

	// 准备请求
	req, err := d.createRequest(ctx, http.MethodPost, "printing/team", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 设置认证
	if err := d.setAuthentication(req, options.Credentials); err != nil {
		return nil, err
	}

	// 发送请求
	var response DOMjudgePrintResponse
	err = d.doRequest(req, &response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return &submitter.SubmissionResult{
			Success: false,
			Message: fmt.Sprintf("Print job failed: %s", response.Output),
		}, nil
	}

	return &submitter.SubmissionResult{
		Success: true,
		Message: "Print job successfully submitted",
	}, nil
}

// InferProblemAndLanguage 从文件名推断问题和语言
func (d *DOMjudgeSubmitter) InferProblemAndLanguage(ctx context.Context, contestID string, filename string) (problemID string, languageID string, err error) {
	// 从文件名中提取基本名称和扩展名
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	baseName := strings.TrimSuffix(base, ext)

	// 扩展名处理
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
	}

	// 尝试根据扩展名推断语言
	languages, err := d.GetLanguages(ctx, contestID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get languages: %w", err)
	}

	// 查找匹配的语言
	for _, lang := range languages {
		for _, langExt := range lang.Extensions {
			if strings.EqualFold(langExt, ext) {
				languageID = lang.ID
				break
			}
		}
		if languageID != "" {
			break
		}
	}

	// 尝试根据基本名称推断问题
	problems, err := d.GetProblems(ctx, contestID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get problems: %w", err)
	}

	// 查找匹配的问题
	for _, problem := range problems {
		if strings.EqualFold(problem.Label, baseName) || strings.EqualFold(problem.ID, baseName) {
			problemID = problem.ID
			break
		}
	}

	return problemID, languageID, nil
}

// InferEntryPoint 推断入口点
func (d *DOMjudgeSubmitter) InferEntryPoint(ctx context.Context, languageInfo *submitter.LanguageInfo, filename string) (string, error) {
	if languageInfo == nil || !languageInfo.EntryPointRequired {
		return "", nil
	}

	// 获取文件基本名称（不含扩展名）
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	baseName := strings.TrimSuffix(base, ext)

	// 根据语言类型生成入口点
	switch languageInfo.Name {
	case "Java":
		return baseName, nil
	case "Kotlin":
		return d.kotlinBaseEntryPoint(baseName) + "Kt", nil
	case "Python 3":
		// 对于Python，入口点应该是带扩展名的模块名
		return baseName + ext, nil
	default:
		return "", nil
	}
}

// kotlinBaseEntryPoint 为Kotlin生成正确的入口点
func (d *DOMjudgeSubmitter) kotlinBaseEntryPoint(filebase string) string {
	if filebase == "" {
		return "_"
	}

	var result strings.Builder

	for i, c := range filebase {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			if i == 0 {
				// 首字符需要大写
				if c >= 'a' && c <= 'z' {
					result.WriteRune(c - 'a' + 'A')
				} else {
					result.WriteRune(c)
				}
			} else {
				result.WriteRune(c)
			}
		} else {
			result.WriteRune('_')
		}
	}

	// 如果首字符不是字母或数字，添加前缀
	if len(filebase) > 0 && !((filebase[0] >= 'a' && filebase[0] <= 'z') ||
		(filebase[0] >= 'A' && filebase[0] <= 'Z') ||
		(filebase[0] >= '0' && filebase[0] <= '9')) {
		return "_" + result.String()
	}

	return result.String()
}

// apiGet 发送GET请求到API端点并解析响应
func (d *DOMjudgeSubmitter) apiGet(ctx context.Context, endpoint string, result interface{}) error {
	req, err := d.createRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	return d.doRequest(req, result)
}

// createRequest 创建API请求
func (d *DOMjudgeSubmitter) createRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	apiPath := fmt.Sprintf("api/%s%s", d.apiVersion, endpoint)
	requestURL, err := url.JoinPath(d.baseURL, apiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置通用头
	req.Header.Set("User-Agent", "domjudge-submit-client/go")
	if method == http.MethodPost && body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// doRequest 执行HTTP请求并解析响应
func (d *DOMjudgeSubmitter) doRequest(req *http.Request, result interface{}) error {
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode >= 300 {
		// 尝试解析错误信息
		var errorResponse struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResponse); err == nil && (errorResponse.Message != "" || errorResponse.Error != "") {
			if errorResponse.Message != "" {
				return fmt.Errorf("API error (code %d): %s", resp.StatusCode, errorResponse.Message)
			}
			return fmt.Errorf("API error (code %d): %s", resp.StatusCode, errorResponse.Error)
		}

		// 如果无法解析错误信息，返回状态码
		return fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	// 解析JSON响应
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// setAuthentication 设置请求的认证信息
func (d *DOMjudgeSubmitter) setAuthentication(req *http.Request, credentials submitter.Credentials) error {
	if credentials.Token != "" {
		// 使用令牌认证
		req.Header.Set("Authorization", "Bearer "+credentials.Token)
	} else if credentials.Username != "" && credentials.Password != "" {
		// 使用基本认证
		req.SetBasicAuth(credentials.Username, credentials.Password)
	} else {
		// 尝试从netrc文件读取凭据
		netrcFile := filepath.Join(os.Getenv("HOME"), ".netrc")
		if _, err := os.Stat(netrcFile); err == nil {
			// 这里需要实现netrc文件解析
			// 简化起见，这里不实现完整的netrc解析
			return fmt.Errorf("no credentials provided and .netrc support not implemented")
		}
		return fmt.Errorf("no authentication credentials provided")
	}
	return nil
}

// GetProblemByID 通过ID或标签查找问题
func (d *DOMjudgeSubmitter) GetProblemByID(ctx context.Context, contestID, problemID string) (*submitter.ProblemInfo, error) {
	problems, err := d.GetProblems(ctx, contestID)
	if err != nil {
		return nil, err
	}

	lowercaseID := strings.ToLower(problemID)
	for _, problem := range problems {
		if strings.ToLower(problem.ID) == lowercaseID || strings.ToLower(problem.Label) == lowercaseID {
			return &problem, nil
		}
	}

	return nil, fmt.Errorf("problem '%s' not found", problemID)
}

// GetLanguageByID 通过ID或扩展名查找语言
func (d *DOMjudgeSubmitter) GetLanguageByID(ctx context.Context, contestID, languageID string) (*submitter.LanguageInfo, error) {
	languages, err := d.GetLanguages(ctx, contestID)
	if err != nil {
		return nil, err
	}

	lowercaseID := strings.ToLower(languageID)
	for _, language := range languages {
		if strings.ToLower(language.ID) == lowercaseID {
			return &language, nil
		}

		// 检查扩展名
		for _, ext := range language.Extensions {
			if strings.ToLower(ext) == lowercaseID {
				return &language, nil
			}
		}
	}

	return nil, fmt.Errorf("language '%s' not found", languageID)
}

// GetContestByID 通过ID或短名称查找竞赛
func (d *DOMjudgeSubmitter) GetContestByID(ctx context.Context, contestID string) (*submitter.ContestInfo, error) {
	contests, err := d.GetContests(ctx)
	if err != nil {
		return nil, err
	}

	lowercaseID := strings.ToLower(contestID)
	for _, contest := range contests {
		if strings.ToLower(contest.ID) == lowercaseID || strings.ToLower(contest.ShortName) == lowercaseID {
			return &contest, nil
		}
	}

	return nil, fmt.Errorf("contest '%s' not found", contestID)
}

// ValidateFiles 验证提交的文件
func (d *DOMjudgeSubmitter) ValidateFiles(files []string) ([]string, error) {
	validFiles := make([]string, 0, len(files))
	warnings := make([]string, 0)

	for _, filename := range files {
		// 检查文件是否存在
		fileInfo, err := os.Stat(filename)
		if err != nil {
			return nil, fmt.Errorf("file '%s' not found or not accessible", filename)
		}

		// 检查是否是常规文件
		if !fileInfo.Mode().IsRegular() {
			warnings = append(warnings, fmt.Sprintf("'%s' is not a regular file", filename))
		}

		// 检查文件是否为空
		if fileInfo.Size() == 0 {
			warnings = append(warnings, fmt.Sprintf("'%s' is empty", filename))
		}

		// 检查文件修改时间
		fileAge := time.Since(fileInfo.ModTime()).Minutes()
		if fileAge > 5 { // 5分钟是Python脚本中的warn_mtime_minutes
			warnings = append(warnings, fmt.Sprintf("'%s' has not been modified for %.0f minutes", filename, fileAge))
		}

		// 如果文件有效，添加到列表
		if !contains(validFiles, filename) {
			validFiles = append(validFiles, filename)
		}
	}

	// 如果有警告，打印出来
	for _, warning := range warnings {
		fmt.Printf("WARNING: %s!\n", warning)
	}

	return validFiles, nil
}

// contains 检查切片是否包含特定值
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

//func (d *DOMjudgeSubmitter) ValidateCredentials(ctx context.Context, credentials submitter.Credentials) error {
//	//TODO implement me
//	panic("implement me")
//}
