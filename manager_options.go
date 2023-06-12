package out

import (
	"io"

	"github.com/fatih/color"
)

func WithColorScheme(schemes map[VerbosityLevel]color.Attribute) OutputManagerOption {
	return func(com *OutputManager) {
		for k, v := range schemes {
			com.colorSchemes[k] = color.New(v).SprintfFunc()
		}
	}
}

func WithOut(out io.Writer) OutputManagerOption {
	return func(com *OutputManager) {
		com.out = out
	}
}
func WithRawOutput() OutputManagerOption {
	return func(com *OutputManager) {
		com.keepOutput = true
		com.keepTasksOnComplete = true
		com.tty = false
	}
}

func WithKeepOutput() OutputManagerOption {
	return func(com *OutputManager) {
		com.keepTasksOnComplete = true
		com.keepOutput = true
	}
}

func WithVerbosity(verb VerbosityLevel) OutputManagerOption {
	return func(com *OutputManager) {
		com.verbosity = verb
	}
}

func WithLogsRoot(path string) OutputManagerOption {
	return func(com *OutputManager) {
		com.logsRoot = path
	}
}
