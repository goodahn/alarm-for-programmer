package alarm

import (
	"sync"
	"time"
)

type ProcessInfoMonitor struct {
	targetProcessList       []string
	processStatusHistoryMap map[string](map[int]ProcessStatus)
	period                  time.Duration
	start                   bool
	mutexForStart           sync.Mutex

	now time.Time

	processInfoReader *ProcessInfoReader
}

func NewProcessInfoMonitor(targetProcessList []string) *ProcessInfoMonitor {
	pim := &ProcessInfoMonitor{}
	pim.Init()
	pim.SetTargetProcessList(targetProcessList)
	pim.SetPeriod(defaultPeriod)
	pim.Start()
	time.Sleep(defaultPeriod)
	return pim
}

func (pim *ProcessInfoMonitor) Init() {
	pim.processStatusHistoryMap = map[string](map[int]ProcessStatus){}
	pim.processInfoReader = NewProcessInfoReader()
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (pim *ProcessInfoMonitor) Start() {
	pim.mutexForStart.Lock()
	defer pim.mutexForStart.Unlock()

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
	for _, targetProcess := range pim.targetProcessList {
		pim.updateProcessStatusHistory(targetProcess)
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

	pim.mutexForStart.Lock()
	defer pim.mutexForStart.Unlock()
	processStatusHistory := pim.processStatusHistoryMap[processName]

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
	pim.mutexForStart.Lock()
	defer pim.mutexForStart.Unlock()
	pim.processStatusHistoryMap[processName] = processStatusHistory
}

func (pim *ProcessInfoMonitor) Stop() {
	pim.start = false
	pim.processInfoReader.Stop()
}

func (pim *ProcessInfoMonitor) SetPeriod(period time.Duration) {
	pim.period = period
	pim.processInfoReader.SetPeriod(period)
}

func (pim *ProcessInfoMonitor) SetTargetProcessList(targetProcessList []string) {
	pim.targetProcessList = targetProcessList
	for _, processName := range targetProcessList {
		pim.processStatusHistoryMap[processName] = map[int]ProcessStatus{}
	}
}

func (pim *ProcessInfoMonitor) GetCurrentProcessStatusHistoryByName(processName string) map[int]ProcessStatus {
	pim.mutexForStart.Lock()
	defer pim.mutexForStart.Unlock()
	return pim.getProcessStatusHistory(processName)
}

func (pim *ProcessInfoMonitor) getProcessStatusHistory(processName string) map[int]ProcessStatus {
	processStatusHistory, ok := pim.processStatusHistoryMap[processName]
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
