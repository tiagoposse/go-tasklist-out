package out

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	atomics "github.com/tiagoposse/go-sync-types"
)

type TaskLoggerImplOption func(tl *TaskLoggerImpl)

type TaskLoggerImpl struct {
	title       *atomics.Value[string]
	format      func(msg string, verb VerbosityLevel) string
	err         *atomics.Value[bool]
	complete    *atomics.Value[bool]
	hidden      *atomics.Value[bool]
	execLog     *atomics.Slice[byte]
	verbosity   VerbosityLevel
	logFilePath string
	*os.File
}

func NewTaskLogger(title string, opts ...TaskLoggerImplOption) (*TaskLoggerImpl, error) {
	tl := &TaskLoggerImpl{
		title:       &atomics.Value[string]{},
		format:      func(msg string, verb VerbosityLevel) string { return msg },
		err:         &atomics.Value[bool]{},
		complete:    &atomics.Value[bool]{},
		hidden:      &atomics.Value[bool]{},
		verbosity:   Info,
		logFilePath: "",
		execLog:     atomics.NewSlice[byte](),
	}

	for _, o := range opts {
		o(tl)
	}

	if tl.logFilePath != "" {
		if err := os.MkdirAll(filepath.Dir(tl.logFilePath), os.ModePerm); err != nil {
			return nil, err
		}
		if f, err := os.Create(tl.logFilePath); err != nil {
			return nil, err
		} else {
			tl.File = f
		}
	}

	return tl, nil
}

func (tl *TaskLoggerImpl) write(msg string, verb VerbosityLevel) {
	if verb > tl.verbosity {
		return
	}

	tl.execLog.Append([]byte(tl.format(msg, verb))...)
}

func (tl *TaskLoggerImpl) Warn(msg string) {
	tl.write(msg, Warning)
}

func (tl *TaskLoggerImpl) Warnf(msg string, args ...any) {
	tl.Warn(fmt.Sprintf(msg, args...))
}

func (tl *TaskLoggerImpl) Warnln(msg string, args ...any) {
	tl.Warnf(msg+"\n", args...)
}

func (tl *TaskLoggerImpl) Error(msg string) {
	tl.err.Set(true)
	tl.write(msg, Error)
}

func (tl *TaskLoggerImpl) Errorf(msg string, args ...any) {
	tl.Error(fmt.Errorf(msg, args...).Error())
}

func (tl *TaskLoggerImpl) Errorln(msg string, args ...any) {
	tl.Errorf(msg+"\n", args...)
}

func (tl *TaskLoggerImpl) Info(msg string) {
	tl.write(msg, Info)
}

func (tl *TaskLoggerImpl) Infof(msg string, args ...any) {
	tl.Info(fmt.Sprintf(msg, args...))
}

func (tl *TaskLoggerImpl) Infoln(msg string, args ...any) {
	tl.Infof(msg+"\n", args...)
}

func (tl *TaskLoggerImpl) Debug(msg string) {
	tl.write(msg, Debug)
}

func (tl *TaskLoggerImpl) Debugf(msg string, args ...any) {
	tl.Debug(fmt.Sprintf(msg, args...))
}

func (tl *TaskLoggerImpl) Debugln(msg string, args ...any) {
	tl.Debugf(msg+"\n", args...)
}

func (tl *TaskLoggerImpl) Trace(msg string) {
	tl.write(msg, Trace)
}

func (tl *TaskLoggerImpl) Tracef(msg string, args ...any) {
	tl.Trace(fmt.Sprintf(msg, args...))
}

func (tl *TaskLoggerImpl) Traceln(msg string, args ...any) {
	tl.Tracef(fmt.Sprintf(msg+"\n", args...))
}

func (tl *TaskLoggerImpl) SetStatus(status string, args ...any) {
	tl.title.Set(fmt.Sprintf(status, args...))
}

func (tl *TaskLoggerImpl) GetTitle() string {
	return tl.title.Get()
}

func (tl *TaskLoggerImpl) IsErr() bool {
	return tl.err.Get()
}

func (tl *TaskLoggerImpl) IsComplete() bool {
	return tl.complete.Get()
}

func (tl *TaskLoggerImpl) IsHidden() bool {
	return tl.hidden.Get()
}

func (tl *TaskLoggerImpl) GetLog() ([]byte, error) {
	return ioutil.ReadFile(tl.logFilePath)
}

func (tl *TaskLoggerImpl) GetExecutionLog() []byte {
	return tl.execLog.GetAndClear()
}

func (tl *TaskLoggerImpl) Done() {
	if tl.logFilePath != "" {
		tl.File.Close()
	}
	tl.complete.Set(true)
}

func WithFormat(format func(msg string, verb VerbosityLevel) string) TaskLoggerImplOption {
	return func(tl *TaskLoggerImpl) {
		tl.format = format
	}
}

func WithHidden() TaskLoggerImplOption {
	return func(tl *TaskLoggerImpl) {
		tl.hidden.Set(true)
	}
}

func WithTaskVerbosityLevel(verb VerbosityLevel) TaskLoggerImplOption {
	return func(tl *TaskLoggerImpl) {
		tl.verbosity = verb
	}
}

func WithLogFile(path string) TaskLoggerImplOption {
	return func(tl *TaskLoggerImpl) {
		tl.logFilePath = path
	}
}
