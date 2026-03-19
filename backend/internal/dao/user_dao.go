package dao

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

var DB *gorm.DB

// CreateUser 新增用户（注册功能）
func CreateUser(user *model.User) error {
	return DB.Create(user).Error
}

// GetUserByPhone 根据手机号查询用户
func GetUserByPhone(phone string) (*model.User, error) {
	var user model.User
	err := DB.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据ID查询用户
func GetUserById(id int64) (*model.User, error) {
	var user model.User
	err := DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 删除用户
func DeleteUserById(id int64) error {
	var user model.User
	err := DB.Where("id = ?", id).Delete(&user).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(id int64, avatar string) error {
	return DB.Model(&model.User{}).Where("id = ?", id).Update("avatar", avatar).Error
}

// UpdateUserProfile 更新用户基础信息（目前仅用户名）
func UpdateUserProfile(id int64, username string) error {
	return DB.Model(&model.User{}).Where("id = ?", id).Update("username", username).Error
}
