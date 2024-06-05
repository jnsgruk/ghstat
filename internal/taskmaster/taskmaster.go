package taskmaster

import (
	"os"

	"github.com/slok/gospinner"
)

// Taskmaster handles the lifecycle of the application, and instructs the
// processing of the configured roles
type Taskmaster struct {
	tasks   []*Task
	verbose bool
	spinner *gospinner.Spinner
}

// NewTaskmaster constructs a new Manager with the specified config and
// formatter
func NewTaskmaster(verbose bool) (*Taskmaster, error) {
	var spinner *gospinner.Spinner

	if !verbose {
		spinner, _ = gospinner.NewSpinnerWithColor(gospinner.Dots, gospinner.FgGreen)
		spinner.Writer = os.Stderr
	}

	return &Taskmaster{
		spinner: spinner,
		verbose: verbose,
	}, nil
}

// AddTask is used to add tasks to the Taskmaster for future execution
func (m *Taskmaster) AddTask(task *Task) {
	task.Spinner = m.spinner
	task.Verbose = m.verbose
	m.tasks = append(m.tasks, task)
}

// Execute runs through the tasks in the Taskmaster, executing them sequentially
func (m *Taskmaster) Execute() error {
	for _, task := range m.tasks {
		if task.Status() == Ready {
			err := task.Execute()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
