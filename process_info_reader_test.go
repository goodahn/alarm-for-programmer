package alarm

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessInfoReader(t *testing.T) {
	t.Run("ExecutingStatus", CheckExecutingStatus())
	t.Run("DirectoryOfExecutingBinary", CheckDirectoryOfExecutingBinary())
	t.Run("PackageNameOfGoLangProcess", CheckPackageNameOfGoLangProcess())
}

func CheckExecutingStatus() func(*testing.T) {
	return func(t *testing.T) {
		t.Run("Executing", CheckExecuting())
		t.Run("NotExecuting", CheckNotExecuting())
	}
}

// check that ProcessInfoReader catches
// currently executed "go test *" process well
func CheckExecuting() func(*testing.T) {
	return func(t *testing.T) {
		pir := NewProcessInfoReader()
		defer pir.Stop()

		pidList := pir.GetPidListByName("go test")
		require.Equal(t, 1, len(pidList))
		require.True(t, pir.IsExecuting(pidList[0]))
	}
}

func CheckNotExecuting() func(*testing.T) {
	return func(t *testing.T) {
		pir := NewProcessInfoReader()
		defer pir.Stop()

		pidList := pir.GetPidListByName("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS")
		require.Equal(t, 0, len(pidList))
	}
}

func CheckDirectoryOfExecutingBinary() func(*testing.T) {
	return func(t *testing.T) {
		pir := NewProcessInfoReader()
		defer pir.Stop()

		pidList := pir.GetPidListByName("go test")
		require.Equal(t, 1, len(pidList))
		require.Contains(t,
			pir.GetLocationOfExecutedBinary(pidList[0]),
			getCurrentFileLocation(),
		)
	}
}

func getCurrentFileLocation() (fileLocation string) {
	_, currentFilePath, _, _ := runtime.Caller(0)

	parsedFilePath := strings.Split(currentFilePath, "/")

	// there is no "/" at the last of file location
	fileLocation = strings.Join(parsedFilePath[0:len(parsedFilePath)-1], "/")
	return fileLocation
}

func CheckPackageNameOfGoLangProcess() func(*testing.T) {
	return func(t *testing.T) {
		pir := NewProcessInfoReader()
		defer pir.Stop()

		pidList := pir.GetPidListByName("go test")
		require.Equal(t, 1, len(pidList))
		require.Equal(t,
			"alarm",
			pir.GetPackageNameOfGolangProcess(pidList[0]),
		)

		pidList = pir.GetPidListByName("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS")
		require.Equal(t, 0, len(pidList))
	}
}
