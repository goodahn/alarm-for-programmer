package alarm

import (
	"sync"
	"time"
)

type ProcessInfoMonitor struct {
	namePatternList                   []string
	processStatusHistoryByNamePattern map[string](map[int]([]ProcessStatus))
	period                            time.Duration
	start                             bool
	mutex                             sync.Mutex

	now time.Time

	processInfoReader *ProcessInfoReader
}

func NewProcessInfoMonitor(namePatternList []string) *ProcessInfoMonitor {
	pim := &ProcessInfoMonitor{}
	pim.Init()
	pim.SetNamePatternList(namePatternList)
	pim.Start()
	time.Sleep(defaultPeriod)
	return pim
}

func (pim *ProcessInfoMonitor) Init() {
	pim.processStatusHistoryByNamePattern = map[string](map[int]([]ProcessStatus)){}
	pim.processInfoReader = NewProcessInfoReader()
	pim.SetPeriod(defaultPeriod)
}

// there will be only one go routine for monitoring ProcessInfo
// mutex is used to achieve it
func (pim *ProcessInfoMonitor) Start() {
	pim.mutex.Lock()
	defer pim.mutex.Unlock()

	if pim.start {
		return
	}
	pim.start = true

	pim.processInfoReader.Start()
	go func() {
		for {
			if !pim.start {
				break
			}

			pim.now = time.Now()
			pim.updateTargetProcessStatusHistory()
		}
	}()
}

func (pim *ProcessInfoMonitor) updateTargetProcessStatusHistory() {
	for _, namePattern := range pim.namePatternList {
		pim.updateProcessStatusHistoryByNamePattern(namePattern)
	}
}

func (pim *ProcessInfoMonitor) updateProcessStatusHistoryByNamePattern(namePattern string) {
	changedProcessStatusMap := pim.getChangedProcessStatus(namePattern)
	if len(changedProcessStatusMap) == 0 {
		return
	}
	pim.mutex.Lock()
	defer pim.mutex.Unlock()
	for pid, changedProcessStatus := range changedProcessStatusMap {
		pim.processStatusHistoryByNamePattern[namePattern][pid] = append(
			pim.processStatusHistoryByNamePattern[namePattern][pid],
			changedProcessStatus,
		)
	}
}

func (pim *ProcessInfoMonitor) getChangedProcessStatus(namePattern string) map[int]ProcessStatus {
	changedProcessStatusHistory := map[int]ProcessStatus{}

	pim.mutex.Lock()
	defer pim.mutex.Unlock()
	processStatusHistory := pim.processStatusHistoryByNamePattern[namePattern]

	pidList := pim.processInfoReader.GetPidListByName(namePattern)
	for _, pid := range pidList {
		latestProcessStatus := pim.getUpdatedProcessStatusInLogWithTimestamp(processStatusHistory, pid)
		if (latestProcessStatus == ProcessStatus{}) {
			continue
		}
		changedProcessStatusHistory[pid] = latestProcessStatus
	}
	return changedProcessStatusHistory
}

func (pim *ProcessInfoMonitor) getUpdatedProcessStatusInLogWithTimestamp(processStatusHistory map[int]([]ProcessStatus), pid int) ProcessStatus {
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
	for _, namePattern := range namePatternList {
		pim.processStatusHistoryByNamePattern[namePattern] = map[int]([]ProcessStatus){}
	}
}

func (pim *ProcessInfoMonitor) GetProcessStatusLogByNamePattern(namePattern string) map[int]([]ProcessStatus) {
	pim.mutex.Lock()
	defer pim.mutex.Unlock()
	return pim.getProcessStatusHistory(namePattern)
}

func (pim *ProcessInfoMonitor) getProcessStatusHistory(namePattern string) map[int]([]ProcessStatus) {
	processStatusHistory, ok := pim.processStatusHistoryByNamePattern[namePattern]
	if !ok {
		return map[int]([]ProcessStatus){}
	}
	return processStatusHistory
}

// findProcessStatusInHistory finds the ProcessStatus
// whose timestamp is the biggest while its timestamp is smaller than given timestamp
func findProcessStatusInHistory(wholeProcessStatusHistory map[int]([]ProcessStatus), pid int) ProcessStatus {
	if len(wholeProcessStatusHistory) == 0 {
		return NewProcessStatus(pid, ProcessNeverStarted)
	}
	mostRecentProcessStatus := NewProcessStatus(pid, ProcessNeverStarted)
	mostRecentProcessStatus.SetTimestamp(time.Time{})

	for _pid, processStatusHistory := range wholeProcessStatusHistory {
		if pid != _pid {
			continue
		}
		mostRecentProcessStatus = processStatusHistory[len(processStatusHistory)-1]
	}
	return mostRecentProcessStatus
}
