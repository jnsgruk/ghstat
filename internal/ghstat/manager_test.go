package ghstat

import (
	"bytes"
	"fmt"
	"jnsgruk/ghstat/internal/taskmaster"
	"os"
	"path"
	"testing"

	"github.com/go-rod/rod"
)

func TestNewManagerSuccess(t *testing.T) {
	m, _, tmpDir, err := testManager(t)
	defer os.RemoveAll(tmpDir)
	if err != nil {
		t.Errorf("failed to construct a manager instance")
	}

	if m.taskmaster == nil {
		t.Error("manager was not constructed with a valid taskmaster")
	}
}

func TestNewManagerFailure(t *testing.T) {
	_, err := NewManager(&config{
		Leads:   []lead{},
		Verbose: false,
		Filter:  []string{},
		// This is the attribute that should cause the failure
		Formatter: "foobar",
	}, os.Stdout)

	if err == nil {
		t.Errorf("failed to catch an invalid formatter type when constructing a manager")
	}
}

func TestManagerTasksNoRoles(t *testing.T) {
	m, _, tmpDir, _ := testManager(t)
	defer os.RemoveAll(tmpDir)

	m.Execute()

	expectedTasks := []string{"init", "login", "processing", "save-state", "output"}
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
	m, _, tmpDir, _ := testManager(t)
	defer os.RemoveAll(tmpDir)

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
	m, b, tmpDir, _ := testManager(t)
	defer os.RemoveAll(tmpDir)

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

func TestManagerTasksBrowserSavesState(t *testing.T) {
	m, _, tmpDir, _ := testManager(t)
	defer os.RemoveAll(tmpDir)

	m.config.Leads = []lead{{
		Name:  "Joe Bloggs",
		Roles: []int64{123, 456, 789},
	}}

	err := m.Execute()
	if err != nil {
		t.Errorf("error executing the manager: %s", err.Error())
	}

	_, err = os.Stat(path.Join(tmpDir, "cookies.json"))
	if err != nil {
		t.Errorf("state was not correctly saved after manager execution")
	}
}

func testManager(t *testing.T) (*Manager, *bytes.Buffer, string, error) {
	var b bytes.Buffer

	m, err := NewManager(&config{
		Leads:     []lead{},
		Verbose:   true,
		Filter:    []string{},
		Formatter: "markdown",
	}, &b)

	tmpDir := t.TempDir()
	m.greenhouse = &FakeGreenhouse{}
	m.browser = &FakeBrowser{tmpDir: tmpDir}

	return m, &b, tmpDir, err
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

type FakeBrowser struct {
	tmpDir string
}

func (fb *FakeBrowser) Init() error {
	return nil
}

func (fb *FakeBrowser) LoadCookies() error {
	return nil
}

func (fb *FakeBrowser) SaveCookies() error {
	os.Create(path.Join(fb.tmpDir, "cookies.json"))
	return nil
}

func (fb *FakeBrowser) Browser() *rod.Browser {
	return &rod.Browser{}
}
