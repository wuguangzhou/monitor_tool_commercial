package model

import "time"

// Monitor 监控项模型
type Monitor struct {
	Id          int64     `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	UserId      int64     `json:"userId" gorm:"not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Url         string    `json:"url" gorm:"size:255;not null"`
	MonitorType int       `json:"monitorType" gorm:"default:1"`
	Frequency   int       `json:"frequency" gorm:"default:60"`
	Status      int       `json:"status" gorm:"default:1"`
	Remark      string    `json:"remark" gorm:"size:500"`
	CreateAt    time.Time `json:"createAt" gorm:"autoCreateTime"`
	UpdateAt    time.Time `json:"updateAt" gorm:"autoUpdateTime"`
	ErrorMsg    string    `json:"errorMsg" gorm:"size:500"`
	LastStatus  int       `json:"lastStatus" gorm:"default:0"` //记录上一次状态（判断是否从宕机恢复）
}

func (Monitor) TableName() string {
	return "monitor"
}

type MonitorHistory struct {
	Id           int64     `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	MonitorId    int64     `json:"monitorId" gorm:"not null"`
	Status       int       `json:"status" gorm:"not null"`
	ResponseTime int       `json:"responseTime" gorm:"not null"`
	ErrorMsg     string    `json:"errorMsg" gorm:"size:500"`
	MonitorTime  time.Time `json:"monitorTime" gorm:"autoCreateTime"`
}

func (MonitorHistory) TableName() string {
	return "monitor_history"
}
