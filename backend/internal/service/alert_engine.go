package service

import (
	"backend/internal/dao"
	"backend/internal/model"
	"fmt"
	"time"
)

// BuildDownContent / BuildUpContent：Phase A 先直接拼字符串（你也可以复用旧 alert 内容格式）
func BuildDownContent(m *model.Monitor, at time.Time) string {
	return fmt.Sprintf("【监控工具宕机告警】您的监控项「%s」（%s）于%s发生宕机，请及时处理！",
		m.Name, m.Url, at.Format("2006-01-02 15:04:05"))
}

func BuildUpContent(m *model.Monitor, at time.Time) string {
	return fmt.Sprintf("【监控工具恢复通知】\n您的监控项「%s」（%s）于%s已恢复正常状态，请确认！",
		m.Name, m.Url, at.Format("2006-01-02 15:04:05"))
}

// ApplyAlertFsmEvents：第4步核心，把 FSM 事件落库到 incident + send_task
func ApplyAlertFsmEvents(m *model.Monitor, alertType int, events []AlertFsmEvent) error {
	now := time.Now()

	for _, event := range events {
		switch event.Type {
		case EventDownConfirmed:
			// 1) 创建 incident（down_active）
			inc, err := dao.CreateIncidentDown(event.UserId, event.MonitorId, event.IncidentSeq, event.At, event.ErrMsg)
			if err != nil {
				return err
			}

			// 2) 创建 down send_task（pending）
			payload := BuildDownContent(m, event.At)
			_, err = dao.CreateSendTask(event.UserId, event.MonitorId, inc.Id, alertType, model.SendTaskDown, payload, now)
			if err != nil {
				return err
			}
		case EventDownObserved:
			// 宕机持续：默认只更新聚合，不新增 send_task。
			// 但需要补偿一种情况：首次 down_confirmed 时写库失败（如字段为 0000-00-00 导致 incident 插入失败）
			// 之后 FSM 进入 down_active，只会产生 down_observed，导致“永远不会再创建 down 发送任务”。
			// 因此这里做一次幂等补偿：只要 incident 存在且还没有 down task，就补建一次。

			// 1) 先尝试找到 incident（若不存在走后续 ensure）
			inc, getErr := dao.GetOpenIncidentBySeq(event.UserId, event.MonitorId, event.IncidentSeq)
			if getErr == nil && inc != nil {
				hasTask, htErr := dao.HasSendTaskForIncident(inc.Id, model.SendTaskDown)
				if htErr != nil {
					return htErr
				}
				if !hasTask {
					payload := BuildDownContent(m, event.At)
					if _, ctErr := dao.CreateSendTask(event.UserId, event.MonitorId, inc.Id, alertType, model.SendTaskDown, payload, now); ctErr != nil {
						return ctErr
					}
				}
			}

			// 2) 更新 incident 聚合字段；若不存在则补偿创建再更新
			if err := dao.TouchIncidentDown(event.UserId, event.MonitorId, event.IncidentSeq, event.At, event.ErrMsg); err != nil {
				inc2, e2 := dao.EnsureIncidentExistsForSeq(event.UserId, event.MonitorId, event.IncidentSeq, event.At)
				if e2 != nil {
					return e2
				}
				// 若刚创建，也做一次 down task 补偿（幂等）
				hasTask, htErr := dao.HasSendTaskForIncident(inc2.Id, model.SendTaskDown)
				if htErr != nil {
					return htErr
				}
				if !hasTask {
					payload := BuildDownContent(m, event.At)
					if _, ctErr := dao.CreateSendTask(event.UserId, event.MonitorId, inc2.Id, alertType, model.SendTaskDown, payload, now); ctErr != nil {
						return ctErr
					}
				}
				if e3 := dao.TouchIncidentDown(event.UserId, event.MonitorId, event.IncidentSeq, event.At, event.ErrMsg); e3 != nil {
					return e3
				}
			}
		case EventUpPendingObserved:
			// 恢复疑似/压制窗口：更新 incident 的 up 聚合（可选但建议保留）
			_ = dao.TouchIncidentUpPending(event.UserId, event.MonitorId, event.IncidentSeq, event.At)

		case EventUpFinalConfirmed:
			// 1) 取 incident
			inc, err := dao.GetOpenIncidentBySeq(event.UserId, event.MonitorId, event.IncidentSeq)
			if err != nil {
				inc, err = dao.EnsureIncidentExistsForSeq(event.UserId, event.MonitorId, event.IncidentSeq, event.At)
				if err != nil {
					return err
				}
			}
			payload := BuildUpContent(m, event.At)
			if _, err := dao.CreateSendTask(event.UserId, event.MonitorId, inc.Id, alertType, model.SendTaskUp, payload, now); err != nil {
				return err
			}

			//关闭incident
			if err := dao.CloseIncident(event.UserId, event.MonitorId, event.IncidentSeq, event.At); err != nil {
				return err
			}
		}
	}
	return nil
}
