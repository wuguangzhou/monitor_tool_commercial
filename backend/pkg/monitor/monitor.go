package monitor

import (
	"net/http"
	"time"
)

// HTTPMonitor HTTP/HTTPS监控（核心检测逻辑）
func HTTPMonitor(url string) (status int, responeTime int, err error) {
	// 设置请求超时时间（5秒，避免监控卡住）
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	// 记录开始时间（用于计算响应时间）
	startTime := time.Now()

	//发送GET请求（检测URL可用性）
	resp, err := client.Get(url)
	if err != nil {
		return 2, 0, err
	}
	defer resp.Body.Close()

	// 计算响应时间
	responeTime = int(time.Since(startTime).Milliseconds())

	// 判断状态码（200-399视为正常，其他视为宕机）
	if resp.StatusCode >= 200 && resp.StatusCode <= 400 {
		return 1, responeTime, nil
	} else {
		return 2, responeTime, nil
	}
}
