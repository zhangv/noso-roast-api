package baidu

import (
	"bytes"
	"encoding/json"
	"hahacking/common/logger"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheFile = "/tmp/baidu_token.json"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type CachedToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

/*
参考：短文本在线合成 https://cloud.baidu.com/doc/SPEECH/s/mlbxh7xie
*/

func Text2Audio(text string) ([]byte, error) {
	token, err := GetToken(false)
	if err != nil {
		return nil, err
	}
	url := "https://tsn.baidu.com/text2audio?tex=" + text + "&tok=" + token + "&cuid=1234567890&ctp=1&lan=zh&aue=3"
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type RecognizeResponse struct {
	// 根据百度API的实际返回结构定义字段
	// 这里只是一个示例，需要根据实际API响应调整
	Result []struct {
		Name  string  `json:"name"`
		Score float64 `json:"score"`
	} `json:"result"`
}

func Recognize(imageBase64 string) (*RecognizeResponse, error) {
	// 获取token
	token, err := GetToken(true)
	if err != nil {
		return nil, err
	}

	// 构建请求URL
	url := "https://aip.baidubce.com/rest/2.0/image-classify/v1/classify/ingredient?access_token=" + token
	logger.Sugar.Info("url", url)
	formData := neturl.Values{}
	formData.Set("image", imageBase64)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// logger.Sugar.Info("req", req.URL.String())
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logger.Sugar.Info("resp", resp)
	logger.Sugar.Info("resp.body", resp.Body)
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Sugar.Error("io.ReadAll", err)
		return nil, err
	}
	logger.Sugar.Info("body", string(body))
	// 解析响应
	var recognizeResp RecognizeResponse
	err = json.Unmarshal(body, &recognizeResp)
	if err != nil {
		logger.Sugar.Error("json.Unmarshal", err)
		return nil, err
	}
	logger.Sugar.Info("recognizeResp", recognizeResp)

	return &recognizeResp, nil
}

func GetToken(cache bool) (string, error) {
	// 尝试从缓存获取token
	if cache {
		if token := getCachedToken(); token != "" {
			return token, nil
		}
	}

	// 如果没有缓存或已过期，获取新token
	tokenResp, err := fetchToken()

	if err != nil {
		return "", err
	}

	// 缓存新token
	if cache {
		err = cacheToken(tokenResp.AccessToken, tokenResp.ExpiresIn)
		if err != nil {
			return "", err
		}
	}

	return tokenResp.AccessToken, nil
}

func getCachedToken() string {
	// 检查缓存文件是否存在
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return ""
	}

	// 读取缓存文件
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}

	var cached CachedToken
	if err := json.Unmarshal(data, &cached); err != nil {
		return ""
	}

	// 检查token是否过期
	if cached.ExpiresAt <= time.Now().Unix() {
		return ""
	}

	return cached.Token
}

func cacheToken(token string, expiresIn int) error {
	cached := CachedToken{
		Token:     token,
		ExpiresAt: time.Now().Unix() + int64(expiresIn) - 60, // 提前60秒过期
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

func fetchToken() (*TokenResponse, error) {
	//hardcode:
	//佳泰电子秤
	//116157830
	apiKey := "TWSV3mFtv3TMTpULqBM12Jwb"
	secretKey := "l7BMRXdQEWYZ3vNQxCKp4SvXnw4lAdRp"

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := "https://aip.baidubce.com/oauth/2.0/token?client_id=" + apiKey + "&client_secret=" + secretKey + "&grant_type=client_credentials"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
