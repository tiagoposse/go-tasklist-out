package out

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"atomicgo.dev/cursor"
	atomics "github.com/tiagoposse/go-sync-types"

	"github.com/fatih/color"

	"golang.org/x/crypto/ssh/terminal"
)

var characterSet = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

var termWidth int
var termHeight int

type OutputManager struct {
	out                 io.Writer
	outstandingLog      *atomics.Slice[byte]
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
	running             sync.WaitGroup
	closeOut            chan struct{}
	colorSchemes        map[VerbosityLevel]func(format string, a ...interface{}) string
}

type OutputManagerOption func(tl *OutputManager)

func NewOutputManager(opts ...OutputManagerOption) (*OutputManager, error) {
	now := time.Now()
	mgr := &OutputManager{
		tasks:           atomics.NewOrderedMap[string, TaskLogger](),
		time:            now.Format("2006_01_02_15_04_05"),
		tty:             true,
		refreshInterval: 100,
		closeOut:        make(chan struct{}),
		logsRoot:        "",
		out:             os.Stdout,
		outstandingLog:  atomics.NewSlice[byte](),
		verbosity:       Info,
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

	if mgr.logsRoot != "" {
		if err := os.MkdirAll(mgr.logsRoot, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed creating log root for output: %w", err)
		}
	}

	termWidth, termHeight, _ = terminal.GetSize(0)

	return mgr, nil
}

func (om *OutputManager) Verbosity() VerbosityLevel {
	return om.verbosity
}

func (com *OutputManager) AddTask(id string, logger TaskLogger) error {
	com.tasks.Put(id, logger)

	return nil
}

func (com *OutputManager) CreateTask(id, title string, opts ...TaskLoggerImplOption) (*TaskLoggerImpl, error) {
	baseOpts := []TaskLoggerImplOption{
		WithFormat(func(msg string, verb VerbosityLevel) string {
			return com.colorSchemes[verb](id + ": " + msg)
		}),
		WithTaskVerbosityLevel(com.verbosity),
	}

	if com.logsRoot != "" {
		baseOpts = append(baseOpts, WithLogFile(filepath.Join(com.logsRoot, strings.ReplaceAll(id, "/", "_"), com.time+".log")))
	}

	task, err := NewTaskLogger(title, append(baseOpts, opts...)...)
	if err != nil {
		return nil, err
	}
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

		com.outstandingLog.Append([]byte(char + " " + task.GetTitle() + "\n")...)
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
		}

		com.outstandingLog.Append(task.GetExecutionLog()...)
	})

	if com.currentLineCount > 0 && com.tty {
		cursor.ClearLinesUp(com.currentLineCount)
	}

	com.out.Write(com.outstandingLog.GetAndClear())
	com.out.Write(fullTaskText)

	if com.currentLineCount > currentLineCount {
		for i := 0; i < (com.currentLineCount - currentLineCount); i++ {
			fmt.Println("")
		}
		cursor.ClearLinesUp(com.currentLineCount - currentLineCount)
	}

	com.spinner++
	if com.spinner == 6 {
		com.spinner = 0
	}

	com.currentLineCount = currentLineCount
}
