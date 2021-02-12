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
		require.True(t, pir.IsExecuting("go test"))
	}
}

func CheckNotExecuting() func(*testing.T) {
	return func(t *testing.T) {
		pir := NewProcessInfoReader()
		require.False(t, pir.IsExecuting("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS"))
	}
}

func CheckDirectoryOfExecutingBinary() func(*testing.T) {
	return func(t *testing.T) {
		pir := NewProcessInfoReader()
		require.Equal(t,
			getCurrentFileLocation(),
			pir.GetLocationOfExecutedBinary("go test"),
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
		require.Equal(t,
			"alarm",
			pir.GetPackageNameOfGolangProcess("go test"),
		)
		require.Equal(t,
			"",
			pir.GetPackageNameOfGolangProcess("THERE WILL BE NO PROCESS WHOSE NAME LIKE THIS"),
		)
	}
}
