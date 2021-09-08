package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	alarm "github.com/goodahn/AlarmForProgrammer"
)

type SlackWebHookAlarmer struct {
	isStarted bool

	mutexForSynchronousMethodCall sync.Mutex

	alreadyFinishedMonitoringCommandPidListMap map[string]([]int)

	alarmCountMap         map[string]int
	mutexForAlarmCountMap sync.Mutex

	configMonitor      *alarm.ConfigMonitor
	processInfoMonitor *alarm.ProcessInfoMonitor
}

func (a *SlackWebHookAlarmer) Init(configMonitor *alarm.ConfigMonitor) {
	a.configMonitor = configMonitor

	a.alreadyFinishedMonitoringCommandPidListMap = map[string]([]int){}
	a.alarmCountMap = map[string]int{}

	a.processInfoMonitor = alarm.NewProcessInfoMonitor(a.GetMonitoringCommandList())
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (a *SlackWebHookAlarmer) Start() {
	a.mutexForSynchronousMethodCall.Lock()
	defer a.mutexForSynchronousMethodCall.Unlock()

	if a.isStarted {
		return
	}
	a.isStarted = true

	a.configMonitor.Start()
	a.processInfoMonitor.Start()
	go func() {
		for {
			if !a.isStarted {
				break
			}

			a.alarmIfProcessFinished()
			time.Sleep(a.GetMonitoringPeriod())
		}
	}()
}

func (a *SlackWebHookAlarmer) alarmIfProcessFinished() {
	for _, namePattern := range a.GetMonitoringCommandList() {
		processStatusList := a.findNewlyFinishedProcessesWithMonitoringCommand(namePattern)
		a.alarm(namePattern, processStatusList)
	}
}

func (a *SlackWebHookAlarmer) findNewlyFinishedProcessesWithMonitoringCommand(namePattern string) map[int]alarm.ProcessStatus {
	newlyFinishedProcessStatusMap := map[int]alarm.ProcessStatus{}

	finishedProcessStatusMap := a.processInfoMonitor.GetProcessStatusLogByMonitoringCommand(namePattern)
	alreadyFinishedPidList := a.alreadyFinishedMonitoringCommandPidListMap[namePattern]
	for pid, processStatusHistory := range finishedProcessStatusMap {
		processStatus := processStatusHistory[len(processStatusHistory)-1]
		if processStatus.Status() != alarm.ProcessFinished {
			continue
		}
		if !findPidInPidList(pid, alreadyFinishedPidList) {
			newlyFinishedProcessStatusMap[pid] = processStatus
			a.alreadyFinishedMonitoringCommandPidListMap[namePattern] = append(a.alreadyFinishedMonitoringCommandPidListMap[namePattern], pid)
		}
	}
	return newlyFinishedProcessStatusMap
}

func findPidInPidList(pid int, pidList []int) bool {
	for _, _pid := range pidList {
		if pid == _pid {
			return true
		}
	}
	return false
}

func (a *SlackWebHookAlarmer) alarm(namePattern string, processStatusHistory map[int]alarm.ProcessStatus) {
	a.mutexForSynchronousMethodCall.Lock()
	defer a.mutexForSynchronousMethodCall.Unlock()

	a.alarmCountMap[namePattern] += len(processStatusHistory)
	alarmConfig := a.GetAlarmConfig()

	for pid, processStatus := range processStatusHistory {
		webHookUrl := alarmConfig["webHookUrl"]
		msg := fmt.Sprintf("MonitoringCommand=%s | PID=%d | STATUS=%s", namePattern, pid, processStatus.Status())

		data := map[string]string{
			"text": msg,
		}
		rawData, _ := json.Marshal(data)
		buff := bytes.NewBuffer(rawData)
		requestTimeout, err := strconv.Atoi(alarmConfig["requestTimeout"])
		if err != nil {
			panic(err)
		}
		client := http.Client{
			Timeout: time.Duration(requestTimeout),
		}
		_, err = client.Post(webHookUrl, "application/json", buff)
		if err != nil {
			log.Printf("web hook request is failed: %v\n", err)
		}
	}
}

func (a *SlackWebHookAlarmer) Stop() {
	a.isStarted = false
}

func (a *SlackWebHookAlarmer) GetTotalAlarmCountOfMonitoringCommand(namePattern string) int {
	a.mutexForAlarmCountMap.Lock()
	defer a.mutexForAlarmCountMap.Unlock()
	count, ok := a.alarmCountMap[namePattern]
	if ok == false {
		return 0
	}
	return count
}

func (a *SlackWebHookAlarmer) GetAlarmConfig() map[string]string {
	return a.configMonitor.GetAlarmConfig()
}

func (a *SlackWebHookAlarmer) GetAlarmCountMap() map[string]int {
	return a.alarmCountMap
}

func (a *SlackWebHookAlarmer) GetMonitoringCommandList() []string {
	return a.configMonitor.GetMonitoringCommandList()
}

func (a *SlackWebHookAlarmer) GetMonitoringPeriod() time.Duration {
	return time.Duration(a.configMonitor.GetMonitoringPeriod())
}

func (a *SlackWebHookAlarmer) IsStarted() bool {
	return a.isStarted
}
