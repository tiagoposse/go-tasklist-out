package out

type VerbosityLevel int

func (c VerbosityLevel) String() string {
	return string(c)
}

const (
	Error   VerbosityLevel = 1
	Warning VerbosityLevel = 2
	Info    VerbosityLevel = 3
	Debug   VerbosityLevel = 4
	Trace   VerbosityLevel = 5
)
