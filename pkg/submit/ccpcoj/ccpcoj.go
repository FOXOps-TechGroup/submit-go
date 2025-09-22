package ccpcoj

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/FOXOps-TechGroup/submit-go/pkg/submit"
)

type CCPCOJ struct {
	account string
	passwd  string
	cid     uint
	//phpSession 是ccpcoj鉴权的cookie
	phpSession string
}

func NewCCPCOJSubmitter(account string, passwd string) *CCPCOJ {
	return &CCPCOJ{account: account, passwd: passwd}
}

func (C *CCPCOJ) Name() string {
	return "ccpcoj"
}

func (C *CCPCOJ) Initialize(ctx context.Context, baseURL string) error {
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

func (C *CCPCOJ) GetContests(ctx context.Context) ([]submit.ContestInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (C *CCPCOJ) GetProblems(ctx context.Context, contestID string) ([]submit.ProblemInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (C *CCPCOJ) GetLanguages(ctx context.Context, contestID string) ([]submit.LanguageInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (C *CCPCOJ) ValidateSubmission(ctx context.Context, options *submit.SubmissionOptions) error {
	//TODO implement me
	panic("implement me")
}

func (C *CCPCOJ) Submit(ctx context.Context, options *submit.SubmissionOptions, files map[string]io.Reader) (*submit.SubmissionResult, error) {
	//TODO implement me
	panic("implement me")
}

func (C *CCPCOJ) InferProblemAndLanguage(ctx context.Context, contestID string, filename string) (problemID string, languageID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (C *CCPCOJ) InferEntryPoint(ctx context.Context, languageInfo *submit.LanguageInfo, filename string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (C *CCPCOJ) ValidateCredentials(ctx context.Context, credentials submit.Credentials) error {
	//TODO implement me
	panic("implement me")
}
