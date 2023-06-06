package out

type TaskLogger interface {
	Warn(msg string)
	Warnln(msg string, args ...any)
	Error(msg string)
	Errorf(msg string, args ...any)
	Info(msg string)
	Infoln(msg string, args ...any)
	Debug(msg string)
	Debugln(msg string, args ...any)
	Trace(msg string)
	Traceln(msg string, args ...any)
	SetStatus(status string, args ...any)
	GetText() string
	GetTitle() string
	GetTextAndClear() string
	GetLastNLines(numLines, charLimit int) ([]byte, int)
	IsErr() bool
	IsComplete() bool
	IsHidden() bool
	Done()
}
