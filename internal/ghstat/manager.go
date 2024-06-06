package ghstat

import (
	"cmp"
	"fmt"
	"log/slog"
	"sync/atomic"

	"jnsgruk/ghstat/internal/formatters"
	"jnsgruk/ghstat/internal/greenhouse"
	"jnsgruk/ghstat/internal/taskmaster"
	"slices"
	"sync"
)

// Manager is used for controlling the execution of the task workflow
type Manager struct {
	taskmaster *taskmaster.Taskmaster
	roles      []*greenhouse.Role
	browser    *browser
	config     *config
	formatter  formatters.Formatter
}

// NewManager constructs a new Manager, ensuring that a valid formatter has been chosen,
// and ensures it has an associated Taskmaster instance
func NewManager(c *config) (*Manager, error) {
	m := &Manager{config: c}

	m.formatter = formatters.NewFormatter(c.Formatter)
	if m.formatter == nil {
		return nil, fmt.Errorf("invalid output formatter specified, please choose one of 'pretty', 'markdown' or 'json'")
	}

	mgr, err := taskmaster.NewTaskmaster(m.config.Verbose)
	if err != nil {
		return nil, fmt.Errorf("couldn't create taskmaster: %w", err)
	}

	m.taskmaster = mgr

	return m, nil
}

// Execute is the main entrypoint into the ghstat manager
func (m *Manager) Execute() error {
	m.taskmaster.AddTask(taskmaster.NewTask("init", "Initialising", m.init, false))
	m.taskmaster.AddTask(taskmaster.NewTask("login", "Logging in", m.login, false))
	m.taskmaster.AddTask(taskmaster.NewTask("processing", "Processing roles", m.process, false))
	m.taskmaster.AddTask(taskmaster.NewTask("save-state", "Saving state", m.saveState, false))
	m.taskmaster.AddTask(taskmaster.NewTask("output", "Output", m.output, true))

	return m.taskmaster.Execute()
}

// init takes care of finding and starting a browser instance, and loading
// cookies from any previous ghstat sessions
func (m *Manager) init(tc *taskmaster.TaskCtl) error {
	browser := &browser{}
	err := browser.Init()
	if err != nil {
		return err
	}

	err = browser.LoadCookies()
	if err != nil {
		slog.Debug("failed to load cookies", "error", err.Error())
	}

	m.browser = browser
	return nil
}

// login checks if the app is logged into Greenhouse from the cookies
// created before, and if not walks the user through the checkLoggedIn flow by prompting
// for their username, password and OTP
func (m *Manager) login(tc *taskmaster.TaskCtl) error {
	g := greenhouse.NewGreenhouse(m.browser.RodBrowser)
	err := g.Login()
	if err != nil {
		return fmt.Errorf("failed to login to Greenhouse: %w", err)
	}
	return nil
}

// saveState saves any cookies from the current browser session to
// the user's config directory
func (m *Manager) saveState(tc *taskmaster.TaskCtl) error {
	// Save the cookies from the current session for the next time ghstat is used
	err := m.browser.SaveCookies()
	if err != nil {
		slog.Debug("failed to save cookies cookies", "error", err.Error())
	}
	return nil
}

// process iterates over the configured roles and gathers statistics about them
func (m *Manager) process(tc *taskmaster.TaskCtl) error {
	// Filter the leads where a filter was specified
	if len(m.config.Filter) > 0 {
		m.config.Leads = slices.DeleteFunc(m.config.Leads, func(l lead) bool {
			return !slices.Contains(m.config.Filter, l.Name)
		})
	}

	// Iterate over the list of leads/roles and construct new Role's for them
	for _, lead := range m.config.Leads {
		for _, roleId := range lead.Roles {
			role := greenhouse.NewRole(roleId, lead.Name)
			m.roles = append(m.roles, role)
		}
	}

	// Update the spinner message to include the number of roles to process
	tc.SetMessage(fmt.Sprintf("Processing %d roles", len(m.roles)))

	// Calculate the number of fields that need fetching from Greenhouse
	totalFields := len(m.roles) * greenhouse.NumRoleFields
	var fetchedFields atomic.Int64

	// Helper method so that individual Role populate funcs can report back
	// their progress
	incProgress := func(amount int64) {
		fetchedFields.Add(amount)
		tc.SetProgress(float64(fetchedFields.Load()) / float64(totalFields) * 100)
	}

	g := greenhouse.NewGreenhouse(m.browser.RodBrowser)

	// Iterate over the roles, process each in its own goroutine
	wg := sync.WaitGroup{}

	for _, r := range m.roles {
		r := r
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Populate(g, incProgress)
		}()
	}

	wg.Wait()

	return nil
}

// output uses the selected formatter to print the results to the terminal
func (m *Manager) output(tc *taskmaster.TaskCtl) error {
	// Sort the roles in ascending order by lead, then descending by number
	// of outstanding app reviews
	slices.SortFunc(m.roles, func(a, b *greenhouse.Role) int {
		return cmp.Or(
			cmp.Compare(a.Lead, b.Lead),
			cmp.Compare(b.AppReviews(), a.AppReviews()),
		)
	})

	if len(m.roles) > 0 {
		m.formatter.Output(m.roles)
	}
	return nil
}
