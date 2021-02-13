package alarm

import (
	"sync"
	"time"
)

type Alarmer struct {
	targetProcessList             []string
	period                        time.Duration
	start                         bool
	mutexForStart                 sync.Mutex
	finishedProcessNamePidListMap map[string]([]int)
	alarmCountMap                 map[string]int

	processInfoMonitor *ProcessInfoMonitor
}

func NewAlarmer(targetProcessList []string) *Alarmer {
	alarmer := &Alarmer{}
	alarmer.Init(targetProcessList)
	alarmer.SetPeriod(defaultPeriod)
	alarmer.Start()
	time.Sleep(defaultPeriod)
	return alarmer
}

func (a *Alarmer) Init(targetProcessList []string) {
	a.targetProcessList = targetProcessList
	a.finishedProcessNamePidListMap = map[string]([]int){}
	a.alarmCountMap = map[string]int{}

	a.processInfoMonitor = NewProcessInfoMonitor(targetProcessList)
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (a *Alarmer) Start() {
	a.mutexForStart.Lock()
	defer a.mutexForStart.Unlock()

	if a.start == true {
		return
	}
	a.start = true

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
	for _, targetProcess := range a.targetProcessList {
		processStatusHistory := a.findNewlyFinishedProcess(targetProcess)
		a.alarm(targetProcess, processStatusHistory)
	}
}

func (a *Alarmer) findNewlyFinishedProcess(processName string) map[int]ProcessStatus {
	finishedProcessStatusHistory := map[int]ProcessStatus{}

	processStatusHistory := a.processInfoMonitor.GetCurrentProcessStatusHistoryByName(processName)
	finishedPidList := a.finishedProcessNamePidListMap[processName]
	for pid, processStatus := range processStatusHistory {
		if findPidInPidList(pid, finishedPidList) == false {
			finishedProcessStatusHistory[pid] = processStatus
		}
	}
	return finishedProcessStatusHistory
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
	a.mutexForStart.Lock()
	defer a.mutexForStart.Unlock()
	a.alarmCountMap[processName] += len(processStatusHistory)
}

func (a *Alarmer) Stop() {
	a.start = false
}

func (a *Alarmer) SetPeriod(period time.Duration) {
	a.period = period
	a.processInfoMonitor.SetPeriod(period)
}

func (a *Alarmer) GetAlarmCount(processName string) int {
	a.mutexForStart.Lock()
	defer a.mutexForStart.Unlock()
	count, ok := a.alarmCountMap[processName]
	if ok == false {
		return 0
	}
	return count
}
