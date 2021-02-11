package alarm

type ProcessInfoReader struct {
	targetProcessList []string
}

func NewProcessInfoReader() ProcessInfoReader {
	pir := ProcessInfoReader{}
	return pir
}

func (pir *ProcessInfoReader) IsExecuting(processName string) (executing bool) {
	return executing
}

func (pir *ProcessInfoReader) GetDirectoryOfExecutedBinary(processName string) (directory string) {
	return directory
}
