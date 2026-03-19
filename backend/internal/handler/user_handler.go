package handler

import (
	"backend/internal/service"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// 统一响应结构体（所有接口统一返回格式，便于前端对接）
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 注册接口
func UserRegisterHandler(c *gin.Context) {
	// 接收前端传递的参数
	type RegisterParam struct {
		Phone    string `json:"phone" binding:"required,len=11"`
		Password string `json:"password" binding:"required,min=8,max=20"`
		Username string `json:"username" binding:"required,min=3,max=20"`
	}
	var param RegisterParam

	//校验参数(绑定+校验，不符合规则直接返回错误)
	if err := c.ShouldBind(&param); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "参数错误：手机号必须11位，密码至少8位，最多20位，用户名至少3位，做多20位",
			Data: nil,
		})
		return
	}

	// 调用service层业务逻辑
	err := service.UserRegister(param.Phone, param.Username, param.Password)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}
	// 注册成功
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "注册成功",
		Data: param,
	})
}

// 登录接口
func UserLoginHandler(c *gin.Context) {
	type LoginParam struct {
		Phone    string `json:"phone" binding:"required,len=11"`
		Password string `json:"password" binding:"required,min=8,max=20"`
	}
	var param LoginParam
	if err := c.ShouldBind(&param); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "参数错误：手机号必须11位，密码至少8位，最多20位，用户名至少3位，最多20位",
			Data: nil,
		})
		return
	}

	token, user, err := service.UserLogin(param.Phone, param.Password)
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
		Msg:  "登录成功",
		Data: gin.H{
			"token": token,
			"user":  user,
		},
	})
}

func UserInfoHandler(c *gin.Context) {
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
	user, err := service.GetUserInfo(userId)
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
		Msg:  "success",
		Data: user,
	})
}

func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  message,
		Data: nil,
	})
}

func UserDeleteHandler(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	if !exists || userIdVal == nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code: 401,
			Msg:  "未找到用户ID",
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
	err := service.UserDelete(userId)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "delete success",
	})
}

// UploadAvatarHandler 用户头像上传
func UploadAvatarHandler(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	if !exists || userIdVal == nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code: 401,
			Msg:  "未找到用户ID",
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

	// 读取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "上传文件失败：" + err.Error(),
		})
		return
	}

	// 创建存储目录
	baseDir := "./static/avatar"
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建头像目录失败：" + err.Error(),
		})
		return
	}

	// 生成文件名：userId_时间戳.ext
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", userId, time.Now().Unix(), ext)
	filePath := filepath.Join(baseDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "保存头像失败：" + err.Error(),
		})
		return
	}

	// 头像访问URL（前端通过该地址访问）
	avatarURL := "/static/avatar/" + filename

	// 更新数据库中的头像字段
	if err := service.UpdateUserAvatar(userId, avatarURL); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "更新头像信息失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "头像上传成功",
		Data: gin.H{
			"avatar": avatarURL,
		},
	})
}

// UpdateUserProfileHandler 更新个人基础信息（昵称）
func UpdateUserProfileHandler(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	if !exists || userIdVal == nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code: 401,
			Msg:  "未找到用户ID",
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

	type ProfileParam struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
	}
	var param ProfileParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "用户名长度需在3-20个字符之间",
			Data: nil,
		})
		return
	}

	if err := service.UpdateUserProfile(userId, param.Username); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "个人信息更新成功",
		Data: nil,
	})
}
