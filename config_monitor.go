package alarm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

const (
	defaultPeriod = 250 * time.Millisecond
)

type ConfigMonitor struct {
	configPath       string
	config           map[string]interface{}
	monitoringPeriod time.Duration
	isStarted        bool

	mutexForSynchronousMethodCall sync.Mutex
	mutexForConfig                sync.Mutex
}

func NewConfigMonitor(configPath string) *ConfigMonitor {
	cm := &ConfigMonitor{}
	cm.Init(configPath)
	cm.Start()
	time.Sleep(defaultPeriod)
	return cm
}

func (cm *ConfigMonitor) Init(configPath string) {
	cm.configPath = configPath
	cm.config = map[string]interface{}{}
	cm.UpdateConfig()
	cm.SetPeriod(defaultPeriod)
}

func (cm *ConfigMonitor) Start() {
	cm.mutexForSynchronousMethodCall.Lock()
	defer cm.mutexForSynchronousMethodCall.Unlock()

	if cm.isStarted {
		return
	}
	cm.isStarted = true

	go func(cm *ConfigMonitor) {
		for {
			if !cm.isStarted {
				break
			}

			cm.UpdateConfig()
			time.Sleep(cm.monitoringPeriod)
		}
	}(cm)
}

func (cm *ConfigMonitor) UpdateConfig() {
	config, err := readJsonFile(cm.configPath)
	if err != nil {
		errMsg := fmt.Sprintf("error occured during reading config file: %v", err)
		fmt.Println(errMsg)
		return
	}
	cm.mutexForConfig.Lock()
	defer cm.mutexForConfig.Unlock()
	cm.config = map[string]interface{}{}
	for key, val := range config {
		cm.config[key] = val
	}
}

func readJsonFile(configPath string) (map[string]interface{}, error) {
	config := map[string]interface{}{}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	if len(data) == 0 {
		return config, nil
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func (cm *ConfigMonitor) Stop() {
	cm.isStarted = false
}

func (cm *ConfigMonitor) SetPeriod(period time.Duration) {
	cm.monitoringPeriod = period
}

func (cm *ConfigMonitor) GetConfig() map[string]interface{} {
	cm.mutexForConfig.Lock()
	defer cm.mutexForConfig.Unlock()
	config := map[string]interface{}{}
	for key, val := range cm.config {
		config[key] = val
	}
	return config
}

func (cm *ConfigMonitor) GetMonitoringCommandList() []string {
	cm.mutexForConfig.Lock()
	defer cm.mutexForConfig.Unlock()
	monitoringCommandList := []string{}

	rawMonitoringCommandList := cm.config["monitoringCommandList"].([]interface{})
	for _, rawMonitoringCommand := range rawMonitoringCommandList {
		namePattern, ok := rawMonitoringCommand.(string)
		if !ok {
			continue
		}
		monitoringCommandList = append(monitoringCommandList, namePattern)
	}
	return monitoringCommandList
}

func (cm *ConfigMonitor) GetAlarmConfig() map[string]string {
	cm.mutexForConfig.Lock()
	defer cm.mutexForConfig.Unlock()
	alarmConfig := map[string]string{}

	for key, val := range cm.config["alarmConfig"].(map[string]interface{}) {
		alarmConfig[key] = val.(string)
	}
	return alarmConfig
}
