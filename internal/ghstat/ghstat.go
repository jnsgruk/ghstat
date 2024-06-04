package ghstat

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/kirsle/configdir"
	"github.com/manifoldco/promptui"
)

// Manager handles the lifecycle of the application, and instructs the
// processing of the configured roles
type Manager struct {
	config    *Config
	browser   *rod.Browser
	formatter Formatter
	roles     []Role
}

// NewManager constructs a new Manager with the specified config and
// formatter
func NewManager(config *Config, formatter Formatter) *Manager {
	return &Manager{
		config:    config,
		formatter: formatter,
	}
}

// Init ensures that the application can launch the configured browser, and
// attempts to login in Greenhouse, first by using saved cookies in a known
// location, and secondly by prompting for login, password and OTP.
func (m *Manager) Init() error {
	// Get the path to the user's browser using a list of predefined browser bin names
	path, err := m.findBrowser()
	if err != nil {
		return err
	}

	b := launcher.New().Bin(path).Set("blink-settings", "imagesEnabled=false")

	// If we're inside a snap, use the `--no-sandbox` flag.
	if len(os.Getenv("SNAP")) > 0 {
		b.NoSandbox(true)
	}

	// Launch the browser and ensure that we can connect to the dev tools port.
	u, err := b.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	m.browser = rod.New().ControlURL(u)
	slog.Debug("got browser control url", "url", u)

	err = m.browser.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to browser control url: %w", err)
	}
	slog.Debug("connected to browser control url")

	// Load cookies if available, short-circuiting the need to login again
	savedCookies, err := m.loadCookies()
	if err == nil {
		m.browser.SetCookies(savedCookies)
	} else {
		slog.Debug(err.Error())
	}

	// Ensure the app is logged in to Greenhouse
	err = m.checkLoggedIn()
	if err != nil {
		return fmt.Errorf("failed to login to Greenhouse: %w", err)
	}

	return nil
}

// Process iterates over the configured roles and gathers statistics about them
func (m *Manager) Process(filter []string) {
	wg := sync.WaitGroup{}
	for _, lead := range m.filterLeads(filter) {
		lead := lead
		for _, roleId := range lead.Roles {
			roleId := roleId
			wg.Add(1)
			go func() {
				defer wg.Done()
				slog.Debug("processing role", "roleId", roleId, "lead", lead.Name)
				g := Greenhouse{browser: m.browser}

				m.roles = append(m.roles, Role{
					RawID:              roleId,
					RawName:            g.RoleTitle(roleId),
					RawLead:            lead.Name,
					RawAppReviews:      g.AppReviews(roleId),
					RawNeedsDecision:   g.OutstandingDecisions(roleId),
					RawNeedsScheduling: g.OutstandingScheduling(roleId),
					RawWIScreening:     g.OutstandingWIScreening(roleId),
					RawWIGrading:       g.OutstandingWIGrading(roleId),
					RawStale:           g.StaleCandidates(roleId),
				})
			}()
		}
	}
	wg.Wait()

	// Sort the roles in descending order by number of outstanding app reviews
	slices.SortFunc(m.roles, func(a, b Role) int {
		return cmp.Compare(b.RawAppReviews, a.RawAppReviews)
	})

	// Save the cookies from the current session for the next time ghstat is used
	err := m.saveCookies(m.browser)
	if err != nil {
		// Don't bother doing much with this, it's just a convenience feature and
		// not critical to the operation such that the command should fail
		slog.Debug(err.Error())
	}
}

// Output uses the selected formatter to print the results to the terminal
func (m *Manager) Output() {
	if len(m.roles) > 0 {
		m.formatter.Output(m.roles)
	}
}

// filterLeads takes a list of repo names and returns a list of only those Leads
// from the manager's config.
func (m *Manager) filterLeads(filter []string) []Lead {
	leads := m.config.Leads
	if len(filter) > 0 {
		filteredLeads := []Lead{}
		for _, lead := range leads {
			if slices.Contains(filter, lead.Name) {
				filteredLeads = append(filteredLeads, lead)
			}
		}
		leads = filteredLeads
	}
	return leads
}

