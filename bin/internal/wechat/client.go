package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"skills/bin/internal/db"
	"skills/bin/internal/proxy"
)

const (
	apiBase = "https://api.weixin.qq.com"

	configKeyAppID     = "wechat_appid"
	configKeyAppSecret = "wechat_appsecret"
)

// Client 微信公众平台 API 客户端。
type Client struct {
	http       *http.Client
	mu         sync.Mutex
	token      string
	tokenExpAt time.Time
}

// NewClient 创建微信 API 客户端。
func NewClient() *Client {
	transport := &http.Transport{
		Proxy: proxy.ProxyFromEnvironment(),
	}
	return &Client{
		http: &http.Client{Transport: transport},
	}
}

// getAccessToken 获取 access_token，缓存过期自动刷新。
func (c *Client) getAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpAt) {
		return c.token, nil
	}

	appID, err := db.ConfigGet(configKeyAppID)
	if err != nil {
		return "", fmt.Errorf("get appid: %w", err)
	}
	appSecret, err := db.ConfigGet(configKeyAppSecret)
	if err != nil {
		return "", fmt.Errorf("get appsecret: %w", err)
	}
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("AppID or AppSecret not configured. Run `wechat login` first")
	}

	url := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		apiBase, appID, appSecret)

	resp, err := c.http.Get(url)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	var r AccessTokenResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	if r.AccessToken == "" {
		var errResp ErrorResponse
		json.Unmarshal(body, &errResp)
		return "", fmt.Errorf("token error (errcode=%d): %s", errResp.ErrCode, errResp.ErrMsg)
	}

	c.token = r.AccessToken
	// 提前 5 分钟过期，留有余量
	c.tokenExpAt = time.Now().Add(time.Duration(r.ExpiresIn-300) * time.Second)

	return c.token, nil
}

// get 发起带 access_token 的 GET 请求。
func (c *Client) get(path string) ([]byte, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s&access_token=%s", apiBase, path, token)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if err := checkError(body); err != nil {
		return nil, fmt.Errorf("get %s: %w", path, err)
	}

	return body, nil
}

// postJSON 发起带 access_token 的 POST JSON 请求。
func (c *Client) postJSON(path string, reqBody any) ([]byte, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s&access_token=%s", apiBase, path, token)
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("post %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if err := checkError(body); err != nil {
		return nil, fmt.Errorf("post %s: %w", path, err)
	}

	return body, nil
}

// postForm 发起带 access_token 的 multipart/form-data POST 请求。
// 用于上传素材文件。
func (c *Client) postForm(path, fieldName, filePath string) ([]byte, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}
	w.Close()

	url := fmt.Sprintf("%s%s&access_token=%s", apiBase, path, token)
	resp, err := c.http.Post(url, w.FormDataContentType(), &buf)
	if err != nil {
		return nil, fmt.Errorf("upload %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upload response: %w", err)
	}

	if err := checkError(body); err != nil {
		return nil, fmt.Errorf("upload %s: %w", path, err)
	}

	return body, nil
}

// checkError 检查微信 API 响应中是否有错误。
func checkError(body []byte) error {
	var e ErrorResponse
	if err := json.Unmarshal(body, &e); err != nil {
		return nil // 不是 JSON 或没有 errcode 字段
	}
	if e.ErrCode != 0 {
		return fmt.Errorf("API error (errcode=%d): %s", e.ErrCode, e.ErrMsg)
	}
	return nil
}
