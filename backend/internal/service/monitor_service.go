package service

import (
	"backend/internal/dao"
	"backend/internal/model"
	monitor2 "backend/pkg/monitor"
	"backend/pkg/redis"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// MonitorStatus 监控状态结构体
type MonitorStatus struct {
	MonitorId int64  `json:"monitor_id"`
	Name      string `json:"name"`
	Status    int    `json:"status"` //1-正常，2-宕机，3-暂停
}

// 定时任务实例
var cronInstance *cron.Cron
var once sync.Once // 单例模式，确保定时任务只启动一次

// CreateMonitor 新增监控项
func CreateMonitor(monitor *model.Monitor) error {
	// 校验：用户不能添加过多监控项（普通会员最多5个，后续会员权限完善）
	_, total, err := dao.GetMonitorListByUserId(monitor.UserId, 1, 100)
	if err != nil {
		return errors.New("查询监控项失败，请重试")
	}
	if total >= 5 {
		return errors.New("普通会员最多添加5个监控项，升级会员可解锁更多")
	}

	// 校验URL格式
	if monitor.MonitorType == 1 && (len(monitor.Url) < 7 || (monitor.Url[:7] != "http://" && monitor.Url[:8] != "https://")) {
		return errors.New("HTTP监控URL必须以thhp://或https://开头")
	}

	//新增监控项
	err = dao.CreateMonitor(monitor)
	if err != nil {
		return errors.New("添加监控项失败，请重试")
	}

	// 启动该监控项的定时检测（如果是首次启动，初始化全局定时任务）
	StartMonitorCron()

	// 立即执行一次检测，让用户快速看到结果
	go RunMonitorOnce(monitor.Id, true)

	return nil
}

// DeleteMonitor 删除监控项
func DeleteMonitor(id, userId int64) error {
	// 先校验监控项是否存在且属于当前用户
	_, err := dao.GetMonitorById(id, userId)
	if err != nil {
		return errors.New("监控项不存在或无权限删除")
	}

	return dao.DeleteMonitor(id, userId)
}

// UpdateMonitor 更新监控项
func UpdateMonitor(monitor *model.Monitor) error {
	_, err := dao.GetMonitorById(monitor.Id, monitor.UserId)
	if err != nil {
		return errors.New("监控项不存在或无权限删除")
	}

	// 校验URL格式（如果是HTTP监控）
	if monitor.MonitorType == 1 && (len(monitor.Url) < 7 || (monitor.Url[:7] != "http://" && monitor.Url[:8] != "https://")) {
		return errors.New("HTTP监控URL必须以http://或https://开头")
	}

	return dao.UpdateMonitor(monitor)
}

// GetMonitorList 获取用户的监控项列表
func GetMonitorList(userId int64, page, size int) ([]*model.Monitor, int64, error) {
	monitors, total, err := dao.GetMonitorListByUserId(userId, page, size)
	if err != nil {
		return nil, 0, errors.New("查询监控列表失败，请重试")
	}

	// 补充Redis中的实时状态
	for _, monitor := range monitors {
		NewMonitorstatus, err := GetMonitorStatus(monitor.Id)
		if err == nil {
			monitor.Status = NewMonitorstatus.Status //用Redis缓存的实时状态覆盖数据库状态
		}

	}

	return monitors, total, nil
}

// GetMonitorDetail 获取监控项详情+历史记录
func GetMonitorDetail(id, usesrId int64, page, size int) (*model.Monitor, []*model.MonitorHistory, int64, error) {
	//查询监控项详情
	monitor, err := dao.GetMonitorById(id, usesrId)
	if err != nil {
		return nil, nil, 0, errors.New("监控项不存在在或无权查看")
	}

	NewMonitorstatus, err := GetMonitorStatus(id)
	if err == nil {
		monitor.Status = NewMonitorstatus.Status
	}

	//查询监控历史记录
	histories, total, err := dao.GetMonitorHistoryByMonitorId(id, page, size)
	if err != nil {
		return monitor, histories, total, errors.New("查询监控历史失败，请重试")
	}
	return monitor, histories, total, nil
}

// RunMonitorOnce 手动执行一次监控检测
func RunMonitorOnce(monitorId int64, fromOneHand bool) error {
	//查询监控项信息
	fmt.Printf("开始检测监控项ID：%d\n", monitorId)
	monitor, err := dao.GetMonitorWithLastStatus(monitorId)
	if err != nil {
		fmt.Printf("查询监控项失败：%v\n", err)
		return fmt.Errorf("查询监控项失败：%v", err)
	}
	fmt.Printf("监控项信息：名称=%s，URL=%s，当前状态=%d，上一次状态=%d\n", monitor.Name, monitor.Url, monitor.Status, monitor.LastStatus)

	//执行检测
	var newStatus int
	var responseTime int
	var errMsg string
	switch monitor.MonitorType {
	case 1: //HTTP/HTTPS监控
		newStatus, responseTime, err = monitor2.HTTPMonitor(monitor.Url)
		if err != nil {
			errMsg = err.Error()
		}
	default: // 暂时只支持HTTP，其他默认宕机
		newStatus = 2
		errMsg = "暂不支持该监控项"
	}

	//记录监控历史
	history := &model.MonitorHistory{
		MonitorId:    monitorId,
		Status:       newStatus,
		ResponseTime: responseTime,
		ErrorMsg:     errMsg,
		MonitorTime:  time.Now(),
	}
	err = dao.CreateMonitorHistory(history)
	if err != nil {
		fmt.Printf("插入历史记录失败%v\n", err)
		return fmt.Errorf("插入历史记录失败%v", err)
	}

	//更新监控项状态
	err = dao.UpdateMonitorStatusWithLast(monitorId, newStatus)
	if err != nil {
		fmt.Printf("更新监控项状态失败%v\n", err)
		return fmt.Errorf("更新监控项状态失败%v", err)
	}
	// 修改后
	redisErr := UpdateMonitorStatus(monitor.Id, newStatus)
	if redisErr != nil {
		fmt.Printf("更新Redis监控状态失败：监控项ID=%d，错误=%v\n", monitor.Id, redisErr)
	}
	//刷新内存中的监控项状态
	monitor.LastStatus = monitor.Status
	monitor.Status = newStatus
	monitor.ErrorMsg = errMsg

	// 宕机时触发告警：仅在状态从非宕机(!=2) -> 宕机(2) 的“第一次切换”时触发
	// 避免定时任务在持续宕机期间反复覆盖告警记录，导致列表一直显示最新的“未发送”告警
	if newStatus == 2 {
		fmt.Printf("监控项ID=%d宕机，首次进入宕机状态，触发告警\n", monitorId)
		go func() {
			alertErr := CreateAlert(monitor)
			if alertErr != nil {
				fmt.Printf("触发告警失败%v\n", alertErr)
			}
			//触发告警成功就立即发送告警，而非等待定时任务
			SendUnsentAlert()
			//手动一次即可发送告警
			if fromOneHand {
				SendUnsentAlert()
			}
		}()
	}

	// 恢复通知
	if newStatus == 1 && monitor.LastStatus == 2 {
		fmt.Printf("监控项ID=%d从宕机恢复，触发恢复通知\n", monitorId)
		go func() {
			recoveryErr := CreateRecoveryAlert(monitor)
			if recoveryErr != nil {
				fmt.Printf("触发恢复通知失败%v\n", recoveryErr)
			}
			//触发恢复通知成功就立即发送告警，而非等待定时任务
			SendUnsentAlert()
			//手动一次即可发送告警
			if fromOneHand {
				SendUnsentAlert()
			}
		}()
	}

	fmt.Printf("监控项ID=%d检测完成，新状态=%d,历史记录插入成功\n", monitorId, newStatus)
	return nil
}

// StartMonitorCron 启动全局定时监控任务（单例）
func StartMonitorCron() {
	once.Do(func() {
		cronInstance := cron.New(cron.WithSeconds()) //支持秒级定时
		//添加全局定时任务（每10秒检查所有监控项，按频率执行检测）
		cronInstance.AddFunc("*/10 * * * * *", func() {
			RunAllMonitor()
		})
		cronInstance.Start()
		fmt.Println("定时任务已启动")
	})
}

// RunAllMonitor 执行所有有效监控项的检测
func RunAllMonitor() {
	monitors, err := dao.GetAllValidMonitors()
	if err != nil {
		fmt.Printf("查询有效监控项失败：%v\n", err)
		return
	}
	// 并发执行检测
	var wg sync.WaitGroup
	for _, monitor := range monitors {
		wg.Add(1)
		go func(monitorId int64) {
			defer wg.Done()
			//校验频率
			now := time.Now().Unix()
			if now%int64(monitor.Frequency) != 0 {
				return
			}

			err := RunMonitorOnce(monitorId, false)
			if err != nil {
				return
			}
		}(monitor.Id)
	}
	wg.Wait()
}

// PauseMonitor 暂停监控项
func PauseMonitor(id, userId int64) error {
	// 校验权限
	_, err := dao.GetMonitorById(id, userId)
	if err != nil {
		return errors.New("监控项不存在或无权操作")
	}

	// 统一通过 UpdateMonitorStatus 更新数据库和缓存为暂停（3）
	return UpdateMonitorStatus(id, 3)
}

// ResumeMonitor 恢复监控项
func ResumeMonitor(id, userId int64) error {
	_, err := dao.GetMonitorById(id, userId)
	if err != nil {
		return errors.New("监控项不存在或无权操作")
	}
	// 统一通过 UpdateMonitorStatus 更新数据库和缓存为正常（1）
	if err := UpdateMonitorStatus(id, 1); err != nil {
		return err
	}

	// 立即执行一次检测
	go RunMonitorOnce(id, true)
	return nil
}

// GetMonitorStatus 获取监控项状态（优先从Redis缓存获取）
func GetMonitorStatus(monitorId int64) (*MonitorStatus, error) {
	//定义缓存键
	cacheKey := "monitor:status:" + strconv.FormatInt(monitorId, 10)

	//先查redis缓存
	var status MonitorStatus
	err := redis.GetJSON(cacheKey, &status)
	if err == nil {
		//缓存命中，直接返回
		return &status, nil
	}
	//缓存未命中，查数据库
	monitor, err := dao.GetMonitorWithLastStatus(monitorId)
	if err != nil {
		return nil, err
	}

	//组装状态数据
	status = MonitorStatus{
		MonitorId: monitorId,
		Name:      monitor.Name,
		Status:    monitor.Status,
	}
	// 写入Redis缓存（过期时间5分钟）
	err = redis.Set(cacheKey, status, 5*time.Minute)
	if err != nil {
		log.Printf("缓存监控状态失败: monitorId=%d, err=%v", monitorId, err)
	}
	return &status, nil
}

// UpdateMonitorStatus 更新监控状态（只更新redis缓存,因为数据库在另一个地方更新了）
func UpdateMonitorStatus(monitorId int64, status int) error {
	// 更新redis缓存（先查原数据，再更新；如果没有则重建）
	cacheKey := "monitor:status:" + strconv.FormatInt(monitorId, 10)
	var monitorStatus MonitorStatus
	redisGetErr := redis.GetJSON(cacheKey, &monitorStatus)
	if redisGetErr == nil {
		//已有缓存
		monitorStatus.Status = status
		redisSetErr := redis.Set(cacheKey, monitorStatus, 5*time.Minute)
		if redisGetErr != nil {
			fmt.Printf("更新Redis缓存失败（已有缓存）：key=%s，错误=%v\n", cacheKey, redisSetErr)
		} else {
			fmt.Printf("Redis缓存更新成功（已有缓存）：key=%s，新状态=%d\n", cacheKey, status)
		}
	} else {
		// 3.2 缓存不存在/读取失败：从数据库重建缓存
		fmt.Printf("Redis缓存不存在，从数据库重建：key=%s，错误=%v\n", cacheKey, redisGetErr)
		monitor, dbErr := dao.GetMonitorWithLastStatus(monitorId)
		if dbErr != nil {
			fmt.Printf("从数据库重建缓存失败：monitorId=%d，错误=%v\n", monitorId, dbErr)
			return fmt.Errorf("重建缓存失败：%v", dbErr)
		}
		if monitor == nil {
			fmt.Printf("数据库中未找到监控项：monitorId=%d\n", monitorId)
			return fmt.Errorf("监控项%d不存在", monitorId)
		}
		// 构建缓存结构体
		monitorStatus = MonitorStatus{
			MonitorId: monitorId,
			Name:      monitor.Name,
			Status:    status, // 使用最新的状态，而非数据库中的旧状态
		}

		redisSetErr := redis.Set(cacheKey, monitorStatus, 5*time.Minute)
		if redisSetErr != nil {
			fmt.Printf("写入Redis缓存失败（重建）：key=%s，错误=%v\n", cacheKey, redisSetErr)
		} else {
			fmt.Printf("Redis缓存重建成功：key=%s，状态=%d\n", cacheKey, status)
		}
	}
	return nil
}

// ClearMonitorCache 清除指定监控项的缓存
func ClearMonitorCache(monitorId int64) {
	cacheKey := "monitor:status:" + strconv.FormatInt(monitorId, 10)
	_ = redis.Del(cacheKey)
}
