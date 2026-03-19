package model

import "time"

type User struct {
	Id          int64     `json:"id" gorm:"primary_key;auto_increment"`
	Username    string    `json:"username" gorm:"size:50;not null"`
	Phone       string    `json:"phone" gorm:"size:20;unique;not null"`
	Password    string    `json:"password" gorm:"size:100;not null"`
	Avatar      string    `json:"avatar" gorm:"size:255"` // 头像地址
	MemberLevel int64     `json:"member_level" gorm:"default:1"`
	MemberEndAt time.Time `json:"member_end_at" gorm:"default:null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// 表名指定
func (User) TableName() string {
	return "user"
}
