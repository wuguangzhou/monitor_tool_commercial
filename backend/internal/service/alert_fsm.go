package service

import (
	"backend/internal/model"
	"backend/pkg/redis"
	"fmt"
	"math"
	"time"
)

type AlertFsmEventType string

const (
	EventDownConfirmed     AlertFsmEventType = "down_confirmed"      // 需要创建 down incident + 生成 down send_task
	EventDownObserved      AlertFsmEventType = "down_observed"       // 需要更新 incident 聚合，但不生成新 send_task
	EventUpPendingObserved AlertFsmEventType = "up_pending_observed" // 仍在抖动压制窗口内，不发恢复通知，但可更新 incident 聚合
	EventUpFinalConfirmed  AlertFsmEventType = "up_final_confirmed"  // 抖动压制窗口结束，恢复通知发出（生成 up send_task + 关闭 incident）
)

type AlertFsmEvent struct {
	Type        AlertFsmEventType
	UserId      int64
	MonitorId   int64
	IncidentSeq int
	At          time.Time
	ErrMsg      string
}
type AlertFsmState struct {
	State              string `json:"state"` // ok/down_suspect/down_active/up_suspect/recover_pending
	IncidentSeq        int    `json:"incident_seq"`
	DownSuspectCount   int    `json:"down_suspect_count"`
	DownSuspectFirstAt int64  `json:"down_suspect_first_at_unix"` // unix seconds
	UpSuspectCount     int    `json:"up_suspect_count"`
	UpSuspectFirstAt   int64  `json:"up_suspect_first_at_unix"` // unix seconds
	// 恢复进入“压制窗口”的起始时间
	RecoverPendingAt int64 `json:"recover_pending_at_unix"` // unix seconds
}

const (
	stateOK             = "ok"
	stateDownSuspect    = "down_suspect"
	stateDownActive     = "down_active"
	stateUpSuspect      = "up_suspect"
	stateRecoverPending = "recover_pending"
)

func alertFsmKey(userId int64, monitorId int64) string {
	return fmt.Sprintf("alert:fsm:u:%d:m:%d", userId, monitorId)
}

