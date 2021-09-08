package alarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Alarmer struct {
	monitoringCommandList []string
	alarmConfig           map[string]string
	monitoringPeriod      time.Duration

	isStarted bool

	mutexForSynchronousMethodCall sync.Mutex

	alreadyFinishedMonitoringCommandPidListMap map[string]([]int)

	alarmCountMap         map[string]int
	mutexForAlarmCountMap sync.Mutex

	configMonitor      *ConfigMonitor
	processInfoMonitor *ProcessInfoMonitor
}

func NewAlarmer(configPath string) *Alarmer {
	alarmer := &Alarmer{}
	alarmer.Init(configPath)
	alarmer.Start()
	time.Sleep(defaultPeriod)
	return alarmer
}

func (a *Alarmer) Init(configPath string) {
	a.configMonitor = NewConfigMonitor(configPath)
	namePatternList := a.configMonitor.GetNamePatternList()
	alarmConfig := a.configMonitor.GetAlarmConfig()

	a.monitoringCommandList = namePatternList
	a.alarmConfig = alarmConfig
	a.alreadyFinishedMonitoringCommandPidListMap = map[string]([]int){}
	a.alarmCountMap = map[string]int{}

	a.processInfoMonitor = NewProcessInfoMonitor(namePatternList)
	a.SetPeriod(defaultPeriod)
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (a *Alarmer) Start() {
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
			time.Sleep(a.monitoringPeriod)
		}
	}()
}

func (a *Alarmer) alarmIfProcessFinished() {
	for _, namePattern := range a.monitoringCommandList {
		processStatusList := a.findNewlyFinishedProcessesWithNamePattern(namePattern)
		a.alarm(namePattern, processStatusList)
	}
}

func (a *Alarmer) findNewlyFinishedProcessesWithNamePattern(namePattern string) map[int]ProcessStatus {
	newlyFinishedProcessStatusMap := map[int]ProcessStatus{}

	finishedProcessStatusMap := a.processInfoMonitor.GetProcessStatusLogByNamePattern(namePattern)
	alreadyFinishedPidList := a.alreadyFinishedMonitoringCommandPidListMap[namePattern]
	for pid, processStatusHistory := range finishedProcessStatusMap {
		processStatus := processStatusHistory[len(processStatusHistory)-1]
		if processStatus.Status() != ProcessFinished {
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

func (a *Alarmer) alarm(namePattern string, processStatusHistory map[int]ProcessStatus) {
	a.mutexForSynchronousMethodCall.Lock()
	defer a.mutexForSynchronousMethodCall.Unlock()
	a.alarmCountMap[namePattern] += len(processStatusHistory)
	for pid, processStatus := range processStatusHistory {
		fmt.Println("alarmed pid", pid)
		if a.alarmConfig["type"] == "slack-webhook" {
			webHookUrl := a.alarmConfig["webHookUrl"]
			msg := fmt.Sprintf("NamePattern=%s | PID=%d | STATUS=%s", namePattern, pid, processStatus.Status())
			data := map[string]string{
				"text": msg,
			}
			rawData, _ := json.Marshal(data)
			buff := bytes.NewBuffer(rawData)
			_, err := http.Post(webHookUrl, "application/json", buff)
			if err != nil {
				log.Printf("web hook request is failed: %v\n", err)
			}
		} else {
			panic("NotImplementedError")
		}
	}
}

func (a *Alarmer) Stop() {
	a.isStarted = false
}

func (a *Alarmer) SetPeriod(period time.Duration) {
	a.monitoringPeriod = period
	a.processInfoMonitor.SetPeriod(period)
}

func (a *Alarmer) GetTotalAlarmCountOfNamePattern(namePattern string) int {
	a.mutexForAlarmCountMap.Lock()
	defer a.mutexForAlarmCountMap.Unlock()
	count, ok := a.alarmCountMap[namePattern]
	if ok == false {
		return 0
	}
	return count
}
