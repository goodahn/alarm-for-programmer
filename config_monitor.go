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
	configPath string
	config     map[string]interface{}
	period     time.Duration
	start      bool
	mutex      sync.Mutex
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
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.start == true {
		return
	}
	cm.start = true

	go func(cm *ConfigMonitor) {
		for {
			if cm.start == false {
				break
			}

			cm.UpdateConfig()
			time.Sleep(cm.period)
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
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
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
	cm.start = false
}

func (cm *ConfigMonitor) SetPeriod(period time.Duration) {
	cm.period = period
}

func (cm *ConfigMonitor) GetConfig() map[string]interface{} {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	config := map[string]interface{}{}
	for key, val := range cm.config {
		config[key] = val
	}
	return config
}

func (cm *ConfigMonitor) GetNamePatternList() []string {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	namePatternList := []string{}

	rawNamePatternList := cm.config["namePatternList"].([]interface{})
	for _, rawNamePattern := range rawNamePatternList {
		namePattern, ok := rawNamePattern.(string)
		if !ok {
			continue
		}
		namePatternList = append(namePatternList, namePattern)
	}
	return namePatternList
}

func (cm *ConfigMonitor) GetAlarmConfig() map[string]string {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	alarmConfig := map[string]string{}

	for key, val := range cm.config["alarmConfig"].(map[string]interface{}) {
		alarmConfig[key] = val.(string)
	}
	return alarmConfig
}
