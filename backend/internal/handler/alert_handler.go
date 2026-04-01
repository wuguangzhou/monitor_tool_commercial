package handler

import (
	"backend/internal/dao"
	"backend/internal/model"
	"backend/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 更新告警配置接口
func UpdateAlertConfigHandler(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	if !exists || userIdVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "用户未登录或认证信息无效",
		})
		return // 终止请求处理，避免panic
	}

	// 安全断言为int64
	userId, ok := userIdVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "用户ID类型错误",
		})
		return
	}

	//接收参数
	type AlertConfigParam struct {
		Email     string `json:"email" binding:"required"`
		AlertType int    `json:"alert_type" binding:"required"`
		IsEnable  int    `json:"is_enabled" binding:"required"`

		// 钉钉配置（当 alert_type=2 时由前端传入）
		DingTalkWebhook string `json:"dingtalk_webhook"`
		DingTalkSecret  string `json:"dingtalk_secret"`
		DingTalkKeyword string `json:"dingtalk_keyword"`
	}
	var param AlertConfigParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "参数错误：邮箱和告警方式必填",
			Data: nil,
		})
		return
	}

	//构建配置模型
	config := &model.AlertConfig{
		UserId:          userId,
		Email:           param.Email,
		AlertType:       param.AlertType,
		IsEnabled:       param.IsEnable,
		DingTalkWebhook: param.DingTalkWebhook,
		DingTalkSecret:  param.DingTalkSecret,
		DingTalkKeyword: param.DingTalkKeyword,
	}

	err := service.UpdateAlertConfig(config)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "更新告警配置失败：" + err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "告警配置更新成功",
		Data: config,
	})
}

// 获取用户告警配置接口
func GetAlertConfigHandler(c *gin.Context) {
	//获取当前登录用户ID
	// 注意：这里的 key 必须与 AuthMiddleware 中保持一致（AuthMiddleware 使用的是 "userId"）
	userIdVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, Response{
			Code: 401,
			Msg:  "未获取用户信息，请查询登录",
			Data: nil,
		})
		return
	}
	userId, ok := userIdVal.(int64)
	if !ok {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "用户ID格式错误",
			Data: nil,
		})
		return
	}

	config, err := service.GetAlertConfig(userId)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 200,
			Msg:  "用户暂无告警配置",
			Data: nil,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "查询成功",
		Data: config,
	})
}

// 获取用户告警记录列表
func GetAlertListHandler(c *gin.Context) {
	//获取当前登录用户ID
	// 注意：这里的 key 必须与 AuthMiddleware 中保持一致（AuthMiddleware 使用的是 "userId"）
	userIdVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, Response{
			Code: 401,
			Msg:  "未获取用户信息，请查询登录",
			Data: nil,
		})
		return
	}
	userId, ok := userIdVal.(int64)
	if !ok {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "用户ID格式错误",
			Data: nil,
		})
		return
	}

	// 获取分页参数
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	// 获取筛选参数
	keyword := c.Query("keyword")
	alertSubTypeStr := c.Query("alert_sub_type")
	statusStr := c.Query("status")

	// 默认 -1 表示不过滤
	alertSubType := -1
	if alertSubTypeStr != "" {
		if v, err := strconv.Atoi(alertSubTypeStr); err == nil {
			alertSubType = v
		}
	}

	status := -1
	if statusStr != "" {
		if v, err := strconv.Atoi(statusStr); err == nil {
			status = v
		}
	}

	// 告警列表兼容新体系：直接从 alert_send_task JOIN monitor 查询，并映射为前端期望字段
	rows, total, err := dao.GetAlertListRowsByUserId(userId, page, size, keyword, alertSubType, status)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "查询告警记录失败" + err.Error(),
			Data: nil,
		})
		return
	}

	// 构建带有监控项名称和展示时间的视图对象
	type AlertView struct {
		Id           int64     `json:"id"`
		MonitorId    int64     `json:"monitorId"`
		MonitorName  string    `json:"monitorName"`
		AlertType    int       `json:"alertType"`
		AlertSubType int       `json:"alertSubType"`
		Status       int       `json:"status"`
		Content      string    `json:"content"`
		SendTime     time.Time `json:"sendTime"`
		CreatedAt    string    `json:"createdAt"`
	}

	alertViews := make([]*AlertView, 0, len(rows))
	for _, r := range rows {
		createdStr := ""
		if !r.CreateTime.IsZero() {
			createdStr = r.CreateTime.Format("2006-01-02 15:04:05")
		}
		alertViews = append(alertViews, &AlertView{
			Id:           r.Id,
			MonitorId:    r.MonitorId,
			MonitorName:  r.MonitorName,
			AlertType:    r.AlertType,
			AlertSubType: r.AlertSubType,
			Status:       r.Status,
			Content:      r.Content,
			SendTime:     r.SendTime,
			CreatedAt:    createdStr,
		})
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "查询告警记录成功",
		Data: gin.H{
			// 前端期望字段名为 list/total/page/size
			"list":  alertViews,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}
