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
	namePatternList                      []string
	alarmConfig                          map[string]string
	period                               time.Duration
	start                                bool
	mutex                                sync.Mutex
	alreadyFinishedNamePatternPidListMap map[string]([]int)
	alarmCountMap                        map[string]int

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

	a.namePatternList = namePatternList
	a.alarmConfig = alarmConfig
	a.alreadyFinishedNamePatternPidListMap = map[string]([]int){}
	a.alarmCountMap = map[string]int{}

	a.processInfoMonitor = NewProcessInfoMonitor(namePatternList)
	a.SetPeriod(defaultPeriod)
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (a *Alarmer) Start() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.start {
		return
	}
	a.start = true

	a.configMonitor.Start()
	a.processInfoMonitor.Start()
	go func() {
		for {
			if !a.start {
				break
			}

			a.alarmIfProcessFinished()
			time.Sleep(a.period)
		}
	}()
}

func (a *Alarmer) alarmIfProcessFinished() {
	for _, namePattern := range a.namePatternList {
		processStatusList := a.findNewlyFinishedProcessesWithNamePattern(namePattern)
		a.alarm(namePattern, processStatusList)
	}
}

func (a *Alarmer) findNewlyFinishedProcessesWithNamePattern(namePattern string) map[int]ProcessStatus {
	newlyFinishedProcessStatusMap := map[int]ProcessStatus{}

	finishedProcessStatusMap := a.processInfoMonitor.GetProcessStatusLogByNamePattern(namePattern)
	alreadyFinishedPidList := a.alreadyFinishedNamePatternPidListMap[namePattern]
	for pid, processStatusHistory := range finishedProcessStatusMap {
		processStatus := processStatusHistory[len(processStatusHistory)-1]
		if processStatus.Status() != ProcessFinished {
			continue
		}
		if !findPidInPidList(pid, alreadyFinishedPidList) {
			newlyFinishedProcessStatusMap[pid] = processStatus
			a.alreadyFinishedNamePatternPidListMap[namePattern] = append(a.alreadyFinishedNamePatternPidListMap[namePattern], pid)
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
	a.mutex.Lock()
	defer a.mutex.Unlock()
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
	a.start = false
}

func (a *Alarmer) SetPeriod(period time.Duration) {
	a.period = period
	a.processInfoMonitor.SetPeriod(period)
}

func (a *Alarmer) GetTotalAlarmCountOfNamePattern(namePattern string) int {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	count, ok := a.alarmCountMap[namePattern]
	if ok == false {
		return 0
	}
	return count
}
