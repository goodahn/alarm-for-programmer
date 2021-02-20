package alarm

import (
	"sync"
	"time"
)

type ProcessInfoMonitor struct {
	namePatternList               []string
	processStatusLogByNamePattern map[string](map[int]ProcessStatus)
	period                        time.Duration
	start                         bool
	mutex                         sync.Mutex

	now time.Time

	processInfoReader *ProcessInfoReader
}

func NewProcessInfoMonitor(namePatternList []string) *ProcessInfoMonitor {
	pim := &ProcessInfoMonitor{}
	pim.Init()
	pim.SetNamePatternList(namePatternList)
	pim.SetPeriod(defaultPeriod)
	pim.Start()
	time.Sleep(defaultPeriod)
	return pim
}

func (pim *ProcessInfoMonitor) Init() {
	pim.processStatusLogByNamePattern = map[string](map[int]ProcessStatus){}
	pim.processInfoReader = NewProcessInfoReader()
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (pim *ProcessInfoMonitor) Start() {
	pim.mutex.Lock()
	defer pim.mutex.Unlock()

	if pim.start == true {
		return
	}
	pim.start = true

	pim.processInfoReader.Start()
	go func() {
		for {
			if pim.start == false {
				break
			}

			pim.now = time.Now()
			pim.updateTargetProcessStatusHistory()
		}
	}()
}

func (pim *ProcessInfoMonitor) updateTargetProcessStatusHistory() {
	for _, namePattern := range pim.namePatternList {
		pim.updateProcessStatusHistory(namePattern)
	}
}

func (pim *ProcessInfoMonitor) updateProcessStatusHistory(processName string) {
	changedProcessStatusHistory := pim.checkProcessStatusChanged(processName)
	if len(changedProcessStatusHistory) == 0 {
		return
	}
	pim.appendProcessStatusHistory(processName, changedProcessStatusHistory)
}

func (pim *ProcessInfoMonitor) checkProcessStatusChanged(processName string) map[int]ProcessStatus {
	changedProcessStatusHistory := map[int]ProcessStatus{}

	pim.mutex.Lock()
	defer pim.mutex.Unlock()
	processStatusHistory := pim.processStatusLogByNamePattern[processName]

	pidList := pim.processInfoReader.GetPidListByName(processName)
	for _, pid := range pidList {
		processStatus := pim.getUpdatedProcessStatusInHistoryWithTimestamp(processStatusHistory, pid)
		if (processStatus == ProcessStatus{}) {
			continue
		}
		changedProcessStatusHistory[pid] = processStatus
	}
	return changedProcessStatusHistory
}

func (pim *ProcessInfoMonitor) getUpdatedProcessStatusInHistoryWithTimestamp(processStatusHistory map[int]ProcessStatus, pid int) ProcessStatus {
	isExecuting := pim.processInfoReader.IsExecuting(pid)
	if isExecuting {
		// process is started after its previous execution
		mostRecentProcessStatus := findProcessStatusInHistory(processStatusHistory, pid)
		if mostRecentProcessStatus.Status() == ProcessFinished ||
			mostRecentProcessStatus.Status() == ProcessNeverStarted {
			return NewProcessStatus(pid, ProcessStarted)
		}
	} else {
		// process is finshed after start
		mostRecentProcessStatus := findProcessStatusInHistory(processStatusHistory, pid)
		if mostRecentProcessStatus.Status() == ProcessStarted {
			return NewProcessStatus(pid, ProcessFinished)
		}
	}
	return ProcessStatus{}
}

func (pim *ProcessInfoMonitor) appendProcessStatusHistory(processName string, processStatusHistory map[int]ProcessStatus) {
	pim.mutex.Lock()
	defer pim.mutex.Unlock()
	pim.processStatusLogByNamePattern[processName] = processStatusHistory
}

func (pim *ProcessInfoMonitor) Stop() {
	pim.start = false
	pim.processInfoReader.Stop()
}

func (pim *ProcessInfoMonitor) SetPeriod(period time.Duration) {
	pim.period = period
	pim.processInfoReader.SetPeriod(period)
}

func (pim *ProcessInfoMonitor) SetNamePatternList(namePatternList []string) {
	pim.namePatternList = namePatternList
	for _, processName := range namePatternList {
		pim.processStatusLogByNamePattern[processName] = map[int]ProcessStatus{}
	}
}

func (pim *ProcessInfoMonitor) GetCurrentProcessStatusLogByNamePattern(namePattern string) map[int]ProcessStatus {
	pim.mutex.Lock()
	defer pim.mutex.Unlock()
	return pim.getProcessStatusLog(namePattern)
}

func (pim *ProcessInfoMonitor) getProcessStatusLog(processName string) map[int]ProcessStatus {
	processStatusHistory, ok := pim.processStatusLogByNamePattern[processName]
	if ok == false {
		return map[int]ProcessStatus{}
	}
	return processStatusHistory
}

// findProcessStatusInHistory finds the ProcessStatus
// whose timestamp is the biggest while its timestamp is smaller than given timestamp
func findProcessStatusInHistory(processStatusHistory map[int]ProcessStatus, pid int) ProcessStatus {
	if len(processStatusHistory) == 0 {
		return NewProcessStatus(pid, ProcessNeverStarted)
	}
	mostRecentProcessStatus := NewProcessStatus(pid, ProcessNeverStarted)
	mostRecentProcessStatus.SetTimestamp(time.Time{})

	for _, processStatus := range processStatusHistory {
		if processStatus.Pid() == pid {
			mostRecentProcessStatus = processStatus
		}
	}
	return mostRecentProcessStatus
}
