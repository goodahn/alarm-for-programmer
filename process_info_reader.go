package alarm

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type ProcessInfoReader struct {
	processInfoList []ProcessInfo
	period          time.Duration
	stop            bool
}

func NewProcessInfoReader() ProcessInfoReader {
	pir := ProcessInfoReader{}
	pir.SetPeriod(500 * time.Millisecond)
	pir.Start()
	time.Sleep(500 * time.Millisecond)
	return pir
}

func (pir *ProcessInfoReader) Start() {
	go func() {
		for {
			if pir.stop == true {
				break
			}

			processInfoList, err := GetProcessInfoList()
			if err == nil {
				pir.processInfoList = processInfoList
			}

			time.Sleep(pir.period)
		}
	}()
}

func (pir *ProcessInfoReader) Stop() {
	pir.stop = true
}

func (pir *ProcessInfoReader) SetPeriod(period time.Duration) {
	pir.period = period
}

func (pir *ProcessInfoReader) IsExecuting(processName string) bool {
	return pir.findProcessInfoByName(processName) != ProcessInfo{}
}

// there is no "/" at the last of file location
func (pir *ProcessInfoReader) GetLocationOfExecutedBinary(processName string) string {
	processInfo := pir.findProcessInfoByName(processName)
	return processInfo.BinaryLocation()
}

func (pir *ProcessInfoReader) findProcessInfoByName(processName string) ProcessInfo {
	for _, processInfo := range pir.processInfoList {
		if strings.Contains(processInfo.Cmd(), processName) {
			return processInfo
		}
	}
	return ProcessInfo{}
}

func (pir *ProcessInfoReader) GetPackageNameOfGolangProcess(processName string) string {
	binaryLocation := pir.GetLocationOfExecutedBinary(processName)
	packageName := getPackageNameOfGolangProcessFromDirectory(binaryLocation)
	return packageName
}

func getPackageNameOfGolangProcessFromDirectory(directory string) (packageName string) {
	fileList, err := ioutil.ReadDir(directory)
	if err != nil {
		return ""
	}
	for _, f := range fileList {
		path := fmt.Sprintf("%s/%s", directory, f.Name())
		fo, err := os.Open(path)
		if err != nil {
			continue
		}
		if packageName = getPackageNameOfGolangProcessFromFile(fo); packageName != "" {
			break
		}
		fo.Close()
	}
	return packageName
}

func getPackageNameOfGolangProcessFromFile(file *os.File) (packageName string) {
	reader := bufio.NewReader(file)
	_line, _, err := reader.ReadLine()
	if err != nil {
		return ""
	}

	line := string(_line)
	if strings.Contains(line, "package") {
		packageName = strings.Split(line, " ")[1]
	}
	return packageName
}
