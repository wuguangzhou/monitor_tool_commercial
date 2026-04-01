package dao

import (
	"backend/internal/model"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AlertListRow 是告警列表页所需的“兼容旧 alert 表”的行结构（由新体系 alert_send_task 聚合而来）。
// 注意：字段命名与前端展示对齐（monitorId/monitorName/alertType/alertSubType/...）。
type AlertListRow struct {
	Id           int64     `json:"id"`
	MonitorId    int64     `json:"monitorId"`
	MonitorName  string    `json:"monitorName"`
	AlertType    int       `json:"alertType"`
	AlertSubType int       `json:"alertSubType"` // 1=down 2=up
	Status       int       `json:"status"`       // 0/1/2
	Content      string    `json:"content"`
	SendTime     time.Time `json:"sendTime"`
	CreateTime   time.Time `json:"createTime"`
}

// GetAlertListRowsByUserId 从 alert_send_task + monitor 聚合查询告警记录列表（用于 /api/alert/list 兼容新体系）。
// - keyword：匹配任务 payload 或 monitor.name
// - alertSubType：1=宕机(down) 2=恢复(up)；<=0 表示不筛选
// - status：0=未发送(pending/processing) 1=已发送(sent) 2=失败(failed)；<0 表示不筛选
func GetAlertListRowsByUserId(userId int64, page, size int, keyword string, alertSubType, status int) ([]*AlertListRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	db := DB.Table("alert_send_task AS t").
		Joins("LEFT JOIN monitor m ON m.id = t.monitor_id").
		Where("t.user_id = ?", userId)

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("(t.payload LIKE ? OR m.name LIKE ?)", like, like)
	}

	// 1=宕机(down) 2=恢复(up)
	if alertSubType == 1 {
		db = db.Where("t.task_type = ?", string(model.SendTaskDown))
	} else if alertSubType == 2 {
		db = db.Where("t.task_type = ?", string(model.SendTaskUp))
	}

	// 发送状态映射：0=未发送(pending/processing) 1=sent 2=failed
	if status == 0 {
		db = db.Where("t.status IN (?, ?)", string(model.SendTaskPending), string(model.SendTaskProcessing))
	} else if status == 1 {
		db = db.Where("t.status = ?", string(model.SendTaskSent))
	} else if status == 2 {
		db = db.Where("t.status = ?", string(model.SendTaskFailed))
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 选择字段并做映射：task_type/status 在 SQL 中映射为旧前端字段
	selectSQL := `
		t.id AS id,
		t.monitor_id AS monitor_id,
		COALESCE(m.name, '') AS monitor_name,
		t.alert_type AS alert_type,
		CASE t.task_type
			WHEN 'down' THEN 1
			WHEN 'up' THEN 2
			ELSE 1
		END AS alert_sub_type,
		CASE t.status
			WHEN 'sent' THEN 1
			WHEN 'failed' THEN 2
			ELSE 0
		END AS status,
		t.payload AS content,
		t.send_time AS send_time,
		t.created_at AS create_time
	`

	type rowScan struct {
		Id           int64        `gorm:"column:id"`
		MonitorId    int64        `gorm:"column:monitor_id"`
		MonitorName  string       `gorm:"column:monitor_name"`
		AlertType    int          `gorm:"column:alert_type"`
		AlertSubType int          `gorm:"column:alert_sub_type"`
		Status       int          `gorm:"column:status"`
		Content      string       `gorm:"column:content"`
		SendTime     sql.NullTime `gorm:"column:send_time"`
		CreateTime   time.Time    `gorm:"column:create_time"`
	}

	var rows []rowScan
	err := db.Select(selectSQL).
		Order("t.created_at DESC").
		Limit(size).
		Offset((page - 1) * size).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]*AlertListRow, 0, len(rows))
	for _, r := range rows {
		var sendTime time.Time
		if r.SendTime.Valid {
			sendTime = r.SendTime.Time
		}
		out = append(out, &AlertListRow{
			Id:           r.Id,
			MonitorId:    r.MonitorId,
			MonitorName:  r.MonitorName,
			AlertType:    r.AlertType,
			AlertSubType: r.AlertSubType,
			Status:       r.Status,
			Content:      r.Content,
			SendTime:     sendTime,
			CreateTime:   r.CreateTime,
		})
	}
	return out, total, nil
}

