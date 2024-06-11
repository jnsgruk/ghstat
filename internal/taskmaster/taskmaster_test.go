package taskmaster

import (
	"os"
	"slices"
	"testing"
)

// TestNewTaskmaster ensures that Taskmasters are created with the correct properties
func TestNewTaskmaster(t *testing.T) {
	tm, err := NewTaskmaster(false)
	if err != nil {
		t.Error("failed to construct a new taskmaster")
	}

	if tm.spinner == nil {
		t.Error("taskmaster's spinner is nil")
	}

	if tm.spinner.Writer != os.Stderr {
		t.Error("taskmaster's spinner is not outputting to stderr")
	}

}

// TestVerboseTaskmaster tests that no spinner is created when the
// taskmaster is asked to be verbose
func TestNewVerboseTaskmaster(t *testing.T) {
	tm, err := NewTaskmaster(true)
	if err != nil {
		t.Error("failed to construct a new taskmaster")
	}

	if tm.spinner != nil {
		t.Error("verbose taskmaster's spinner should be nil")
	}
}

// TestAddTask tests that when tasks are added to the taskmaster,
// their spinner and verbose status is propagated correctly
func TestAddTask(t *testing.T) {
	tm, err := NewTaskmaster(false)
	if err != nil {
		t.Error("failed to construct a new taskmaster")
	}

	task := NewTask("foo", "foobar", successWorker(), false)
	tm.AddTask(task)

	if !slices.Contains(tm.tasks, task) {
		t.Error("failed to add task to taskmaster")
	}

	if task.Spinner != tm.spinner {
		t.Error("task did not have it's spinner set when added to taskmaster")
	}

	if task.Verbose {
		t.Error("task did not have it's verbose property set when added to taskmaster")
	}
}

// TestExecuteTasksSuccess tests a clean run where two tasks are added
// and executed successfully
func TestExecuteTasksSuccess(t *testing.T) {
	tm, err := NewTaskmaster(false)
	if err != nil {
		t.Error("failed to construct a new taskmaster")
	}

	tm.AddTask(NewTask("foo", "foobar", successWorker(), true))
	tm.AddTask(NewTask("bar", "barbaz", successWorker(), true))

	err = tm.Execute()
	if err != nil {
		t.Error("taskmaster execution returned an error")
	}

	for _, task := range tm.tasks {
		if task.Status() != Succeeded {
			t.Errorf("task not marked as succeeded: %s", task.Name)
		}
	}
}

// TestExecuteTasksFailure tries to execute two tasks, where the first fails
func TestExecuteTasksFailure(t *testing.T) {
	tm, err := NewTaskmaster(false)
	if err != nil {
		t.Error("failed to construct a new taskmaster")
	}

	tm.AddTask(NewTask("foo", "foobar", failWorker("foo"), true))
	tm.AddTask(NewTask("bar", "barbaz", successWorker(), true))

	err = tm.Execute()
	if err == nil {
		t.Error("taskmaster execution should have failed")
	}

	if tm.tasks[0].status != Failed || tm.tasks[1].status != Ready {
		t.Error("tasks have inconsistent statuses")
	}
}

// TestExecuteTasksFailureRetry tests running two tasks, where the first fails.
// After the failure, the Taskmaster is executed again to run the next task
// to completion
func TestExecuteTasksFailureRetry(t *testing.T) {
	tm, err := NewTaskmaster(false)
	if err != nil {
		t.Error("failed to construct a new taskmaster")
	}

	tm.AddTask(NewTask("foo", "foobar", failWorker("foo"), true))
	tm.AddTask(NewTask("bar", "barbaz", successWorker(), true))

	err = tm.Execute()
	if err == nil {
		t.Error("taskmaster execution should have failed")
	}

	err = tm.Execute()
	if err != nil {
		t.Error("taskmaster execution should have failed")
	}

	if tm.tasks[0].status != Failed || tm.tasks[1].status != Succeeded {
		t.Error("tasks have inconsistent statuses")
	}
}
