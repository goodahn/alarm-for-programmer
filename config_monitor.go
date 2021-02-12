package alarm

type ConfigMonitor struct {
	config map[string]string
}

func NewConfigMonitor(configPath string) ConfigMonitor {
	cm := ConfigMonitor{}
	return cm
}

func (cm *ConfigMonitor) GetConfig() map[string]string {
	return cm.config
}