func CreateSendTask(userId, monitorId, incidentId int64, alertType int, taskType model.SendTaskType, payload string, now time.Time) (*model.AlertSendTask, error) {
	task := &model.AlertSendTask{
		UserId:     userId,
		MonitorId:  monitorId,
		IncidentId: incidentId,
		AlertType:  alertType,
		TaskType:   taskType,
		Payload:    payload,
		Status:     model.SendTaskPending,
	}
	if err := DB.Create(task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

// HasSendTaskForIncident 判断某个 incident 是否已存在指定类型的发送任务。
// 用于补偿：如果首次 down_confirmed 时写库失败导致没建 task，后续 down_observed 可补建一次。
func HasSendTaskForIncident(incidentId int64, taskType model.SendTaskType) (bool, error) {
	var cnt int64
	err := DB.Model(&model.AlertSendTask{}).
		Where("incident_id = ? AND task_type = ?", incidentId, taskType).
		Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// claimLockTTL：processing 超过多久仍视为“可抢占”（进程崩溃、发送卡死）
const claimLockTTL = 2 * time.Minute

// generateLockToken：每次 claim 生成随机串，便于日志排查“哪次领取”
func generateLockToken() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond()%1e9)
}

/*
ClaimAlertSendTasks
作用：原子领取一批待发送任务，避免并发下同一行被多个 worker 重复发送。
实现要点：
- 在事务里 SELECT ... FOR UPDATE 锁住候选行，再 UPDATE 为 processing
- 支持“过期 processing”重领：locked_at 早于 now-claimLockTTL 的任务可被重新 claim
注意：
- 需要 InnoDB + 事务；FOR UPDATE 在 GORM 里用 clause.Locking
*/
func ClaimAlertSendTasks(limit int) ([]*model.AlertSendTask, error) {
	if limit <= 0 {
		limit = 100
	}
	staleBefore := time.Now().Add(-claimLockTTL)

	var out []*model.AlertSendTask
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1) 查询候选：pending，或 processing 但锁已过期
		var candidates []*model.AlertSendTask
		q := tx.Model(&model.AlertSendTask{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(
				"status = ? OR (status = ? AND (locked_at IS NULL OR locked_at < ?))",
				model.SendTaskPending,
				model.SendTaskProcessing,
				staleBefore,
			).
			Order("id ASC").
			Limit(limit).
			Find(&candidates)
		if q.Error != nil {
			return q.Error
		}
		if len(candidates) == 0 {
			return nil
		}

		token := generateLockToken()
		now := time.Now()

		// 2) 逐行更新为 processing（仍在同一事务内，行已被锁住）
		for _, t := range candidates {
			res := tx.Model(&model.AlertSendTask{}).
				Where("id = ? AND (status = ? OR (status = ? AND (locked_at IS NULL OR locked_at < ?)))",
					t.Id,
					model.SendTaskPending,
					model.SendTaskProcessing,
					staleBefore).
				Updates(map[string]interface{}{
					"status":     model.SendTaskProcessing,
					"lock_token": token,
					"locked_at":  now,
				})
			if res.Error != nil {
				return res.Error
			}
			// RowsAffected==0 说明被并发抢走，跳过
			if res.RowsAffected == 0 {
				continue
			}

			// 3) 把内存里的对象也改成最新状态，供调用方直接发送（避免再查库）
			t.Status = model.SendTaskProcessing
			t.LockToken = token
			t.LockedAt = &now
			out = append(out, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MarkSendTaskSent：发送成功
func MarkSendTaskSent(id int64) error {
	now := time.Now()
	return DB.Model(&model.AlertSendTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     model.SendTaskSent,
		"send_time":  now,
		"last_error": "",
	}).Error
}

// MarkSendTaskFailed：发送失败（可重试：把 status 改回 pending 或单独做重试字段；这里先标记 failed）
func MarkSendTaskFailed(id int64, errMsg string) error {
	return DB.Model(&model.AlertSendTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     model.SendTaskFailed,
		"last_error": errMsg,
	}).Error
}

/*
ResetSendTaskToPendingForRetry（可选）
若你希望“失败后可重试”，可把 failed 改回 pending，由定时任务再次 claim。
*/
func ResetSendTaskToPendingForRetry(id int64) error {
	return DB.Model(&model.AlertSendTask{}).Where("id = ? AND status = ?", id, model.SendTaskFailed).Updates(map[string]interface{}{
		"status":     model.SendTaskPending,
		"lock_token": "",
		"locked_at":  nil,
		"last_error": "",
	}).Error
}

// GetSendTaskById：调试或补偿用
func GetSendTaskById(id int64) (*model.AlertSendTask, error) {
	var t model.AlertSendTask
	if err := DB.Where("id = ?", id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

var ErrNoAlertConfig = errors.New("用户未配置告警方式")
