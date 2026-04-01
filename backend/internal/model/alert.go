package model

import "time"

// Alert 告警记录模型
type Alert struct {
	Id           int64     `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	MonitorId    int64     `json:"monitor_id" gorm:"not null"`
	UserId       int64     `json:"user_id" gorm:"not null"`
	AlertType    int       `json:"alert_type" gorm:"default:1;"`     //告警类型（1-邮箱，2-钉钉）
	AlertSubType int       `json:"alert_sub_type" gorm:"default:1;"` // 告警子类型：1-宕机告警 2-恢复通知
	Status       int       `json:"status" gorm:"default:0;"`         // 告警状态（0-未发送，1-已发送，2-发送失败）
	Content      string    `json:"content" gorm:"size:500;not null"`
	SendTime     time.Time `json:"send_time" gorm:"default:null;"`
	CreateTime   time.Time `json:"create_time" gorm:"autoCreateTime;"`
	UpdateTime   time.Time `json:"update_time" gorm:"autoUpdateTime;"`
}

func (Alert) TableName() string {
	return "alert"
}

// AlertConfig 告警配置模型（用户设置的告警方式，对应alert_config表）
type AlertConfig struct {
	Id        int64  `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	UserId    int64  `json:"user_id" gorm:"not null;unique"`
	Email     string `json:"email" gorm:"size:100;not null"`
	AlertType int    `json:"alert_type" gorm:"default:1;"`
	IsEnabled int    `json:"is_enabled" gorm:"default:1;"`
	// 钉钉机器人参数（按用户配置，避免后端硬编码）
	DingTalkWebhook string    `json:"dingtalk_webhook" gorm:"type:varchar(500);default:null"`
	DingTalkSecret  string    `json:"dingtalk_secret" gorm:"type:varchar(200);default:null"`
	DingTalkKeyword string    `json:"dingtalk_keyword" gorm:"type:varchar(50);default:null"`
	UpdateTime      time.Time `json:"update_time" gorm:"autoUpdateTime;"`
}

func (AlertConfig) TableName() string {
	return "alert_config"
}
