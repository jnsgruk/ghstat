package taskmaster

import (
	"fmt"
	"log/slog"

	"github.com/slok/gospinner"
)

// Task represents a given taken within a Taskmaster. Each task has a name, and
// some options that enable the Taskmaster to control its output (verbose, silent).
// It also has a notion of it's progress, and a Spinner used to display progress.
type Task struct {
	Name    string
	Verbose bool

	message  string
	taskFunc func(tc *TaskCtl) error
	silent   bool
	status   Status
	progress float64

	Spinner *gospinner.Spinner
}

// Status is used to represent Task status
type Status int

const (
	// Task is ready to be started
	Ready Status = iota
	// Task is in progress
	Started
	// Task completed successfully
	Succeeded
	// Task failed
	Failed
)

// NewTask constructs a new Task for a given function
func NewTask(name, message string, taskFunc func(tc *TaskCtl) error, silent bool) *Task {
	return &Task{
		Name:     name,
		message:  message,
		taskFunc: taskFunc,
		silent:   silent,
	}
}

// Execute runs the task, performing any common preamble/cleanup
func (t *Task) Execute() error {
	t.start()

	// Create a TaskCtl so the taskFunc can report progress, cancel, etc.
	tc := &TaskCtl{task: t}

	if err := t.taskFunc(tc); err != nil {
		t.fail(err)
		return err
	}

	t.succeed()
	return nil
}

// Status reports the status of the task
func (t *Task) Status() Status {
	return t.status
}

// SetProgress updates the internal progress value, and changes the message on the spinner
func (t *Task) SetProgress(progress float64) {
	t.progress = progress
	if t.Spinner != nil {
		t.Spinner.SetMessage(fmt.Sprintf("%s (%.0f%%)", t.message, t.progress))
	}
}

// SetMessage updates the spinner message for the task
func (t *Task) SetMessage(message string) {
	t.message = message
	if t.Spinner != nil {
		if t.progress != 0 {
			t.Spinner.SetMessage(fmt.Sprintf("%s (%.0f%%)", t.message, t.progress))
		} else {
			t.Spinner.SetMessage(t.message)
		}
	}
}

// start is called at the start of task execution and takes care of logging/output
func (t *Task) start() {
	t.status = Started
	if t.Verbose {
		slog.Debug("started step", "step", t.Name)
	} else if !t.Verbose && !t.silent && t.Spinner != nil {
		t.Spinner.Start(t.message)
	}
}

// fail is called when the task fails, and used to output appropriately
func (t *Task) fail(err error) {
	t.status = Failed
	if t.Verbose {
		slog.Debug("failed step", "step", t.Name, "error", err.Error())
	} else if !t.Verbose && !t.silent && t.Spinner != nil {
		t.Spinner.SetMessage(t.message)
		t.Spinner.Fail()
	}
}

// succeed is called when the task succeeds, and used to output appropriately
func (t *Task) succeed() {
	t.status = Succeeded
	if t.Verbose {
		slog.Debug("completed step", "step", t.Name)
	} else if !t.Verbose && !t.silent && t.Spinner != nil {
		t.Spinner.SetMessage(t.message)
		t.Spinner.Succeed()
	}
}

// TaskCtl is a type that exposes certain functionality to the function executed
// by the task - enabling the function to control failure, progress display, etc.,
// but without exposing too much.
type TaskCtl struct {
	task *Task
}

// SetProgress enables the task's function to report progress back to the Task
func (tc *TaskCtl) SetProgress(progress float64) {
	tc.task.SetProgress(progress)
}

// SetMessage enables the task's function to change the spinner message
func (tc *TaskCtl) SetMessage(message string) {
	tc.task.SetMessage(message)
}
