package out

import (
	"io"
	"os"
	"sync"
	"time"

	atomics "github.com/tiagoposse/go-sync-types"

	"atomicgo.dev/cursor"
	"github.com/fatih/color"

	"golang.org/x/crypto/ssh/terminal"
)

var characterSet = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

var termWidth int
var termHeight int

type OutputManager struct {
	out                 io.Writer
	outstandingLog      *atomics.Slice[string]
	tasks               *atomics.OrderedMap[string, TaskLogger]
	refreshInterval     time.Duration
	currentLineCount    int
	verbosity           VerbosityLevel
	logsRoot            string
	keepOutput          bool
	keepTasksOnComplete bool
	time                string
	tty                 bool
	spinner             int
	maxOutputPerTask    int
	running             sync.WaitGroup
	closeOut            chan struct{}
	colorSchemes        map[VerbosityLevel]func(format string, a ...interface{}) string
}

type OutputManagerOption func(tl *OutputManager)

func NewOutputManager(opts ...OutputManagerOption) (*OutputManager, error) {
	now := time.Now()
	mgr := &OutputManager{
		tasks:            atomics.NewOrderedMap[string, TaskLogger](),
		time:             now.Format("2006_01_02_15_04_05"),
		tty:              true,
		refreshInterval:  100,
		closeOut:         make(chan struct{}),
		maxOutputPerTask: 20,
		logsRoot:         "",
		out:              os.Stdout,
		outstandingLog:   atomics.NewSlice[string](),
		verbosity:        Info,
		colorSchemes: map[VerbosityLevel]func(format string, a ...interface{}) string{
			Error:   color.New(color.FgRed).SprintfFunc(),
			Trace:   color.New(color.FgCyan).SprintfFunc(),
			Warning: color.New(color.FgYellow).SprintfFunc(),
			Debug:   color.New(color.FgGreen).SprintfFunc(),
			Info:    color.New(color.FgWhite).SprintfFunc(),
		},
	}

	// fmt.Printf("Setting max verbosity to %d\n", mgr.verbosity)
	for _, opt := range opts {
		opt(mgr)
	}

	// err := os.MkdirAll(mgr.executionLogsRoot, os.ModePerm)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed creating log root for output: %w", err)
	// }

	termWidth, termHeight, _ = terminal.GetSize(0)

	return mgr, nil
}

func (om *OutputManager) Verbosity() VerbosityLevel {
	return om.verbosity
}

func (com *OutputManager) AddTask(id string, logger TaskLogger) error {
	// var err error
	// if err = os.MkdirAll(filepath.Join(com.executionLogsRoot, strings.ReplaceAll(id, ":", string(os.PathSeparator))), os.ModePerm); err != nil {
	// 	return err
	// }

	com.tasks.Put(id, logger)

	return nil
}

func (com *OutputManager) CreateTask(id, title string, opts ...TaskLoggerImplOption) (*TaskLoggerImpl, error) {
	baseOpts := []TaskLoggerImplOption{
		WithFormat(func(msg string, verb VerbosityLevel) string {
			return com.colorSchemes[verb](msg)
		}),
		WithTaskVerbosityLevel(com.verbosity),
	}

	task := NewTaskLogger(title, append(baseOpts, opts...)...)
	return task, com.AddTask(id, task)
}

func (com *OutputManager) CompleteTask(id string) {
	task, _ := com.tasks.Get(id)
	task.Done()
	com.RemoveTask(id)
}

func (com *OutputManager) RemoveTask(id string) {
	task, ok := com.tasks.Get(id)
	if !ok {
		return
	}

	if com.keepTasksOnComplete && com.keepOutput {
		var char string
		if task.IsErr() {
			fn := color.New(color.FgRed).SprintfFunc()
			char = fn("✘")
		} else if task.IsComplete() {
			fn := color.New(color.FgGreen).SprintfFunc()
			char = fn("✔")
		}

		com.outstandingLog.Append(char + " " + task.GetTitle() + "\n")
	}

	com.tasks.Remove(id)
}

func (com *OutputManager) Stop() {
	close(com.closeOut)
	com.running.Wait()
}

func (com *OutputManager) Start() {
	com.running.Add(1)
	go func() {
		for {
			select {
			// Wait for stop signal
			case <-com.closeOut:
				com._render(true)
				com.running.Done()

				return
			default:
				com._render(false)
			}

			time.Sleep(com.refreshInterval * time.Millisecond)
		}
	}()
}

func (com *OutputManager) _render(last bool) {
	if com.tasks.Length() == 0 {
		return
	}
	currentLineCount := 0
	var maxOutPerTask int
	if termHeight > 0 {
		maxOutPerTask = (termHeight - 1) / com.tasks.Length()
	} else {
		maxOutPerTask = 1
	}
	if maxOutPerTask > com.maxOutputPerTask {
		maxOutPerTask = com.maxOutputPerTask
	}

	fullTaskText := []byte{}
	com.tasks.Iterate(func(id string, task TaskLogger) {
		if com.tty && !task.IsHidden() {
			var char string
			if task.IsErr() {
				fn := color.New(color.FgRed).SprintfFunc()
				char = fn("✘")
			} else if task.IsComplete() {
				fn := color.New(color.FgGreen).SprintfFunc()
				char = fn("✔")
			} else {
				char = characterSet[com.spinner]
			}

			fullTaskText = append(fullTaskText, []byte(char+" "+task.GetTitle()+"\n")...)
			currentLineCount++
			if !last || (last && (task.IsErr() || com.keepOutput)) {
				taskText, numLines := task.GetLastNLines(maxOutPerTask-1, termWidth-2)
				// taskText = append(taskText, []byte(fmt.Sprintf("Num lines: %d: %v\n", numLines, taskText))...)
				// currentLineCount++
				currentLineCount += numLines
				fullTaskText = append(fullTaskText, taskText...)
			}
		} else {
			taskText := task.GetTextAndClear()
			if taskText == "" {
				return
			}

			com.outstandingLog.Append(taskText)
		}
	})

	if com.currentLineCount > 0 && com.tty {
		cursor.ClearLinesUp(com.currentLineCount)
	}

	for _, line := range com.outstandingLog.GetAndClear() {
		com.out.Write([]byte(line))
	}

	com.out.Write(fullTaskText)

	com.spinner++
	if com.spinner == 6 {
		com.spinner = 0
	}

	com.currentLineCount = currentLineCount
}
