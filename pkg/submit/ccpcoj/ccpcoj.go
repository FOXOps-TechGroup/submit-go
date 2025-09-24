package ccpcoj

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	url2 "net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	submitter "github.com/FOXOps-TechGroup/submit-go/pkg/submit"
)

type CCPCOJ struct {
	account string
	passwd  string
	cid     uint
	//phpSession 是ccpcoj鉴权的cookie
	phpSession string
	//baseURL 入口URL
	baseURL string
}

func NewCCPCOJSubmitter(account string, passwd string) *CCPCOJ {
	return &CCPCOJ{account: account, passwd: passwd}
}

func (C *CCPCOJ) Name() string {
	return "ccpcoj"
}

func (C *CCPCOJ) Initialize(ctx context.Context, baseURL string) error {
	C.baseURL = baseURL
	//第一步，获取默认的cid,默认是列表中的第一场比赛
	contestListURL := baseURL + "/cpcsys/contest/contest_list_ajax?search=&sort=contest_id&order=desc"

	// 创建带超时的HTTP客户端
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", contestListURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("非200状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析JSON数据
	var contests []struct {
		ContestID uint `json:"contest_id"`
	}
	if err := json.Unmarshal(body, &contests); err != nil {
		return fmt.Errorf("JSON解析失败: %w", err)
	}

	// 检查数组是否为空
	if len(contests) == 0 {
		return fmt.Errorf("未找到竞赛数据")
	}

	// 获取第一个比赛的ID
	C.cid = contests[0].ContestID

	//第二步：获取php_session，这是鉴权用的
	authURL := baseURL + "/cpcsys/contest/contest_auth_ajax"

	authBody := map[string]any{
		"cid":      C.cid,
		"team_id":  C.account,
		"password": C.passwd,
	}

	reqBody, _ := json.Marshal(authBody)

	req, err = http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(reqBody))

	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("非200状态码: %d", resp.StatusCode)
	}

	//获取cookie
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "PHPSESSID" {
			C.phpSession = cookie.Value
			break
		}
	}
	if C.phpSession == "" {
		return fmt.Errorf("Cannot get Cookie")
	}
	return nil
}

func (C *CCPCOJ) GetContests(ctx context.Context) ([]submitter.ContestInfo, error) {
	//第一步，获取默认的cid,默认是列表中的第一场比赛
	contestListURL := C.baseURL + "/cpcsys/contest/contest_list_ajax?search=&sort=contest_id&order=desc"

	// 创建带超时的HTTP客户端
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", contestListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("非200状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var contests []struct {
		ContestID uint   `json:"contest_id"`
		Title     string `json:"title"`
	}
	if err := json.Unmarshal(body, &contests); err != nil {
		return nil, err
	}

	contestInfos := make([]submitter.ContestInfo, len(contests))
	for i, contest := range contests {
		contestInfos[i] = submitter.ContestInfo{
			ID:        strconv.Itoa(int(contest.ContestID)),
			Name:      contest.Title,
			ShortName: contest.Title, //没有简称，所以统一了
		}
	}
	return contestInfos, nil
}

func (C *CCPCOJ) GetProblems(ctx context.Context, contestID string) ([]submitter.ProblemInfo, error) {
	problemURL := C.baseURL + "/cpcsys/contest/problemset_ajax?cid=" + strconv.Itoa(int(C.cid))
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", problemURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: C.phpSession})
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var problems []struct {
		ProblemID uint   `json:"problem_id"`
		Title     string `json:"title"`
		ID        string `json:"problem_id_show"`
	}
	if err := json.Unmarshal(body, &problems); err != nil {
		return nil, err
	}
	problemInfos := make([]submitter.ProblemInfo, len(problems))
	for i, problem := range problems {
		problemInfos[i] = submitter.ProblemInfo{
			ID:    strconv.Itoa(int(problem.ProblemID)),
			Name:  problem.Title,
			Label: problem.ID,
		}
	}
	return problemInfos, nil
}

// GetLanguages
// 暂时想不到什么好的实现方法，先咕
// 先写死，没有暴露出来的API
func (C *CCPCOJ) GetLanguages(ctx context.Context, contestID string) ([]submitter.LanguageInfo, error) {
	languages := []submitter.LanguageInfo{
		//See：
		//https://github.com/CSGrandeur/CCPCOJ/blob/e470f60300536873f64e9b571f409bd3740c8c1a/ojweb/application/extra/CsgojConfig.php#L28
		{
			"0", //C
			"C",
			[]string{
				"c",
			},
			false,
			"",
		},
		{
			"1", // CPP
			"C++",
			[]string{
				"cc",
				"cpp",
				"cxx",
				"c++",
			},
			false,
			"",
		},
		{
			"3",
			"Java",
			[]string{
				"java",
			},
			//实际上对于CCPCOJ，这两个东西没有一点用……
			true,
			"Main",
		},
		{
			"6",
			"Python3",
			[]string{
				"py",
			},
			true,
			"Main",
		},
	}
	return languages, nil
}

