package alarm

type Alarmer struct {
}

func NewAlarmer(targetProcessList []string) *Alarmer {
	alarmer := &Alarmer{}
	return alarmer
}

func (a *Alarmer) GetAlarmCount(processName string) int {
	return 0
}
