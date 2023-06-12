package out

type TaskLogger interface {
	Warn(msg string)
	Warnln(msg string, args ...any)
	Error(msg string)
	Errorf(msg string, args ...any)
	Errorln(msg string, args ...any)
	Info(msg string)
	Infof(msg string, args ...any)
	Infoln(msg string, args ...any)
	Debug(msg string)
	Debugf(msg string, args ...any)
	Debugln(msg string, args ...any)
	Trace(msg string)
	Tracef(msg string, args ...any)
	Traceln(msg string, args ...any)
	SetStatus(status string, args ...any)
	Write(b []byte) (int, error)
	GetTitle() string
	GetLog() ([]byte, error)
	GetExecutionLog() []byte
	IsErr() bool
	IsComplete() bool
	IsHidden() bool
	Done()
}
