package alarm

import (
	"sync"
	"time"
)

type Alarmer struct {
	namePatternList                      []string
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
	alarmer.SetPeriod(defaultPeriod)
	alarmer.Start()
	time.Sleep(defaultPeriod)
	return alarmer
}

func (a *Alarmer) Init(configPath string) {
	a.configMonitor = NewConfigMonitor(configPath)
	namePatternList := a.configMonitor.GetNamePatternList()

	a.namePatternList = namePatternList
	a.alreadyFinishedNamePatternPidListMap = map[string]([]int){}
	a.alarmCountMap = map[string]int{}

	a.processInfoMonitor = NewProcessInfoMonitor(namePatternList)
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (a *Alarmer) Start() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.start == true {
		return
	}
	a.start = true

	a.configMonitor.Start()
	a.processInfoMonitor.Start()
	go func() {
		for {
			if a.start == false {
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

	finishedProcessStatusMap := a.processInfoMonitor.GetCurrentProcessStatusHistoryByName(namePattern)
	alreadyFinishedPidList := a.alreadyFinishedNamePatternPidListMap[namePattern]
	for pid, processStatus := range finishedProcessStatusMap {
		if findPidInPidList(pid, alreadyFinishedPidList) == false {
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

func (a *Alarmer) alarm(processName string, processStatusHistory map[int]ProcessStatus) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.alarmCountMap[processName] += len(processStatusHistory)
}

func (a *Alarmer) Stop() {
	a.start = false
}

func (a *Alarmer) SetPeriod(period time.Duration) {
	a.period = period
	a.processInfoMonitor.SetPeriod(period)
}

func (a *Alarmer) GetTotalAlarmCountOfNamePattern(processName string) int {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	count, ok := a.alarmCountMap[processName]
	if ok == false {
		return 0
	}
	return count
}
