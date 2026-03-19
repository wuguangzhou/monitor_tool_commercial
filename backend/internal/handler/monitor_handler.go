package handler

import (
	"backend/internal/dao"
	"backend/internal/model"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 新增监控项接口
func CreateMonitorHandler(c *gin.Context) {
	// 获取当前登录用户id
	userIdVal, exists := c.Get("userId")
	if !exists || userIdVal == nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code: 401,
			Msg:  "未获取用户ID",
		})
		return
	}
	userId, ok := userIdVal.(int64)
	if !ok {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "用户ID类型错误，必须是整数",
		})
		return
	}

	//接收参数
	type MonitorParam struct {
		Name        string `json:"name" binding:"required,min=2"`
		Url         string `json:"url" binding:"required"`
		MonitorType int    `json:"monitorType" binding:"required"`
		Frequency   int    `json:"frequency" binding:"required,min=10"`
		Remark      string `json:"remark"`
	}
	var param MonitorParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "参数错误：名称至少2位，URL必填，监控频率最小10秒",
			Data: nil,
		})
		return
	}

	//构建监控项模型
	monitor := &model.Monitor{
		UserId:      userId,
		Name:        param.Name,
		Url:         param.Url,
		MonitorType: param.MonitorType,
		Frequency:   param.Frequency,
		Remark:      param.Remark,
	}

	//调用service层
	err := service.CreateMonitor(monitor)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "监控项添加成功，已开始检测",
		Data: monitor,
	})
}

// 删除监控项接口
func DeleteMonitorHandler(c *gin.Context) {
	monitorIdStr := c.Param("id")
	monitorId, err := strconv.ParseInt(monitorIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "监控项ID格式错误",
			Data: nil,
		})
		return
	}
	userId, _ := c.Get("userId")

	err = service.DeleteMonitor(monitorId, userId.(int64))
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "监控项删除成功",
		Data: nil,
	})
}

func UpdateMonitorHandler(c *gin.Context) {
	monitorIdStr := c.Param("id")
	monitorId, err := strconv.ParseInt(monitorIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "监控项ID格式错误",
			Data: nil,
		})
		return
	}
	userId, _ := c.Get("userId")

	type MonitorParam struct {
		Name        string `json:"name" binding:"required,min=2"`
		Url         string `json:"url" binding:"required"`
		MonitorType int    `json:"monitorType" binding:"required"`
		Frequency   int    `json:"frequency" binding:"required,min=10"`
		Remark      string `json:"remark"`
	}

	var param MonitorParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "参数错误：名称至少2位，URL必填，监控频率最小10秒",
			Data: nil,
		})
		return
	}

	monitor := &model.Monitor{
		Id:          monitorId,
		UserId:      userId.(int64),
		Name:        param.Name,
		Url:         param.Url,
		MonitorType: param.MonitorType,
		Frequency:   param.Frequency,
		Remark:      param.Remark,
	}

	err = service.UpdateMonitor(monitor)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "监控项更新成功",
		Data: nil,
	})
}

