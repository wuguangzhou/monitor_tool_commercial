package model

import "time"

type IncidentState string

const (
	IncidentDownActive     IncidentState = "down_active"
	IncidentRecoverPending IncidentState = "recover_pending"
	IncidentClosed         IncidentState = "closed"
)

type Incident struct {
	Id              int64         `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	UserId          int64         `json:"user_id" gorm:"not null;index:idx_user_monitor,priority:1"`
	MonitorId       int64         `json:"monitor_id" gorm:"not null;index:idx_user_monitor,priority:2"`
	IncidentSeq     int           `json:"incident_seq" gorm:"not null"`
	State           IncidentState `json:"state" gorm:"type:varchar(32);not null"`
	DownFirstSeenAt time.Time     `json:"down_first_seen_at"`
	DownLastSeenAt  time.Time     `json:"down_last_seen_at"`
	DownOccurCount  int           `json:"down_occur_count" gorm:"not null;default:0"`
	// 这些字段在宕机阶段通常未知；用 *time.Time 让 GORM 写入 NULL，避免严格模式下的 0000-00-00 错误
	UpFirstSeenAt *time.Time `json:"up_first_seen_at" gorm:"default:null"`
	UpLastSeenAt  *time.Time `json:"up_last_seen_at" gorm:"default:null"`
	UpOccurCount  int        `json:"up_occur_count" gorm:"not null;default:0"`
	LastError     string     `json:"last_error" gorm:"type:varchar(500)"`
	// 关闭时间在恢复最终确认时写入；允许 NULL
	CloseAt  *time.Time `json:"close_at" gorm:"default:null"`
	CreateAt time.Time  `json:"create_at" gorm:"autoCreateTime"`
	UpdateAt time.Time  `json:"update_at" gorm:"autoUpdateTime"`
}

func (Incident) TableName() string {
	return "incident"
}