func (C *CCPCOJ) ValidateSubmission(ctx context.Context, options *submitter.SubmissionOptions) error {
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
		languages, err := C.GetLanguages(ctx, options.ContestID)
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

func (C *CCPCOJ) Submit(ctx context.Context,
	options *submitter.SubmissionOptions,
	files map[string]io.Reader) (*submitter.SubmissionResult, error) {
	// 验证选项
	err := C.ValidateSubmission(ctx, options)
	if err != nil {
		return nil, err
	}

	// 根据模式选择提交或打印
	if options.PrintMode {
		return C.submitForPrinting(ctx, options, files)
	} else {
		return C.submitForJudging(ctx, options, files)
	}
}

func (C *CCPCOJ) submitForPrinting(ctx context.Context,
	options *submitter.SubmissionOptions,
	files map[string]io.Reader) (*submitter.SubmissionResult, error) {
	//CCPCOJ 也只允许打印单个文件
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
	_ = filename
	// 读取文件内容
	fileContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 准备请求数据
	data := url2.Values{
		"cid":    {options.ContestID},
		"source": {string(fileContent)},
	}

	//构造URL
	url := C.baseURL + "/cpcsys/contest/print_code_ajax"

	//发起请求
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: C.phpSession})
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	//这一步不会有任何响应。
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to submit code: %s", resp.Status)
	}
	//接下来要拉请求……
	url = C.baseURL + "/cpcsys/contest/print_status_ajax?" +
		"cid=" + options.ContestID +
		"&sort=print_status_show&order=desc&offset=0&limit=20&team_id=" +
		C.account +
		"&room_ids=&print_status=-1"

	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: C.phpSession})
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to submit code: %s", resp.Status)
	}
	var response struct {
		Rows []struct {
			PrintID     int    `json:"print_id"`
			ContestID   int    `json:"contest_id"`
			TeamID      string `json:"team_id"`
			Source      string `json:"source"`
			PrintStatus int    `json:"print_status"`
			InDate      string `json:"in_date"`
			IP          string `json:"ip"`
			CodeLength  int    `json:"code_length"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &submitter.SubmissionResult{
		Success:      true,
		SubmissionID: strconv.Itoa(response.Rows[0].PrintID),
		Message:      "Success",
		URL: C.baseURL + "/cpcsys/contest/print_status?cid=" + options.ContestID +
			"#team_id=" + C.account,
	}, nil
}

func (C *CCPCOJ) submitForJudging(ctx context.Context,
	options *submitter.SubmissionOptions,
	files map[string]io.Reader) (*submitter.SubmissionResult, error) {

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加表单字段
	_ = writer.WriteField("pid", options.ProblemID)
	_ = writer.WriteField("language", options.LanguageID)
	_ = writer.WriteField("cid", options.ContestID)

	// 添加文件
	for name, reader := range files {
		fileWriter, err := writer.CreateFormFile("source", filepath.Base(name))
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

	client := &http.Client{Timeout: 10 * time.Second}
	url := C.baseURL + "/cpcsys/contest/submit_ajax"
	req, _ := http.NewRequestWithContext(ctx, "POST", url, &buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: C.phpSession})
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to submit code: %s", resp.Status)
	}
	url = C.baseURL + "/cpcsys/contest/status_ajax?" +
		"cid=" + options.ContestID +
		"&sort=solution_id_show&order=desc&offset=0&limit=20" +
		"&problem_id=" + options.ProblemID +
		"&user_id=" + C.account +
		"&solution_id=&language=-1&result=-1"
	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.AddCookie(&http.Cookie{Name: "PHPSESSID", Value: C.phpSession})
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to submit code: %s", resp.Status)
	}
	var response struct {
		Rows []struct {
			SolutionID int `json:"solution_id"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &submitter.SubmissionResult{
		Success:      true,
		SubmissionID: strconv.Itoa(response.Rows[0].SolutionID),
		Message:      "Success",
		URL: C.baseURL + "/cpcsys/contest/status?cid=" + options.ContestID +
			"#solution_id=" + strconv.Itoa(response.Rows[0].SolutionID),
	}, nil
}

// InferProblemAndLanguage
// 感谢domjudge的思路
func (C *CCPCOJ) InferProblemAndLanguage(ctx context.Context, contestID string, filename string) (problemID string, languageID string, err error) {
	// 从文件名中提取基本名称和扩展名
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	baseName := strings.TrimSuffix(base, ext)

	// 扩展名处理
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
	}

	// 尝试根据扩展名推断语言
	languages, err := C.GetLanguages(ctx, contestID)
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
	problems, err := C.GetProblems(ctx, contestID)
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

// InferEntryPoint
// CCPCOJ 不支持指定EntryPoint，所以，对，就返回nil
func (C *CCPCOJ) InferEntryPoint(ctx context.Context, languageInfo *submitter.LanguageInfo, filename string) (string, error) {
	return "", nil
}