// EvaluateAlertFsm 只做降噪状态机判断+写 Redis，并返回需要触发的事件
func EvaluateAlertFsm(monitor *model.Monitor, isDown bool, errMsg string, now time.Time) ([]AlertFsmEvent, error) {
	key := alertFsmKey(monitor.UserId, monitor.Id)

	var st AlertFsmState
	if e := redis.GetJSON(key, &st); e != nil {
		// 未命中，初始化默认状态
		st = AlertFsmState{
			State:       stateOK,
			IncidentSeq: 0,
		}
	}

	// debounce 参数：按文档默认值
	freqSec := monitor.Frequency
	if freqSec < 0 {
		freqSec = 60
	}

	downConfirmCount := 2
	upConfirmCount := 2

	downConfirSeconds := int64(math.Max(20, float64(2*freqSec)))
	upConfirSeconds := int64(math.Max(20, float64(2*freqSec)))

	// “抖动压制窗口”：默认 max(60s, 5*frequency)
	reopenSuppressSeconds := int64(math.Max(60, float64(5*freqSec)))

	confirmedByCount := func(count, need int) bool {
		return count >= need
	}

	confirmedBySeconds := func(firstAtUnix int64, needSeconds int64) bool {
		if firstAtUnix <= 0 {
			return false
		}
		return (now.Unix() - firstAtUnix) >= needSeconds
	}

	isDownConfirmed := func(count int, firstAtUnix int64) bool {
		return confirmedByCount(count, downConfirmCount) || confirmedBySeconds(firstAtUnix, downConfirSeconds)
	}

	isUpConfirmed := func(count int, firstAtUnix int64) bool {
		return confirmedByCount(count, upConfirmCount) || confirmedBySeconds(firstAtUnix, upConfirSeconds)
	}

	// 事件收集器：FSM只负责“产出事件”
	events := make([]AlertFsmEvent, 0, 2)
	emit := func(t AlertFsmEventType, incidentSeq int) {
		events = append(events, AlertFsmEvent{
			Type:        t,
			UserId:      monitor.UserId,
			MonitorId:   monitor.Id,
			IncidentSeq: incidentSeq,
			At:          now,
			ErrMsg:      errMsg,
		})
	}

	/*
		5) 核心状态机：
		- isDown == true  => 失败观测（fail_observed）
		- isDown == false => 成功观测（success_observed）
		注意：
		- “宕机生命周期版本号 incident_seq”只在 EventDownConfirmed 时递增
		- recover_pending 是为了抑制“短暂恢复”导致的恢复通知刷屏
	*/
	if isDown {
		switch st.State {
		case stateOK:
			// 从OK状态进入宕机疑似
			st.State = stateDownSuspect
			st.DownSuspectCount = 1
			st.DownSuspectFirstAt = now.Unix()

			//极端情况：如果 downConfirmCount=1 或者时间阈值已满足（一般不会）
			if isDownConfirmed(st.DownSuspectCount, st.DownSuspectFirstAt) {
				st.IncidentSeq++ //开启新宕机生命周期
				st.State = stateDownActive
				emit(EventDownConfirmed, st.IncidentSeq)
			}
		case stateDownSuspect:
			// 宕机疑似累积：计数+1；首次时间如果没写则补写
			st.DownSuspectCount++
			if st.DownSuspectFirstAt <= 0 {
				st.DownSuspectFirstAt = now.Unix()
			}

			if isDownConfirmed(st.DownSuspectCount, st.DownSuspectFirstAt) {
				st.IncidentSeq++
				st.State = stateDownActive
				emit(EventDownConfirmed, st.IncidentSeq)
			}
		case stateDownActive:
			// 宕机生命周期持续：不再发新的通知，只产生“down_observed”
			emit(EventDownObserved, st.IncidentSeq)

		case stateUpSuspect:
			// 恢复疑似中又失败：典型抖动
			// 策略：回到 down_active，但不递增 incident_seq（仍属于同一宕机生命周期）
			st.State = stateDownActive
			st.UpSuspectCount = 0
			st.UpSuspectFirstAt = 0
			st.RecoverPendingAt = 0

			emit(EventDownObserved, st.IncidentSeq)
		case stateRecoverPending:
			// 恢复压制窗口内再次失败：说明恢复不稳定
			// 策略：取消恢复，回到 down_active，不递增 incident_seq
			st.State = stateDownActive
			st.UpSuspectCount = 0
			st.UpSuspectFirstAt = 0
			st.RecoverPendingAt = 0
			emit(EventDownObserved, st.IncidentSeq)

		default:
			st.State = stateOK
		}
	} else {
		switch st.State {
		case stateOK:
		case stateDownSuspect:
			// 宕机疑似但还没确认就成功了：视为短暂异常，直接清空疑似
			st.State = stateOK
			st.DownSuspectCount = 0
			st.DownSuspectFirstAt = 0
		case stateDownActive:
			// 宕机生命周期中出现成功：进入恢复疑似（up debounce）
			st.State = stateUpSuspect
			st.UpSuspectCount = 1
			st.UpSuspectFirstAt = now.Unix()

			emit(EventUpPendingObserved, st.IncidentSeq)

			if isUpConfirmed(st.UpSuspectCount, st.UpSuspectFirstAt) {
				st.State = stateRecoverPending
				st.RecoverPendingAt = now.Unix()
			}
		case stateUpSuspect:
			// 持续成功：累积恢复疑似次数/时长
			st.UpSuspectCount++
			if st.UpSuspectFirstAt <= 0 {
				st.UpSuspectFirstAt = now.Unix()
			}

			emit(EventUpPendingObserved, st.IncidentSeq)

			if isUpConfirmed(st.UpSuspectCount, st.UpSuspectFirstAt) {
				st.State = stateRecoverPending
				st.RecoverPendingAt = now.Unix()
			}
		case stateRecoverPending:
			emit(EventUpPendingObserved, st.IncidentSeq)

			// 一旦 suppress 窗口过去，且期间都成功（没有被 isDown 打断），则最终确认恢复
			if st.RecoverPendingAt > 0 && (now.Unix()-st.RecoverPendingAt) >= reopenSuppressSeconds {
				// 产出最终恢复事件：第4步会创建 up send_task 并关闭 incident
				emit(EventUpFinalConfirmed, st.IncidentSeq)

				st.State = stateOK
				st.DownSuspectCount = 0
				st.DownSuspectFirstAt = 0
				st.UpSuspectCount = 0
				st.UpSuspectFirstAt = 0
				st.RecoverPendingAt = 0

				// st.IncidentSeq 不清零：它是“已发生宕机生命周期次数”的计数器
				// 下一次宕机确认会 ++
			}
		default:
			st.State = stateOK
		}
	}

	_ = redis.Set(key, st, 24*time.Hour)
	return events, nil
}
