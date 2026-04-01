package alert

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 注意：钉钉机器人参数不应硬编码在后端代码里。
// 新实现通过 SendDingTalkAlertWithConfig 由调用方传入 webhook/secret/keyword（通常来自 alert_config）。

// DingTalkResp 钉钉返回的响应结构体，用于解析错误信息
type DingTalkResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendDingTalkAlertWithConfig 发送告警到钉钉群（参数来自配置）
// webhook: 机器人 webhook（必须）
// secret:  机器人加签 secret（可选）
// keyword: 安全关键词（可选；如群里配置了“关键词校验”，必须出现在消息中）
func SendDingTalkAlertWithConfig(webhook, secret, keyword, title, content string) (bool, error) {
	if webhook == "" {
		return false, fmt.Errorf("dingtalk webhook 不能为空")
	}

	// 拼接消息：确保包含 keyword（若提供）
	fullContent := fmt.Sprintf("%s\n%s", title, content)
	if keyword != "" && !strings.Contains(fullContent, keyword) {
		fullContent = fmt.Sprintf("%s\n%s", keyword, fullContent)
	}

	// 构建请求体（纯文本格式，更稳定）
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fullContent,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("json marshal err:%s", err)
	}

	// 生成签名并拼接URL
	reqURL, err := url.Parse(webhook)
	if err != nil {
		return false, fmt.Errorf("解析webhook地址失败:%s", err)
	}
	params := reqURL.Query()

	// 加签逻辑（核心修复：时间戳统一为毫秒级）
	if secret != "" {
		timestamp := time.Now().UnixMilli() // 毫秒级时间戳
		sign, err := genDingTalkSign(secret, timestamp)
		if err != nil {
			return false, fmt.Errorf("dingtalk sign err:%s", err)
		}
		// 将timestamp和sign添加到URL参数中
		params.Add("timestamp", fmt.Sprintf("%d", timestamp))
		params.Add("sign", sign)
		reqURL.RawQuery = params.Encode()
	}

	// 发送HTTP POST请求
	resp, err := http.Post(reqURL.String(), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("请求钉钉失败:%s", err)
	}
	defer resp.Body.Close()

	// 读取响应体（关键：解析钉钉的具体错误）
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("读取钉钉响应失败:%s", err)
	}

	// 校验HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("钉钉返回非200：%s，响应内容：%s", resp.Status, string(respBody))
	}

	// 解析钉钉的业务错误码
	var dingResp DingTalkResp
	err = json.Unmarshal(respBody, &dingResp)
	if err != nil {
		return false, fmt.Errorf("解析钉钉响应JSON失败:%s，响应内容：%s", err, string(respBody))
	}
	if dingResp.ErrCode != 0 {
		return false, fmt.Errorf("钉钉返回业务错误：errcode=%d, errmsg=%s", dingResp.ErrCode, dingResp.ErrMsg)
	}

	fmt.Printf("【钉钉告警发送成功】标题：%s | 时间：%s\n", title, time.Now().Format("2006-01-02 15:04:05"))
	return true, nil
}

// SendDingTalkAlert 兼容旧接口（不再推荐使用）
// 旧版本曾硬编码 webhook/secret/keyword；这里保留函数签名避免其它代码编译失败。
// 你应在调用侧使用 SendDingTalkAlertWithConfig(webhook, secret, keyword, title, content)。
func SendDingTalkAlert(title, content string) (bool, error) {
	return false, fmt.Errorf("SendDingTalkAlert 已废弃：请改用 SendDingTalkAlertWithConfig")
}

// 修复：将timestamp作为参数传入，保证签名和URL中使用的是同一个值
func genDingTalkSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(stringToSign))
	if err != nil {
		return "", fmt.Errorf("计算签名失败:%s", err)
	}
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return sign, nil
}
