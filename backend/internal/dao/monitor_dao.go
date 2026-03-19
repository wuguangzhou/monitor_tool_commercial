package dao

import (
	"backend/internal/model"
	"errors"
	"fmt"
)

// CreateMonitor 新增监控项
func CreateMonitor(monitor *model.Monitor) error {
	return DB.Create(monitor).Error
}

// DeleteMonitor 删除监控项（根据id和用户id,避免删除他人的监控项）
func DeleteMonitor(id, userid int64) error {
	return DB.Where("id = ? and user_id = ?", id, userid).Delete(&model.Monitor{}).Error
}

// UpdateMonitor 更新监控项（名称、频率、备注）
func UpdateMonitor(monitor *model.Monitor) error {
	return DB.Where("id = ? and user_id = ?", monitor.Id, monitor.UserId).Updates(monitor).Error
}

// GetMonitorById 根据ID和用户ID查询监控项
func GetMonitorById(id, userId int64) (*model.Monitor, error) {
	var monitor model.Monitor
	err := DB.Where("id = ? and user_id = ?", id, userId).First(&monitor).Error
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}

// GetMonitorListByUserId 根据用户ID查询监控项列表
func GetMonitorListByUserId(userId int64, page, size int) ([]*model.Monitor, int64, error) {
	var monitors []*model.Monitor
	var total int64

	// 方案：直接执行原生SQL查总数（完全避开GORM的错误处理）
	countSQL := "SELECT COUNT(*) FROM monitor WHERE user_id = ?"
	err := DB.Raw(countSQL, userId).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 如果总数为0，直接返回空列表
	if total == 0 {
		return monitors, 0, nil
	}

	// 执行原生SQL查列表（分页）
	listSQL := "SELECT * FROM monitor WHERE user_id = ? LIMIT ? OFFSET ?"
	offset := (page - 1) * size
	err = DB.Raw(listSQL, userId, size, offset).Scan(&monitors).Error
	if err != nil {
		return nil, 0, err
	}

	return monitors, total, nil
}

// CreateMonitorHistory 新增监控历史记录
func CreateMonitorHistory(history *model.MonitorHistory) error {
	return DB.Create(history).Error
}

// GetMonitorHistoryByMonitorId 根据监控项ID查询监控历史（分页、按时间倒序）
func GetMonitorHistoryByMonitorId(monitorId int64, page, size int) ([]*model.MonitorHistory, int64, error) {
	var historys []*model.MonitorHistory
	var total int64

	countSQL := "SELECT COUNT(*) FROM monitor_history WHERE monitor_id = ?"
	err := DB.Raw(countSQL, monitorId).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return historys, 0, nil
	}

	listSQL := "SELECT * FROM monitor_history WHERE monitor_id = ? LIMIT ? OFFSET ?"
	offset := (page - 1) * size
	err = DB.Raw(listSQL, monitorId, size, offset).Scan(&historys).Error
	if err != nil {
		return nil, 0, err
	}

	return historys, total, err
}

// UpdateMonitorStatus 更新监控项状态(正常/宕机/暂停)
func UpdateMonitorStatus(id int64, status int) error {
	dbName := DB.Migrator().CurrentDatabase()
	fmt.Printf("[关键排查] 当前连接的数据库：%s | 要更新的ID：%d\n", dbName, id)

	result := DB.Exec("UPDATE monitor SET status = ? WHERE id = ?", status, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到要更新的监控项")
	}
	return nil
}

// GetAllValidMonitors 获取所有有效监控项（排除暂停状态，用于定时监控）
func GetAllValidMonitors() ([]*model.Monitor, error) {
	var monitors []*model.Monitor
	// status: 1-正常，2-宕机，3-暂停；这里排除暂停的监控项
	err := DB.Where("status <> ?", 3).Find(&monitors).Error
	return monitors, err
}

// UpdateMonitorStatusWithLast 更新监控项状态，同时记录上一次状态
func UpdateMonitorStatusWithLast(mintorId int64, newStatus int) error {
	//先查询当前状态（作为last_status）
	var monitor model.Monitor
	if err := DB.Where("id = ?", mintorId).First(&monitor).Error; err != nil {
		return err
	}
	// 更新状态
	return DB.Model(&model.Monitor{}).Where("id = ?", mintorId).Updates(map[string]interface{}{
		"last_status": monitor.Status,
		"status":      newStatus,
	}).Error
}

// GetMonitorWithLastStatus 根据ID查监控项（包含last_status）
func GetMonitorWithLastStatus(mintorId int64) (*model.Monitor, error) {
	var monitor model.Monitor
	err := DB.Where("id = ?", mintorId).First(&monitor).Error
	if err != nil {
		return nil, err
	}
	return &monitor, nil
}