// GetMonitorListHandler 获取监控项列表接口
func GetMonitorListHandler(c *gin.Context) {
	// 获取分页参数
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	// 获取当前登录用户ID
	userId, _ := c.Get("userId")

	// 调用service层
	monitors, total, err := service.GetMonitorList(userId.(int64), page, size)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	// 构建monitor视图
	type MonitorView struct {
		Id          int64  `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
		UserId      int64  `json:"userId" gorm:"not null"`
		Name        string `json:"name" gorm:"size:100;not null"`
		Url         string `json:"url" gorm:"size:255;not null"`
		MonitorType int    `json:"monitorType" gorm:"default:1"`
		Frequency   int    `json:"frequency" gorm:"default:60"`
		Status      int    `json:"status" gorm:"default:1"`
		Remark      string `json:"remark" gorm:"size:500"`
		CreateAt    string `json:"createAt" gorm:"autoCreateTime"`
		UpdateAt    string `json:"updateAt" gorm:"autoUpdateTime"`
		ErrorMsg    string `json:"errorMsg" gorm:"size:500"`
		LastStatus  int    `json:"lastStatus" gorm:"default:0"`
	}
	monitorViews := make([]*MonitorView, 0, len(monitors))
	for _, monitor := range monitors {
		createdStr := ""
		if !monitor.CreateAt.IsZero() {
			createdStr = monitor.CreateAt.Format("2006-01-02 15:04:05")
		}

		monitorViews = append(monitorViews, &MonitorView{
			Id:          monitor.Id,
			UserId:      monitor.UserId,
			Name:        monitor.Name,
			Url:         monitor.Url,
			MonitorType: monitor.MonitorType,
			Frequency:   monitor.Frequency,
			Remark:      monitor.Remark,
			Status:      monitor.Status,
			CreateAt:    createdStr,
			UpdateAt:    monitor.UpdateAt.Format("2006-01-02 15:04:05"),
			ErrorMsg:    monitor.ErrorMsg,
			LastStatus:  monitor.LastStatus,
		})
	}

	// 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "查询成功",
		Data: gin.H{
			"list":  monitorViews,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// GetMonitorDetailHandler 获取监控项详情接口
func GetMonitorDetailHandler(c *gin.Context) {
	// 获取参数
	monitorIdStr := c.Param("id")
	monitorId, err := strconv.ParseInt(monitorIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "监控项ID格式错误",
			Data: nil,
		})
		return
	}

	// 获取分页参数（历史记录）
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	// 获取当前登录用户ID
	userId, _ := c.Get("userId")

	// 调用service层
	monitor, histories, total, err := service.GetMonitorDetail(monitorId, userId.(int64), page, size)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "查询成功",
		Data: gin.H{
			"detail":    monitor,
			"histories": histories,
			"total":     total,
			"page":      page,
			"size":      size,
		},
	})
}

// RunMonitorOnceHandler 手动执行一次监控接口
func RunMonitorOnceHandler(c *gin.Context) {
	// 获取参数
	monitorIdStr := c.Param("id")
	monitorId, err := strconv.ParseInt(monitorIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "监控项ID格式错误",
			Data: nil,
		})
		return
	}
	// 2. 安全获取 userId（重点修复部分）
	// 第一步：从 Context 取值，先检查是否存在
	userIdVal, exists := c.Get("userId") // 确保 key 和中间件的一致！
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录，无法执行监控"})
		return
	}
	// 第二步：检查是否为 nil
	if userIdVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户ID为空，身份验证失败"})
		return
	}
	// 第三步：安全断言为 int64（带 ok 检查）
	userId, ok := userIdVal.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "用户ID格式错误，应为整数"})
		return
	}
	_, err = dao.GetMonitorById(monitorId, userId)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "监控项不存在或无权操作",
			Data: nil,
		})
		return
	}

	// 异步执行检测（避免接口阻塞）
	go func() {
		err := service.RunMonitorOnce(monitorId, true)
		if err != nil {
			return
		}
	}()
	// 同步执行（方便看报错）
	//err = service.RunMonitorOnce(monitorId)
	//if err != nil {
	//	c.JSON(http.StatusOK, Response{
	//		Code: 400,
	//		Msg:  "检测失败：" + err.Error(),
	//		Data: nil,
	//	})
	//	return
	//}
	// 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "已开始执行手动检测，结果将实时更新",
		Data: nil,
	})
}

// PauseMonitorHandler 暂停监控项接口
func PauseMonitorHandler(c *gin.Context) {
	// 获取参数
	monitorIdStr := c.Param("id")
	monitorId, err := strconv.ParseInt(monitorIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "监控项ID格式错误",
			Data: nil,
		})
		return
	}

	// 安全获取当前登录用户ID
	userIdVal, exists := c.Get("userId")
	if !exists || userIdVal == nil {
		c.JSON(http.StatusOK, Response{
			Code: 401,
			Msg:  "未获取用户ID",
			Data: nil,
		})
		return
	}
	userId, ok := userIdVal.(int64)
	if !ok {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "用户ID类型错误，必须是整数",
			Data: nil,
		})
		return
	}

	// 调用service层
	err = service.PauseMonitor(monitorId, userId)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "监控项已暂停",
		Data: nil,
	})
}

// ResumeMonitorHandler 恢复监控项接口
func ResumeMonitorHandler(c *gin.Context) {
	// 获取参数
	monitorIdStr := c.Param("id")
	monitorId, err := strconv.ParseInt(monitorIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "监控项ID格式错误",
			Data: nil,
		})
		return
	}

	// 安全获取当前登录用户ID
	userIdVal, exists := c.Get("userId")
	if !exists || userIdVal == nil {
		c.JSON(http.StatusOK, Response{
			Code: 401,
			Msg:  "未获取用户ID",
			Data: nil,
		})
		return
	}
	userId, ok := userIdVal.(int64)
	if !ok {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "用户ID类型错误，必须是整数",
			Data: nil,
		})
		return
	}

	// 调用service层
	err = service.ResumeMonitor(monitorId, userId)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "监控项已恢复，开始实时检测",
		Data: nil,
	})
}
