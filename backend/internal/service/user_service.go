package service

import (
	"backend/internal/dao"
	"backend/internal/model"
	"backend/pkg/encrypt"
	"backend/pkg/jwt"
	"backend/pkg/redis"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// UserRegister 用户注册业务逻辑
func UserRegister(phone, username, password string) error {
	// 校验手机号是否注册
	_, err := dao.GetUserByPhone(phone)
	if err == nil {
		return errors.New("该手机号已注册，请更换手机号")
	}
	encryptPwd, err := encrypt.BcryptEncrypt(password)
	if err != nil {
		return fmt.Errorf("密码加密失败，请重试")
	}
	user := &model.User{
		Phone:    phone,
		Username: username,
		Password: encryptPwd,
	}
	return dao.CreateUser(user)
}

// UserLogin 用户登录逻辑(返回token、用户信息)
func UserLogin(phone, password string) (token string, user *model.User, err error) {
	// 手机号校验
	user, err = dao.GetUserByPhone(phone)
	if err != nil {
		return "", nil, errors.New("手机号不存在")
	}

	// 密码校验
	if !encrypt.BcryptVerify(password, user.Password) {
		return "", nil, errors.New("密码错误")
	}

	// 生成JWT Token(登录凭证)
	token, err = jwt.GenerateToken(user.Id, user.Phone)
	if err != nil {
		return "", nil, errors.New("登录失败，请重试")
	}

	// 3. 缓存token到Redis（键：user:token:{userId}，值：token，过期时间7天）
	cacheKey := "user:token:" + strconv.FormatInt(user.Id, 10)
	err = redis.Set(cacheKey, token, 7*24*time.Hour)
	if err != nil {
		return "", nil, err
	}

	//返回token和用户信息
	user.Password = ""
	return token, user, nil
}

// GetUserInfo 根据用户id获取用户信息
func GetUserInfo(userId int64) (*model.User, error) {
	user, err := dao.GetUserById(userId)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

func UserDelete(userId int64) error {
	err := dao.DeleteUserById(userId)
	if err != nil {
		return fmt.Errorf("删除用户%s失败", userId)
	}
	return nil
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(userId int64, avatar string) error {
	if userId == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	return dao.UpdateUserAvatar(userId, avatar)
}

// UpdateUserProfile 更新用户基础信息（目前仅用户名）
func UpdateUserProfile(userId int64, username string) error {
	if userId == 0 {
		return fmt.Errorf("用户ID不能为空")
	}
	if len(username) < 3 || len(username) > 20 {
		return fmt.Errorf("用户名长度需在3-20个字符之间")
	}
	return dao.UpdateUserProfile(userId, username)
}
