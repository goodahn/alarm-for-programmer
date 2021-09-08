package alarm

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type ProcessInfoReader struct {
	processInfoList               []ProcessInfo
	monitoringPeriod              time.Duration
	isStarted                     bool
	mutexForSynchronousMethodCall sync.Mutex
}

func NewProcessInfoReader() *ProcessInfoReader {
	pir := &ProcessInfoReader{}
	pir.Init()
	pir.Start()
	time.Sleep(defaultPeriod)
	return pir
}

func (pir *ProcessInfoReader) Init() {
	pir.SetPeriod(defaultPeriod)
}

func (pir *ProcessInfoReader) Start() {
	pir.mutexForSynchronousMethodCall.Lock()
	defer pir.mutexForSynchronousMethodCall.Unlock()

	if pir.isStarted {
		return
	}
	pir.isStarted = true

	go func() {
		for {
			if !pir.isStarted {
				break
			}

			processInfoList, err := GetProcessInfoList()
			if err == nil {
				pir.processInfoList = processInfoList
			}

			time.Sleep(pir.monitoringPeriod)
		}
	}()
}

func (pir *ProcessInfoReader) Stop() {
	pir.isStarted = false
}

func (pir *ProcessInfoReader) SetPeriod(period time.Duration) {
	pir.monitoringPeriod = period
}

func (pir *ProcessInfoReader) GetPidListByName(namePattern string) []int {
	pidList := []int{}
	for _, processInfo := range pir.processInfoList {
		if strings.Contains(processInfo.Cmd(), namePattern) {
			pidList = append(pidList, processInfo.Pid())
		}
	}
	return pidList
}

func (pir *ProcessInfoReader) IsExecuting(pid int) bool {
	return pir.findProcessInfoByPid(pid) != ProcessInfo{}
}

// there is no "/" at the last of file location
func (pir *ProcessInfoReader) GetLocationOfExecutedBinary(pid int) string {
	processInfo := pir.findProcessInfoByPid(pid)
	return processInfo.BinaryLocation()
}

func (pir *ProcessInfoReader) findProcessInfoByPid(pid int) ProcessInfo {
	for _, processInfo := range pir.processInfoList {
		if processInfo.Pid() == pid {
			return processInfo
		}
	}
	return ProcessInfo{}
}

func (pir *ProcessInfoReader) GetPackageNameOfGolangProcess(pid int) string {
	binaryLocation := pir.GetLocationOfExecutedBinary(pid)
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
