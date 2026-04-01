package model

import "time"

type SendTaskStatus string

const (
	SendTaskPending    SendTaskStatus = "pending"
	SendTaskProcessing SendTaskStatus = "processing"
	SendTaskSent       SendTaskStatus = "sent"
	SendTaskFailed     SendTaskStatus = "failed"
)

type SendTaskType string

const (
	SendTaskDown SendTaskType = "down"
	SendTaskUp   SendTaskType = "up"
)

type AlertSendTask struct {
	Id         int64          `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	UserId     int64          `json:"user_id" gorm:"not null;index:idx_task_user_status,priority:1"`
	MonitorId  int64          `json:"monitor_id" gorm:"not null"`
	IncidentId int64          `json:"incident_id" gorm:"not null;index"`
	AlertType  int            `json:"alert_type" gorm:"not null"`
	TaskType   SendTaskType   `json:"task_type" gorm:"type:varchar(16);not null"`
	Status     SendTaskStatus `json:"status" gorm:"type:varchar(16);not null;index:idx_task_user_status:priority:2"`
	LockToken  string         `json:"lock_token" gorm:"type:varchar(64);index"`
	LockedAt   *time.Time     `json:"locked_at"`
	SendTime   *time.Time     `json:"send_time"`
	LastError  string         `json:"last_error" gorm:"type:varchar(500)"`
	Payload    string         `json:"payload" gorm:"type:text"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (AlertSendTask) TableName() string {
	return "alert_send_task"
}
