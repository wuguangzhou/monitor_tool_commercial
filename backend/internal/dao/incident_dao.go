package dao

import (
	"backend/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// GetOpenIncidentBySeq：按 user+monitor+incident_seq 找“当前生命周期”的 incident（不要求 state）
func GetOpenIncidentBySeq(userId, monitorId int64, incidentSeq int) (*model.Incident, error) {
	var inc model.Incident
	err := DB.Where("user_id=? AND monitor_id=? AND incident_seq=?", userId, monitorId, incidentSeq).
		First(&inc).Error
	if err != nil {
		return nil, err
	}
	return &inc, nil
}

// CreateIncidentDown：宕机确认时创建一行（down_active）
func CreateIncidentDown(userId, monitorId int64, incidentSeq int, now time.Time, errMsg string) (*model.Incident, error) {
	inc := &model.Incident{
		UserId:          userId,
		MonitorId:       monitorId,
		IncidentSeq:     incidentSeq,
		State:           model.IncidentDownActive,
		DownFirstSeenAt: now,
		DownLastSeenAt:  now,
		DownOccurCount:  1,
		LastError:       errMsg,
	}
	// 注意：inc 已经是 *model.Incident，不要再取地址（否则变成 **model.Incident，GORM 可能无法识别类型）
	if err := DB.Create(inc).Error; err != nil {
		return nil, err
	}
	return inc, nil
}

// TouchIncidentDown：宕机持续观测时聚合更新（last_seen/count/error）
func TouchIncidentDown(userId, monitorId int64, incidentSeq int, now time.Time, errMsg string) error {
	updates := map[string]interface{}{
		"down_last_seen_at": now,
		"down_occur_count":  gorm.Expr("down_occur_count + 1"),
		"last_error":        errMsg,
		"state":             model.IncidentDownActive,
	}
	res := DB.Model(&model.Incident{}).
		Where("user_id=? AND monitor_id=? AND incident_seq=?", userId, monitorId, incidentSeq).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	// 关键：RowsAffected==0 不会返回 error，但代表 incident 记录不存在（需要触发上层补偿创建）
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// TouchIncidentUpPending：恢复疑似/压制窗口内成功观测时聚合 up 侧信息（可选但建议）
func TouchIncidentUpPending(userId, monitorId int64, incidentSeq int, now time.Time) error {
	updates := map[string]interface{}{
		"up_last_seen_at": now,
		"up_occur_count":  gorm.Expr("up_occur_count + 1"),
		"state":           model.IncidentRecoverPending,
	}
	return DB.Model(&model.Incident{}).
		Where("user_id=? AND monitor_id=? AND incident_seq=?", userId, monitorId, incidentSeq).
		Updates(updates).Error
}

// CloseIncident：最终恢复确认时关闭 incident
func CloseIncident(userId, monitorId int64, incidentSeq int, now time.Time) error {
	updates := map[string]interface{}{
		"state":     model.IncidentClosed,
		"closed_at": &now,
	}
	return DB.Model(&model.Incident{}).
		Where("user_id=? AND monitor_id=? AND incident_seq=?", userId, monitorId, incidentSeq).
		Updates(updates).Error
}

// EnsureIncidentExistsForSeq：用于容错（例如 Redis 状态丢失/DB 先后顺序异常）
func EnsureIncidentExistsForSeq(userId, monitorId int64, incidentSeq int, now time.Time) (*model.Incident, error) {
	inc, err := GetOpenIncidentBySeq(userId, monitorId, incidentSeq)
	if err == nil {
		return inc, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CreateIncidentDown(userId, monitorId, incidentSeq, now, "")
	}
	return nil, err
}
