package service

import (
	"backend/internal/dao"
	"backend/internal/model"
	alert2 "backend/pkg/alert"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var alertCron *cron.Cron
var alertOnce sync.Once

// CreateAlert 触发告警（监控项宕机时调用）
func CreateAlert(monitor *model.Monitor) error {
	// 校验监控项状态（仅宕机时触发告警）
	if monitor.Status != 2 {
		return errors.New("监控项状态正常，无需触发告警")
	}

	// 获取用户告警配置
	config, err := dao.GetAlertConfigByUserId(monitor.UserId)
	if err != nil {
		return errors.New("用户未配置告警方式，跳过告警")
	}
	if config.IsEnabled == 0 {
		return errors.New("用户已关闭告警，跳过告警")
	}

	//构建告警内容
	content := fmt.Sprintf("【监控工具告警】您的监控项「%s」（%s）于%s发生宕机，请及时处理！", monitor.Name, monitor.Url, time.Now().Format("2006-01-02 15:04:05"))

	//创建告警记录
	alert := &model.Alert{
		MonitorId: monitor.Id,
		UserId:    monitor.UserId,
		AlertType: config.AlertType,
		Content:   content,
		Status:    0,
	}
	err = dao.CreateAlert(alert)
	if err != nil {
		return fmt.Errorf("CreateAlert err:%v", err)
	}
	return nil
}

// 触发恢复通知
func CreateRecoveryAlert(monitor *model.Monitor) error {
	//仅当上一次状态时宕机且当前状态是正常时触发
	if monitor.LastStatus != 2 || monitor.Status != 1 {
		return errors.New("监控项非宕机恢复状态，无需触发恢复通知")
	}

	//获取用户告警配置
	config, err := dao.GetAlertConfigByUserId(monitor.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户未配置告警方式，跳过恢复通知")
		}
		return fmt.Errorf("查询用户告警配置失败：%v", err)
	}

	if config.IsEnabled == 0 {
		return errors.New("用户已关闭告警，跳过恢复通知")
	}

	//构建恢复通知内容
	recoveryContent := alert2.BuildRecoveryContent(monitor)

	//创建恢复通知记录
	alertRecord := &model.Alert{
		MonitorId:    monitor.Id,
		UserId:       monitor.UserId,
		AlertType:    config.AlertType,
		AlertSubType: 2, //恢复通知
		Content:      recoveryContent,
		Status:       0, //未发送
	}
	if err := dao.CreateAlert(alertRecord); err != nil {
		return fmt.Errorf("CreateAlert err:%v", err)
	}

	fmt.Printf("监控项[%d]已恢复，已创建恢复通知记录ID：%d\n", monitor.Id, alertRecord.Id)
	return nil
}

// StartAlertCron 启动告警发送定时任务（每1分钟检查未发送的告警）
func StartAlertCron() {
	alertOnce.Do(func() {
		alertCron = cron.New(cron.WithSeconds())
		alertCron.AddFunc("0 */1 * * * *", func() {
			SendUnsentAlert()
		})
		alertCron.Start()
		fmt.Println("告警发送定时任务已启动")
	})
}

// 停止告警定时任务（程序退出时调用）
func StopAlertCron() {
	if alertCron != nil {
		alertCron.Stop()
		fmt.Println("告警定时任务已停止")
	}
}

func SendUnsentAlert() {
	//获取未发送的告警
	alerts, err := dao.GetUnsentAlert()
	if err != nil {
		fmt.Printf("GetUnsentAlert err:%v", err)
		return
	}
	if len(alerts) == 0 {
		return
	}

	//并发发送告警
	var wg sync.WaitGroup
	for _, alert := range alerts {
		wg.Add(1)
		go func(alert *model.Alert) {
			defer wg.Done()
			//获取用户告警配置
			config, err := dao.GetAlertConfigByUserId(alert.UserId)
			if err != nil {
				fmt.Printf("获取用户%d告警配置失败:%v", alert.UserId, err)
				err := dao.UpdateMonitorStatus(alert.Id, 2)
				if err != nil {
					return
				} //标记为失败
				return
			}

			var sendSuccess bool
			subject := "【告警】"
			switch alert.AlertType {
			case 1: //邮箱
				sendSuccess, err = alert2.SendEmailAlert(config.Email, subject, alert.Content)
			case 2: //钉钉
				sendSuccess, err = alert2.SendDingTalkAlert(fmt.Sprint("监控告警"), alert.Content)
			default:
				err = errors.New("不支持的告警方式")
				sendSuccess = false
			}

			// 更新告警状态
			if sendSuccess && err == nil {
				dao.UpdateAlertStatus(alert.Id, 1) //发送成功
			} else {
				dao.UpdateAlertStatus(alert.Id, 2) //发送失败
				fmt.Printf("告警%d发送失败：:%v\n", alert.UserId, err)
			}
		}(alert)
	}
	wg.Wait()
}

// 更新用户告警配置
func UpdateAlertConfig(config *model.AlertConfig) error {
	return dao.CreateOrUpdateAlertConfig(config)
}

// 获取用户告警配置
func GetAlertConfig(userId int64) (*model.AlertConfig, error) {
	return dao.GetAlertConfigByUserId(userId)
}

// 获取用户告警记录列表（分页，支持筛选）
func GetAlertList(userId int64, page, size int, keyword string, alertSubType, status int) ([]*model.Alert, int64, error) {
	if userId == 0 {
		return nil, 0, errors.New("用户ID不能为空")
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	return dao.GetAlertListByUserId(userId, page, size, keyword, alertSubType, status)
}
