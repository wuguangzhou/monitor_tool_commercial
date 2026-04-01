package dao

import (
	"backend/internal/model"
	"time"
)

// CreateAlert 直接创建新的告警记录（不处理唯一索引冲突，强制新增）
func CreateAlert(alert *model.Alert) error {
	// 设置创建/更新时间（确保记录是最新的）
	alert.CreateTime = time.Now()
	alert.UpdateTime = time.Now()

	// 处理SendTime：未发送时设为NULL
	if alert.SendTime.IsZero() {
		alert.SendTime = time.Time{} // GORM 会自动将空time.Time转为NULL
	}

	// 直接创建新记录，移除OnConflict冲突更新逻辑
	// 注意：如果唯一索引冲突，这里会返回数据库错误（如duplicate key）
	return DB.Create(alert).Error
}

// 创建告警记录
//func CreateAlert(alert *model.Alert) error {
//	// 设置最新的时间（确保记录是最新的）
//	alert.CreateAt = time.Now()
//	alert.UpdateTime = time.Now()
//
//	// 显示处理SendTime(未发送时设为NULL)
//	var sendTime interface{} = nil
//	if !alert.SendTime.IsZero() {
//		sendTime = alert.SendTime
//	}
//
//	//核心逻辑：唯一索引冲突时更新字段，而非新增
//	return DB.Clauses(clause.OnConflict{
//		//指定唯一索引的字段
//		Columns: []clause.Column{
//			{Name: "monitor_id"},
//			{Name: "user_id"},
//			{Name: "alert_type"},
//		},
//		//冲突时要更新的字段
//		DoUpdates: clause.Assignments(map[string]interface{}{
//			"monitor_id":  alert.MonitorId,
//			"user_id":     alert.UserId,
//			"alert_type":  alert.AlertType,
//			"status":      alert.Status,
//			"content":     alert.Content,
//			"send_time":   sendTime,
//			"create_time": alert.CreateAt,
//			"update_time": alert.UpdateTime,
//		}),
//	}).Create(alert).Error
//}

// UpdateAlertStatus 更新告警发送状态
func UpdateAlertStatus(id int64, status int) error {
	return DB.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"send_time": time.Now(),
	}).Error
}

// GetAlertConfigByUserId 获取用户的告警配置
func GetAlertConfigByUserId(userId int64) (*model.AlertConfig, error) {
	var alertConfig model.AlertConfig
	err := DB.Where("user_id = ?", userId).First(&alertConfig).Error
	if err != nil {
		return nil, err
	}
	return &alertConfig, nil
}

// CreateOrUpdateAlertConfig 创建/更新告警配置
func CreateOrUpdateAlertConfig(alertConfig *model.AlertConfig) error {
	//先检查是否存在
	_, err := GetAlertConfigByUserId(alertConfig.UserId)
	if err != nil {
		//不存在则创建
		return DB.Create(alertConfig).Error
	}
	//存在则更新
	return DB.Where("user_id = ?", alertConfig.UserId).Updates(alertConfig).Error
}

// GetUnsentAlert 获取未发送的告警（用于定时发送）
func GetUnsentAlert() ([]*model.Alert, error) {
	var alerts []*model.Alert
	err := DB.Where("status = 0").Limit(100).Find(&alerts).Error
	return alerts, err
}

// GetAlertListByUserId 分页获取用户的告警记录（支持关键词、告警类型、发送状态筛选）
func GetAlertListByUserId(userId int64, page, size int, keyword string, alertSubType, status int) ([]*model.Alert, int64, error) {
	var alerts []*model.Alert
	var total int64

	// 基础查询：当前用户
	db := DB.Model(&model.Alert{}).Where("user_id = ?", userId)

	// 关键词：按告警内容模糊匹配（内容里包含监控项名称和URL）
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("content LIKE ?", like)
	}

	// 告警子类型：1-宕机告警 2-恢复通知；-1 表示不限
	if alertSubType > 0 {
		db = db.Where("alert_sub_type = ?", alertSubType)
	}

	// 发送状态：0-未发送 1-已发送 2-发送失败；-1 表示不限
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	// 先统计总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 再查询当前页数据，按创建时间倒序
	err = db.
		Order("create_time desc").
		Limit(size).
		Offset((page - 1) * size).
		Find(&alerts).Error
	return alerts, total, err
}
