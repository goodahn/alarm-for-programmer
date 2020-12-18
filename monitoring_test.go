package alarm_test

import (
	"os/exec"
	"path"
	"runtime"
	"sync"
	"testing"
	"time"

	alarm "github.com/goodahn/AlarmForProgrammer"
	"github.com/stretchr/testify/require"
)

func TestMonitoringCommand(t *testing.T) {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dirpath := path.Dir(currentFilePath)

	commandNum := 10
	m := alarm.NewCommandMonitor(dirpath + "/" + "test_config.json")
	m.SetMonitoringPeriod(100 * time.Millisecond)
	m.Start()

	time.Sleep(50 * time.Millisecond)

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
	time.Sleep(500 * time.Millisecond)

	require.Equal(t, commandNum, m.TotalAlarmCount())
}
