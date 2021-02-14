package main

import (
	"path"
	"runtime"
	"time"

	alarm "github.com/goodahn/AlarmForProgrammer"
)

func main() {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dirpath := path.Dir(currentFilePath)

	alarmer := alarm.NewAlarmer(dirpath + "/config.json")
	alarmer.Start()
	for {
		time.Sleep(3 * time.Second)
	}
}
