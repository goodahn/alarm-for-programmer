package alarm

import (
	"sync"
	"time"
)

type ProcessInfoMonitor struct {
	targetProcessList       []string
	processStatusHistoryMap map[string][]ProcessStatus
	period                  time.Duration
	start                   bool
	mutexForStart           sync.Mutex

	now time.Time

	processInfoReader ProcessInfoReader
}

func NewProcessInfoMonitor(targetProcessList []string) *ProcessInfoMonitor {
	pim := &ProcessInfoMonitor{}
	pim.Init()
	pim.SetTargetProcessList(targetProcessList)
	pim.SetPeriod(500 * time.Millisecond)
	pim.Start()
	time.Sleep(500 * time.Millisecond)
	return pim
}

func (pim *ProcessInfoMonitor) Init() {
	pim.processStatusHistoryMap = map[string][]ProcessStatus{}
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
	changedProcessStatus := pim.checkProcessStatusChanged(processName)
	if (changedProcessStatus == ProcessStatus{}) {
		return
	}
	pim.appendProcessStatusHistory(processName, changedProcessStatus)
}

func (pim *ProcessInfoMonitor) checkProcessStatusChanged(processName string) ProcessStatus {
	processStatusHistory := pim.processStatusHistoryMap[processName]
	isExecuting := pim.processInfoReader.IsExecuting(processName)
	if isExecuting {
		// process is started after its previous execution
		mostRecentProcessStatus := findProcessStatusInHistoryWithTimestamp(processStatusHistory, pim.now)
		if mostRecentProcessStatus.Status() == ProcessFinished ||
			mostRecentProcessStatus.Status() == ProcessNeverStarted {
			return NewProcessStatus(ProcessStarted)
		}
	} else {
		// process is finshed after start
		mostRecentProcessStatus := findProcessStatusInHistoryWithTimestamp(processStatusHistory, pim.now)
		if mostRecentProcessStatus.Status() == ProcessFinished {
			return NewProcessStatus(ProcessFinished)
		}
	}
	return ProcessStatus{}
}

func (pim *ProcessInfoMonitor) appendProcessStatusHistory(processName string, processStatus ProcessStatus) {
	processStatusHistory := pim.processStatusHistoryMap[processName]
	processStatusHistory = append(processStatusHistory, processStatus)
	pim.processStatusHistoryMap[processName] = processStatusHistory
}

func (pim *ProcessInfoMonitor) Stop() {
	pim.start = false
	pim.processInfoReader.Stop()
}

func (pim *ProcessInfoMonitor) SetPeriod(period time.Duration) {
	pim.period = period
}

func (pim *ProcessInfoMonitor) SetTargetProcessList(targetProcessList []string) {
	pim.targetProcessList = targetProcessList
	for _, processName := range targetProcessList {
		pim.processStatusHistoryMap[processName] = []ProcessStatus{}
	}
}

func (pim *ProcessInfoMonitor) GetCurrentProcessStatus(processName string) ProcessStatus {
	now := time.Now()
	return pim.getProcessStatusWithTimestamp(processName, now)
}

func (pim *ProcessInfoMonitor) getProcessStatusWithTimestamp(processName string, timestamp time.Time) ProcessStatus {
	processStatusHistory, ok := pim.processStatusHistoryMap[processName]
	if ok == false {
		return NewProcessStatus(NotFoundInTargetProcessList)
	}

	processStatus := findProcessStatusInHistoryWithTimestamp(processStatusHistory, timestamp)
	return processStatus
}

// findProcessStatusInHistory finds the ProcessStatus
// whose timestamp is the biggest while its timestamp is smaller than given timestamp
func findProcessStatusInHistoryWithTimestamp(processStatusHistory []ProcessStatus, timestamp time.Time) ProcessStatus {
	if len(processStatusHistory) == 0 {
		return NewProcessStatus(ProcessNeverStarted)
	}
	mostRecentProcessStatus := NewProcessStatus(ProcessNeverStarted)
	mostRecentProcessStatus.SetTimestamp(time.Time{})

	for _, processStatus := range processStatusHistory {
		if timestamp.After(processStatus.TimeStamp()) &&
			mostRecentProcessStatus.TimeStamp().Before(processStatus.TimeStamp()) {
			mostRecentProcessStatus = processStatus
		}
	}
	return mostRecentProcessStatus
}
