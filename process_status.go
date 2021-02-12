package alarm

import "time"

const (
	NotFoundInTargetProcessList = "NOT_FOUND_IN_TARGET_PROCESS_LIST"
	ProcessNeverStarted         = "PROCESS_NEVER_STARTED"
	ProcessStarted              = "PROCESS_STARTED"
	ProcessFinished             = "PROCESS_FINISHED"
)

var (
	CheckProcessStatus = map[string]bool{
		NotFoundInTargetProcessList: true,
		ProcessNeverStarted:         true,
		ProcessStarted:              true,
		ProcessFinished:             true,
	}
)

type ProcessStatus struct {
	status    string
	timestamp time.Time
}

func NewProcessStatus(status string) ProcessStatus {
	ps := ProcessStatus{}
	ps.Init(status)
	return ps
}

func (ps *ProcessStatus) Init(status string) {
	_, ok := CheckProcessStatus[status]
	if ok == false {
		return
	}
	ps.status = status
	ps.timestamp = time.Now()
}

func (ps *ProcessStatus) Status() string {
	return ps.status
}

func (ps *ProcessStatus) TimeStamp() time.Time {
	return ps.timestamp
}

func (ps *ProcessStatus) TimeStampInString() string {
	return ps.timestamp.String()
}

func (ps *ProcessStatus) SetTimestamp(timestamp time.Time) {
	ps.timestamp = timestamp
}
