package alert

import (
	"backend/internal/model"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// SMTP核心配置 - 你的163邮箱配置已保留，无需修改
const (
	SmtpHost   = "smtp.163.com"        // 163邮箱SMTP地址
	SmtpPort   = "465"                 // 587端口（STARTTLS）
	SmtpUser   = "18085594585@163.com" // 你的发件人邮箱
	SmtpPass   = "TTfK5kGtVhH4LeNN"    // 163邮箱SMTP授权码
	SenderName = "zy监控系统告警通知"          // 发件人显示名称
)

// SendEmailAlert 发送邮件告警（修复587端口EOF错误）
// email: 收件人邮箱（多个邮箱用英文逗号分隔）
// subject: 邮件标题
// content: 邮件内容（支持HTML/纯文本）
func SendEmailAlert(email, subject, content string) (bool, error) {
	// 1. 参数校验：避免空收件人/空内容
	if strings.TrimSpace(email) == "" {
		return false, fmt.Errorf("收件人邮箱不能为空")
	}
	if strings.TrimSpace(subject) == "" {
		subject = "系统告警通知" // 默认标题
	}

	// 2. 构建邮件头部（解决中文乱码问题）
	from := fmt.Sprintf("%s<%s>", SenderName, SmtpUser)
	header := map[string]string{
		"From":         from,
		"To":           email,
		"Subject":      fmt.Sprintf("=?UTF-8?B?%s?=", base64Encode(subject)), // 中文标题Base64编码
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
		"Date":         time.Now().Format(time.RFC1123Z),
	}

	// 3. 拼接完整邮件内容（头部+正文）
	var msg strings.Builder
	for k, v := range header {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n") // 头部和正文之间必须空一行
	msg.WriteString(content)

	// 4. 配置SMTP连接基础信息
	addr := fmt.Sprintf("%s:%s", SmtpHost, SmtpPort)
	auth := smtp.PlainAuth("", SmtpUser, SmtpPass, SmtpHost)
	//var err error

	// 根据端口选择加密方式
	if SmtpPort == "465" {
		// 465端口：SSL/TLS加密连接（保留原有逻辑，兼容备用）
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         SmtpHost,
		}
		conn, dialErr := tls.Dial("tcp", addr, tlsConfig)
		if dialErr != nil {
			return false, fmt.Errorf("TLS连接失败: %v", dialErr)
		}
		defer conn.Close()

		client, clientErr := smtp.NewClient(conn, SmtpHost)
		if clientErr != nil {
			return false, fmt.Errorf("创建SMTP客户端失败: %v", clientErr)
		}
		defer client.Quit()

		// 认证
		if authErr := client.Auth(auth); authErr != nil {
			return false, fmt.Errorf("SMTP认证失败: %v", authErr)
		}

		// 设置发件人
		if mailErr := client.Mail(SmtpUser); mailErr != nil {
			return false, fmt.Errorf("设置发件人失败: %v", mailErr)
		}

		// 设置多个收件人
		recipients := strings.Split(email, ",")
		for _, rec := range recipients {
			rec = strings.TrimSpace(rec)
			if rec == "" {
				continue
			}
			if rcptErr := client.Rcpt(rec); rcptErr != nil {
				return false, fmt.Errorf("收件人%s设置失败: %v", rec, rcptErr)
			}
		}

		// 写入并发送邮件内容
		w, dataErr := client.Data()
		if dataErr != nil {
			return false, fmt.Errorf("获取邮件写入器失败: %v", dataErr)
		}
		defer w.Close()
		_, writeErr := w.Write([]byte(msg.String()))
		if writeErr != nil {
			return false, fmt.Errorf("写入邮件内容失败: %v", writeErr)
		}
	} else {
		// 587端口：显式STARTTLS流程（核心修复：替换原简化逻辑）
		// 步骤1：建立纯文本连接
		client, dialErr := smtp.Dial(addr)
		if dialErr != nil {
			return false, fmt.Errorf("连接SMTP服务器失败: %v", dialErr)
		}
		defer client.Quit() // 确保连接最终关闭

		// 步骤2：关键 - 启动STARTTLS加密（587端口必须先加密再认证）
		tlsConfig := &tls.Config{
			ServerName:         SmtpHost, // 必须指定服务器名称，匹配证书
			InsecureSkipVerify: false,    // 生产环境保持false，禁用不安全验证
		}
		if tlsErr := client.StartTLS(tlsConfig); tlsErr != nil {
			return false, fmt.Errorf("启动STARTTLS加密失败: %v", tlsErr)
		}

		// 步骤3：SMTP认证（加密后传输，避免明文）
		if authErr := client.Auth(auth); authErr != nil {
			return false, fmt.Errorf("SMTP认证失败（授权码错误？）: %v", authErr)
		}

		// 步骤4：设置发件人
		if mailErr := client.Mail(SmtpUser); mailErr != nil {
			return false, fmt.Errorf("设置发件人失败: %v", mailErr)
		}

		// 步骤5：设置多个收件人
		recipients := strings.Split(email, ",")
		for _, rec := range recipients {
			rec = strings.TrimSpace(rec)
			if rec == "" {
				continue
			}
			if rcptErr := client.Rcpt(rec); rcptErr != nil {
				return false, fmt.Errorf("收件人%s设置失败: %v", rec, rcptErr)
			}
		}

		// 步骤6：写入并发送邮件内容
		w, dataErr := client.Data()
		if dataErr != nil {
			return false, fmt.Errorf("获取邮件写入器失败: %v", dataErr)
		}
		defer w.Close()
		_, writeErr := w.Write([]byte(msg.String()))
		if writeErr != nil {
			return false, fmt.Errorf("写入邮件内容失败: %v", writeErr)
		}
	}

	// 发送成功日志
	fmt.Printf("【邮件发送成功】收件人：%s | 标题：%s | 时间：%s\n",
		email, subject, time.Now().Format("2006-01-02 15:04:05"))
	return true, nil
}

// base64Encode 对字符串进行Base64编码，解决邮件标题中文乱码
func base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// 构建恢复通知内容
func BuildRecoveryContent(monitor *model.Monitor) string {
	return fmt.Sprintf(
		"【监控工具恢复通知】\n您的监控项「%s」（%s）于%s已恢复正常状态，请确认！",
		monitor.Name,
		monitor.Url,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}
