package taskmaster

import (
	"fmt"
	"testing"
	"time"
)

// TestTaskNewTask ensures the new tasks are constructed with the correct properties
func TestTaskNewTask(t *testing.T) {
	task := NewTask("foo", "foobar", successWorker(), false)

	if task.Name != "foo" {
		t.Errorf("failed to construct a task with the correct name")
	}

	if task.Status() != Ready {
		t.Errorf("invalid task status upon construction: %v", task.Status())
	}
}

// TestTaskExecuteSuccess tests a task with a simple worker that always succeeds,
// and ensures that the status is reported correctly
func TestTaskExecuteSuccess(t *testing.T) {
	task := NewTask("foo", "foobar", successWorker(), false)

	err := task.Execute()
	if err != nil {
		t.Errorf("task should have succeeded, but got error: %s", err.Error())
	}

	if task.Status() != Succeeded {
		t.Errorf("task status should have been set to 'Succeeded', but got: %d", task.Status())
	}
}

// TestTaskExecuteWithProgress ensures that messages/progress is updated correctly
// when a task interacts with TaskCtl
func TestTaskExecuteWithProgress(t *testing.T) {
	task := NewTask("foo", "foobar", progressWorker(1*time.Second), false)

	c := make(chan struct{})

	// Execute the task in a goroutine
	go func() {
		err := task.Execute()
		if err != nil {
			t.Errorf("task should have succeeded, but got error: %s", err.Error())
		}
		c <- struct{}{}
	}()

	time.Sleep(500 * time.Millisecond)
	if task.Status() != Started {
		t.Errorf("task status should have been set to 'Started', but got: %d", task.Status())
	}

	if task.progress != 50 {
		t.Errorf("task progress should be 50, but got: %f", task.progress)
	}

	if task.message != "Half Way" {
		t.Errorf("task message should be 'Half Way', but got: %s", task.message)
	}
	<-c
}

// TestTaskExecuteFail ensures that failing tasks have their status set correctly
func TestTaskExecuteFail(t *testing.T) {
	task := NewTask("foo", "foobar", failWorker("foo"), false)

	err := task.Execute()
	if err == nil {
		t.Errorf("task should have failed")
	}

	if task.Status() != Failed {
		t.Errorf("task status should have been set to 'Failed', but got: %d", task.Status())
	}
}

// progressWorker is a test fixture - a task func that sets progress and a message
func progressWorker(d ...time.Duration) func(tc *TaskCtl) error {
	return func(tc *TaskCtl) error {
		tc.SetProgress(50)
		tc.SetMessage("Half Way")
		for _, v := range d {
			time.Sleep(v)
		}
		return nil
	}
}

// successWorker is a simple task func that always succeeds
func successWorker(d ...time.Duration) func(tc *TaskCtl) error {
	return func(tc *TaskCtl) error {
		for _, v := range d {
			time.Sleep(v)
		}
		return nil
	}
}

// failWorker is a simple task func that always fails
func failWorker(n string) func(tc *TaskCtl) error {
	return func(tc *TaskCtl) error {
		return fmt.Errorf("worker failed: %s", n)
	}
}
