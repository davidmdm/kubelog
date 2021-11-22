package terminal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

// By-pass tty checks and print with color
func init() {
	color.NoColor = false
}

var (
	stdout = colorable.NewColorableStdout()
	stderr = colorable.NewColorableStderr()
)

func Println(args ...interface{}) {
	fmt.Fprintln(stdout, args...)
}

func Print(args ...interface{}) {
	fmt.Fprint(stdout, args...)
}

func Printf(format string, args ...interface{}) {
	fmt.Fprintf(stdout, format, args...)
}

func PrintErrln(args ...interface{}) {
	fmt.Fprintln(stderr, args...)
}

func PrintErr(args ...interface{}) {
	fmt.Fprint(stderr, args...)
}

func PrintErrf(format string, args ...interface{}) {
	fmt.Fprintf(stderr, format, args...)
}

func SaveCursorPosition() {
	fmt.Fprint(stdout, "\0337")
}

func ResetCursorPosition() {
	fmt.Fprint(stdout, "\0338")
}

func ClearLineFromCursor() {
	fmt.Fprint(stdout, "\033[K")
}

func ClearDisplayFromCursor() {
	fmt.Fprint(stdout, "\033[J")
}

func ResetCursorAndClearDisplay() {
	ResetCursorPosition()
	ClearDisplayFromCursor()
}

var (
	Bold      = color.New(color.Bold).Sprintf
	Cyan      = color.New(color.FgCyan).Sprintf
	Green     = color.New(color.FgGreen).Sprintf
	BoldGreen = color.New(color.FgGreen, color.Bold).Sprintf
	Yellow    = color.New(color.FgYellow).Sprintf
)

var spinnerStates = []string{"▀ ", " ▀", " ▄", "▄ "}

func NewSpinner(ctx context.Context, roc time.Duration) (<-chan string, func()) {
	ctx, cancel := context.WithCancel(ctx)

	ticker := time.NewTicker(roc)
	spinner := make(chan string)
	done := make(chan struct{})

	go func() {
		var idx int
		spinner <- spinnerStates[0]
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				close(done)
				return
			case <-ticker.C:
				idx++
				spinner <- spinnerStates[idx%len(spinnerStates)]
			}
		}
	}()

	stop := func() {
		cancel()
		<-done
	}

	return spinner, stop
}

type progressOptions struct {
	KeepAfterStop bool
}

type ProgressOption func(*progressOptions)

func KeepMessageAfterStop(value bool) ProgressOption {
	return func(po *progressOptions) {
		po.KeepAfterStop = value
	}
}

func PrintProgressMsg(ctx context.Context, initialMsg string, options ...ProgressOption) (chan<- string, func()) {
	opts := progressOptions{}
	for _, apply := range options {
		apply(&opts)
	}

	msgChan := make(chan string)
	msg := initialMsg
	var spinnerState string

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(ctx)

	spinner, stop := NewSpinner(ctx, 500*time.Millisecond)

	printMsg := func() {
		Printf("%s %s\r\n", Cyan(spinnerState), msg)
	}

	SaveCursorPosition()

	go func() {
		defer close(done)

		printMsg()
		for {
			select {
			case <-ctx.Done():
				return
			case msg = <-msgChan:
			case spinnerState = <-spinner:
			}
			ResetCursorAndClearDisplay()
			printMsg()
		}
	}()

	var isStopped uint32

	return msgChan, func() {
		if ok := atomic.CompareAndSwapUint32(&isStopped, 0, 1); !ok {
			return
		}
		stop()
		cancel()
		<-done
		if !opts.KeepAfterStop {
			ResetCursorAndClearDisplay()
		}
	}
}

func Confirm(question string, defaultValue bool) (confirmation bool, err error) {
	SaveCursorPosition()
	defer func() {
		if err != nil {
			return
		}
		ResetCursorAndClearDisplay()

		answer := func() string {
			if confirmation {
				return "Yes"
			}
			return "No"
		}()

		fmt.Fprintf(stdout, "%s %s %s\n", Bold(Green("?")), Bold(question), Cyan(answer))
	}()

	hint := func() string {
		if defaultValue {
			return "(Y/n)"
		}
		return "(y/N)"
	}()

	for {
		fmt.Fprintf(stdout, "%s %s %s ", BoldGreen("?"), Bold(question), Bold(hint))

		line, _, err := bufio.NewReader(os.Stdin).ReadLine()
		if err != nil {
			return false, err
		}

		switch strings.ToLower(string(line)) {
		case "y":
			return true, nil
		case "n":
			return false, nil
		case "":
			return defaultValue, nil
		default:
			fmt.Fprintln(stdout, "expected answer to be 'y' or 'n'... try again")
			fmt.Fprintln(stdout)
		}
	}
}

func Select(message string, candidates []string) (answer string, err error) {
	idx, err := SelectIndex(message, candidates)
	if err != nil {
		return "", err
	}
	return candidates[idx], nil
}

func SelectIndex(message string, candidates []string) (index int, err error) {
	prompt := &survey.Select{Message: message, Options: candidates}
	err = survey.AskOne(prompt, &index)
	return
}
