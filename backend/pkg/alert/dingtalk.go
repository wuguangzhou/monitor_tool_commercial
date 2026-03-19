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
	"time"
)

// 钉钉机器人配置
const (
	DingTalkWebhook = "https://oapi.dingtalk.com/robot/send?access_token=c108d3f19d0ddbf6b15eb2d9931500669c58501c91f8a9285c1aafe00c295642"
	//开启加签
	DingTalkSecret = "SECfb3ef041d27afe5cb903ee235893c3692107da946d0231ba1ebc335a04012638"
	// 安全关键词（必须包含在消息里）
	DingTalkKeyword = "告警"
)

// DingTalkResp 钉钉返回的响应结构体，用于解析错误信息
type DingTalkResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// 发送告警到钉钉群
func SendDingTalkAlert(title, content string) (bool, error) {
	// 拼接消息（确保包含关键词）
	fullContent := fmt.Sprintf("告警：%s\n%s", title, content)

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
	reqURL, err := url.Parse(DingTalkWebhook)
	if err != nil {
		return false, fmt.Errorf("解析webhook地址失败:%s", err)
	}
	params := reqURL.Query()

	// 加签逻辑（核心修复：时间戳统一为毫秒级）
	if DingTalkSecret != "" {
		timestamp := time.Now().UnixMilli() // 毫秒级时间戳
		sign, err := genDingTalkSign(DingTalkSecret, timestamp)
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
