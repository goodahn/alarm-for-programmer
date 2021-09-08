package alarm

import alarm "github.com/goodahn/alarm-for-programmer"

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
