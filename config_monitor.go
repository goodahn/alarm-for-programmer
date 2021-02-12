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
	configPath    string
	config        map[string]interface{}
	period        time.Duration
	start         bool
	mutexForStart sync.Mutex
}

func NewConfigMonitor(configPath string) *ConfigMonitor {
	cm := &ConfigMonitor{}
	cm.Init(configPath)
	cm.SetPeriod(defaultPeriod)
	cm.Start()
	time.Sleep(defaultPeriod)
	return cm
}

func (cm *ConfigMonitor) Init(configPath string) {
	cm.configPath = configPath
	cm.config = map[string]interface{}{}
	cm.UpdateConfig()
}

func (cm *ConfigMonitor) Start() {
	cm.mutexForStart.Lock()
	defer cm.mutexForStart.Unlock()

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
	cm.mutexForStart.Lock()
	defer cm.mutexForStart.Unlock()
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
	cm.mutexForStart.Lock()
	defer cm.mutexForStart.Unlock()
	return cm.config
}