// checkLoggedIn checks if the app is logged into Greenhouse from the cookies
// created before, and if not walks the user through the login flow by prompting
// for their username, password and OTP
func (m *Manager) checkLoggedIn() error {
	page, err := m.browser.Page(proto.TargetCreateTarget{URL: "https://canonical.greenhouse.io"})
	if err != nil {
		return fmt.Errorf("failed to open url 'https://canonical.greenhouse.io': %w", err)
	}

	// Wait for the page to settle
	err = page.WaitStable(300 * time.Millisecond)
	if err != nil {
		return fmt.Errorf("failed to check login status: %w", err)
	}

	// If redirected to the Ubuntu One login, handle the login correctly
	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("failed to retrieve page information: %w", err)
	}

	if info.URL == "https://login.ubuntu.com/+login?next=%2Fsaml%2Fprocess" {
		var (
			login    string
			password string
			err      error
		)

		login = os.Getenv("U1_LOGIN")
		if len(login) == 0 {
			prompt := promptui.Prompt{Label: "Ubuntu One Login"}
			login, err = prompt.Run()
			if err != nil {
				return fmt.Errorf("failed to read login: %w", err)
			}
		}

		password = os.Getenv("U1_PASSWORD")
		if len(password) == 0 {
			prompt := promptui.Prompt{
				Label: "Ubuntu One Password",
				Mask:  '*',
			}
			password, err = prompt.Run()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		}

		page.MustElement(`#id_email`).MustInput(login)
		page.MustElement(`#id_password`).MustInput(password)
		page.MustElement("[type=submit]").MustClick()

		err = page.WaitStable(300 * time.Millisecond)
		if err != nil {
			return fmt.Errorf("failed to check login status: %w", err)
		}

		prompt := promptui.Prompt{Label: "Ubuntu One OTP"}
		otp, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("failed to read otp: %w", err)
		}

		page.MustElement(`#id_oath_token`).MustInput(otp)
		page.MustElement("[type=submit]").MustClick()

		err = page.WaitStable(300 * time.Millisecond)
		if err != nil {
			return fmt.Errorf("failed to check login status: %w", err)
		}
	}

	return nil
}

// loadCookies attempts to load cookies from a previous ghstat session
// from the users config directory
func (m *Manager) loadCookies() ([]*proto.NetworkCookieParam, error) {
	cookies := []*proto.NetworkCookieParam{}

	configPath := configdir.LocalConfig("ghstat")
	configFile := filepath.Join(configPath, "ghstat.json")

	b, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open cookie store file: %w", err)
	}

	err = json.Unmarshal(b, &cookies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cookie store file: %w", err)
	}

	return cookies, nil
}

// saveCookies dumps all the cookies from the browser's current session
// into a file in the users config directory
func (m *Manager) saveCookies(browser *rod.Browser) error {
	// Save the cookies to the cookies file
	cookies, _ := browser.GetCookies()
	b, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal cookie data: %w", err)
	}

	configPath := configdir.LocalConfig("ghstat")
	err = configdir.MakePath(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Deal with a JSON configuration file in that folder.
	configFile := filepath.Join(configPath, "ghstat.json")

	f, err := os.Create(configFile)
	if err != nil {
		return fmt.Errorf("could create cookie file: %w", err)
	}

	defer f.Close()
	f.Write(b)

	return nil
}

// findBrowser is a helper utility to get the path of a browser that ghstat
// can use for gathering the information it requires.
func (m *Manager) findBrowser() (string, error) {
	browserCmds := []string{
		"google-chrome-stable",
		"google-chrome",
		"chromium-browser",
		"chromium",
		"chromium.launcher",
		"firefox",
	}

	for _, b := range browserCmds {
		p, _ := exec.LookPath(b)
		if len(p) > 0 {
			return p, nil
		}
	}

	return "", fmt.Errorf("could not find suitable browser in $PATH")
}
