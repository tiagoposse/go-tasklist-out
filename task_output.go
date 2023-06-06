package out

import (
	"sync"
)

type TaskOutput struct {
	sync.RWMutex
	value string
}

func (to *TaskOutput) Get() string {
	to.RLock()
	defer to.RUnlock()
	return to.value
}

func (to *TaskOutput) Set(value string) {
	to.Lock()
	to.value = value
	to.Unlock()
}

func (to *TaskOutput) Append(value string) {
	to.Lock()
	to.value = to.value + value
	to.Unlock()
}

func (to *TaskOutput) GetAndClear() string {
	to.Lock()
	val := to.value
	to.value = ""
	to.Unlock()

	return val
}

func (to *TaskOutput) Length() int {
	to.RLock()
	defer to.RUnlock()
	return len(to.value)
}

func (to *TaskOutput) GetLastNLines(numLines, limit int) ([]byte, int) {
	result := []byte{}
	curLines := 0
	line := []byte{}
	offset := len(to.value) - 1

	appendLine := func(line []byte) {
		curLine := 0
		for i := 0; i < len(line); i++ {
			result = append(result, line[i])

			if i == len(line)-(limit*curLine) {
				result = append(result, '\n')
				curLine++
				curLines++
			}
		}
	}

	to.RLock()
	for offset > -1 && curLines < numLines {
		char := to.value[offset]
		line = append(line, char)

		if char == byte('\n') {
			appendLine(line)
			curLines++
			line = []byte{}
		}

		offset--
	}
	to.RUnlock()

	if len(line) > 0 {
		appendLine(result)
		curLines++
	}

	inverted := []byte{}
	for i := len(result) - 1; i > -1; i-- {
		inverted = append(inverted, result[i])
	}

	return inverted, curLines
}

func NewTaskOutput() *TaskOutput {
	return &TaskOutput{
		value: "",
	}
}
