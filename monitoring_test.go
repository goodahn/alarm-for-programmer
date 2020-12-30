package alarm_test

import (
	"os"
	"os/exec"
	"path"
	"runtime"
	"sync"
	"testing"
	"time"

	alarm "github.com/goodahn/AlarmForProgrammer"
	"github.com/stretchr/testify/require"
)

func TestGetCwd(t *testing.T) {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dirpath := path.Dir(currentFilePath)

	m := alarm.NewCommandMonitor(dirpath + "/" + "test_config.json")
	m.SetMonitoringPeriod(100 * time.Millisecond)
	m.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c := exec.Command("bash", "test_monitoring_command.sh")
		err := c.Run()
		require.Nil(t, err)
	}()

	time.Sleep(500 * time.Millisecond)

	pids := m.GetPids("test_monitoring_command")
	targetDir := m.GetCwd(pids[0])

	m.Stop()
	wg.Wait()

	dir, err := os.Getwd()
	require.Nil(t, err)
	require.Equal(t, dir, targetDir)
}

func TestGetPackageNameOfGo(t *testing.T) {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dirpath := path.Dir(currentFilePath)

	m := alarm.NewCommandMonitor(dirpath + "/" + "test_config.json")
	m.SetMonitoringPeriod(100 * time.Millisecond)
	m.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c := exec.Command("go", "test", "-run", "TestForGetPackageNameOfGo")
		err := c.Run()
		require.Nil(t, err)
	}()

	time.Sleep(500 * time.Millisecond)

	pids := m.GetPids("go test")
	goTestPackageName := m.GetGoPackageName(pids[0])

	m.Stop()
	wg.Wait()

	require.Equal(t, "alarm_test", goTestPackageName)
}

func TestForGetPackageNameOfGo(t *testing.T) {
	time.Sleep(1 * time.Second)
}

func TestMonitoringCommand(t *testing.T) {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dirpath := path.Dir(currentFilePath)

	commandNum := 10
	m := alarm.NewCommandMonitor(dirpath + "/" + "test_config.json")
	m.SetMonitoringPeriod(100 * time.Millisecond)
	m.Start()

	require.Equal(t, "test_monitoring_command", m.CommandList()[0])

	wg := sync.WaitGroup{}
	for i := 0; i < commandNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := exec.Command("bash", "test_monitoring_command.sh")
			err := c.Run()
			require.Nil(t, err)
		}()
	}
	wg.Wait()

	//time.Sleep(500 * time.Millisecond)
	time.Sleep(3 * time.Second)

	require.Equal(t, commandNum, m.TotalAlarmCount())
}
