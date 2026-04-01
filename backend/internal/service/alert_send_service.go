package service

import (
	"backend/internal/dao"
	"backend/internal/model"
	alert2 "backend/pkg/alert"
	"errors"
	"fmt"
	"sync"
)

/*
SendPendingAlertTasks
第 5 步主入口：从 DB claim 一批任务，按用户配置发邮件/钉钉，再更新任务状态。
调用时机建议：
- monitor_service 里在 ApplyAlertFsmEvents 之后调一次（和现在 SendUnsentAlert 类似）
- 或 cron 每分钟扫一次（与旧 alert 并行期间可两者都开，后期只保留本函数）
并发说明：
- 本函数内部可对已 claim 的任务并发发送；claim 已在 DB 层保证同一行不会被双发
*/
func SendPendingAlertTask() {
	tasks, err := dao.ClaimAlertSendTasks(100)
	if err != nil {
		fmt.Printf("[SendPendingAlertTask] claim alert send task err:%v\n]", err)
		return
	}
	if len(tasks) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, task := range tasks {
		t := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			SendOneAlertTask(t)
		}()
	}
	wg.Wait()
}

/*
sendOneAlertTask：单条任务发送
步骤：
1) 查用户告警配置（邮箱、是否启用、渠道）
2) 按 alert_type 调用现有发送函数
3) 成功 -> MarkSendTaskSent；失败 -> MarkSendTaskFailed
*/
func SendOneAlertTask(task *model.AlertSendTask) {
	// 用户配置
	cfg, err := dao.GetAlertConfigByUserId(task.UserId)
	if err != nil {
		fmt.Printf("[SendOneAlertTask] user=%d get alert config err:%v\n]", task.UserId, err)
		_ = dao.MarkSendTaskFailed(task.Id, err.Error())
		return
	}
	if cfg.IsEnabled == 0 {
		_ = dao.MarkSendTaskFailed(task.Id, "用户已关闭告警")
		return
	}

	subject := buildSubject(task)

	var ok bool
	var sendErr error

	switch task.AlertType {
	case 1:
		ok, sendErr = alert2.SendEmailAlert(cfg.Email, subject, task.Payload)
	case 2:
		ok, sendErr = alert2.SendDingTalkAlertWithConfig(cfg.DingTalkWebhook, cfg.DingTalkSecret, cfg.DingTalkKeyword, subject, task.Payload)
	default:
		sendErr = errors.New("不支持的告警方式")
		ok = false
	}
	if ok && sendErr == nil {
		if err := dao.MarkSendTaskSent(task.Id); err != nil {
			fmt.Printf("[SendOneAlertTask] send task err:%v\n]", err)
		}
		return
	}

	msg := ""
	if sendErr != nil {
		msg = sendErr.Error()
	} else {
		msg = "发送返回失败"
	}
	if err := dao.MarkSendTaskFailed(task.Id, msg); err != nil {
		fmt.Printf("[SendOneAlertTask] mark failed err id=%d: %v\n]", task.Id, err)
	}
}

func buildSubject(task *model.AlertSendTask) string {
	switch task.TaskType {
	case model.SendTaskDown:
		return "【监控告警】宕机"
	case model.SendTaskUp:
		return "【监控恢复】恢复"
	default:
		return "【监控通知】"
	}
}
