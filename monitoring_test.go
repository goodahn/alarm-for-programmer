package alarm_test

import (
	"os/exec"
	"sync"
	"testing"
	"time"

	alarm "github.com/goodahn/AlarmForProgrammer"
	"github.com/stretchr/testify/require"
)

func TestMonitoringCommand(t *testing.T) {
	cmd := "test_monitoring_command"

	commandNum := 10
	m := alarm.NewCommandMonitor([]string{
		cmd,
	})
	m.SetMonitoringPeriod(100 * time.Millisecond)
	m.Start()

	time.Sleep(50 * time.Millisecond)

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
