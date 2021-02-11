package alarm_test

import (
	"runtime"
	"strings"
	"testing"

	alarm "github.com/goodahn/AlarmForProgrammer"
	"github.com/stretchr/testify/require"
)

func TestProcessInfoReader(t *testing.T) {
	t.Run("TargetProcessList", CheckTargetProcessList())
	t.Run("PackageNameOfGoLangProcess", CheckPackageNameOfGoLangProcess())
}

func CheckTargetProcessList() func(*testing.T) {
	return func(t *testing.T) {
		t.Run("Executing", CheckExecuting())
		t.Run("NotExecuting", CheckNotExecuting())
	}
}

// check that ProcessInfoReader catches
// currently executed "go test *" process well
func CheckExecuting() func(*testing.T) {
	return func(t *testing.T) {
		pir := alarm.NewProcessInfoReader()
		require.True(t, pir.IsExecuting("go test"))
	}
}

func CheckNotExecuting() func(*testing.T) {
	return func(t *testing.T) {
		pir := alarm.NewProcessInfoReader()
		require.False(t, pir.IsExecuting("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS"))
	}
}

func CheckPackageNameOfGoLangProcess() func(*testing.T) {
	return func(t *testing.T) {
		t.Run("DirectoryOfExecutingBinary", CheckDirectoryOfExecutingBinary())
	}
}

func CheckDirectoryOfExecutingBinary() func(*testing.T) {
	return func(t *testing.T) {
		pir := alarm.NewProcessInfoReader()

		require.Equal(t,
			pir.GetDirectoryOfExecutedBinary("go test"),
			getCurrentFileLocation())
	}
}

func getCurrentFileLocation() (fileLocation string) {
	_, currentFilePath, _, _ := runtime.Caller(0)

	parsedFilePath := strings.Split(currentFilePath, "/")

	fileLocation = strings.Join(parsedFilePath[0:len(parsedFilePath)-1], "/")
	return fileLocation
}
