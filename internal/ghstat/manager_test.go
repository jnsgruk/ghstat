package ghstat

import (
	"bytes"
	"fmt"
	"jnsgruk/ghstat/internal/taskmaster"
	"os"
	"testing"
)

func TestNewManagerSuccess(t *testing.T) {
	m, _, err := testManager()
	if err != nil {
		t.Fatalf("failed to construct a manager instance: %s", err.Error())
	}

	if m.taskmaster == nil {
		t.Fatalf("manager was not constructed with a valid taskmaster")
	}
}

func TestNewManagerFailure(t *testing.T) {
	config := &config{
		Leads:   []lead{},
		Verbose: false,
		Filter:  []string{},
		// This is the attribute that should cause the failure
		Formatter: "foobar",
	}

	_, err := NewManager(config, &FakeGreenhouse{}, os.Stdout)

	if err == nil {
		t.Errorf("failed to catch an invalid formatter type when constructing a manager")
	}
}

func TestManagerTasksNoRoles(t *testing.T) {
	m, _, err := testManager()
	if err != nil {
		t.Fatalf("failed to get a manager instance: %s", err.Error())
	}

	err = m.Execute()
	if err != nil {
		t.Fatalf("failed to execute manager's tasks: %s", err.Error())
	}

	expectedTasks := []string{"login", "processing", "output"}
	tasks := []string{}

	for _, task := range m.taskmaster.Tasks() {
		tasks = append(tasks, task.Name)

		if task.Status != taskmaster.Succeeded {
			t.Errorf("expected task status to be 'Succeeded', got %s", task.Status.String())
		}
	}

	if fmt.Sprint(tasks) != fmt.Sprint(expectedTasks) {
		t.Errorf("manager's tasks do not match the expected list of tasks, expected %s, got %s", fmt.Sprint(expectedTasks), fmt.Sprint(tasks))
	}
}

func TestManagerTasksNumRolesProcessed(t *testing.T) {
	m, _, _ := testManager()

	m.config.Leads = []lead{{
		Name:  "Joe Bloggs",
		Roles: []int64{123, 456, 789},
	}}

	err := m.Execute()
	if err != nil {
		t.Errorf("error executing the manager: %s", err.Error())
	}

	if len(m.roles) != 3 {
		t.Errorf("expected 3 roles, got %d", len(m.roles))
	}
}

func TestManagerTasksFormatterOutputWriter(t *testing.T) {
	m, b, _ := testManager()

	m.config.Leads = []lead{{
		Name:  "Joe Bloggs",
		Roles: []int64{123, 456, 789},
	}}

	err := m.Execute()
	if err != nil {
		t.Errorf("error executing the manager: %s", err.Error())
	}

	expectedOutput := `| Lead       | Role     | CVs | Decisions | Scheduling | WI (Screen) | WI (Grade) | Stale |
| ---------- | -------- | --- | --------- | ---------- | ----------- | ---------- | ----- |
| Joe Bloggs | Role 123 | 17  | 17        | 17         | 17          | 17         | 17    |
| Joe Bloggs | Role 456 | 17  | 17        | 17         | 17          | 17         | 17    |
| Joe Bloggs | Role 789 | 17  | 17        | 17         | 17          | 17         | 17    |
`

	if expectedOutput != b.String() {
		t.Error("formatter output did not match expected output")
	}
}

func testManager() (*Manager, *bytes.Buffer, error) {
	config := &config{
		Leads:     []lead{},
		Verbose:   true,
		Filter:    []string{},
		Formatter: "markdown",
	}

	var b bytes.Buffer
	m, err := NewManager(config, &FakeGreenhouse{}, &b)

	return m, &b, err
}

type FakeGreenhouse struct{}

func (fg *FakeGreenhouse) RoleTitle(roleId int64) (string, error) {
	return fmt.Sprintf("Role %d", roleId), nil
}

func (fg *FakeGreenhouse) CandidateCount(roleId int64, query map[string]string) (int, error) {
	return 17, nil
}

func (fg *FakeGreenhouse) Login() error {
	return nil
}
