package alarm

import (
	"time"

	alarm "github.com/goodahn/AlarmForProgrammer"
)

type Alarmer interface {
	GetMonitoringCommandList() []string
	GetMonitoringPeriod() time.Duration
	GetAlarmConfig() map[string]string
	GetAlarmCountMap() map[string]int
	GetTotalAlarmCountOfMonitoringCommand(monitoringCommand string) int

	Init(configMonitor *alarm.ConfigMonitor)

	Start()
	IsStarted() bool

	Stop()
}

func NewAlarmer(configPath string) Alarmer {
	cm := alarm.NewConfigMonitor(configPath)

	alarmConfig := cm.GetAlarmConfig()
	if alarmConfig["type"] == "slack-webhook" {
		am := SlackWebHookAlarmer{}
		am.Init(cm)
		am.Start()
		return &am
	} else {
		panic("not implemented type of alarmer")
	}
}
