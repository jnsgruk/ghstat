package greenhouse

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/kirsle/configdir"
)

// browser is an interface defining a set of methods that should
// be present on a browser struct for it to be compatible with ghstat
type Browser interface {
	Init() error
	LoadCookies() error
	SaveCookies() error
	Browser() *rod.Browser
}

// ghstatBrowser represents a ghstatBrowser and it's state in ghstat
type ghstatBrowser struct {
	browser *rod.Browser
}

// Browser returns a pointer to the underlying rod.Browser instance
func (b *ghstatBrowser) Browser() *rod.Browser {
	return b.browser
}

// Init ensures that the application can launch the configured browser, and
// attempts to login in Greenhouse, first by using saved cookies in a known
// location, and secondly by prompting for login, password and OTP.
func (b *ghstatBrowser) Init() error {
	// Get the path to the user's browser using a list of predefined browser bin names
	path, err := findBrowser()
	if err != nil {
		return err
	}

	l := launcher.New().Bin(path)

	// If we're inside a snap, use the `--no-sandbox` flag.
	if len(os.Getenv("SNAP")) > 0 {
		l.NoSandbox(true)
	}

	// Launch the browser and ensure that we can connect to the dev tools port.
	u, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	b.browser = rod.New().ControlURL(u)

	err = b.browser.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to browser control url: %w", err)
	}
	slog.Debug("connected to browser control url")
	return nil
}

// loadCookies attempts to load cookies from a previous ghstat session
// from the users config directory
func (b *ghstatBrowser) LoadCookies() error {
	cookies := []*proto.NetworkCookieParam{}

	configPath := configdir.LocalConfig("ghstat")
	configFile := filepath.Join(configPath, "ghstat.json")

	buf, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to open cookie store file: %w", err)
	}

	err = json.Unmarshal(buf, &cookies)
	if err != nil {
		return fmt.Errorf("failed to parse cookie store file: %w", err)
	}

	err = b.browser.SetCookies(cookies)
	if err != nil {
		return fmt.Errorf("failed to load cookies into browser: %w", err)
	}

	return nil
}

// saveCookies dumps all the cookies from the browser's current session
// into a file in the users config directory
func (b *ghstatBrowser) SaveCookies() error {
	// Save the cookies to the cookies file
	cookies, _ := b.browser.GetCookies()
	buf, err := json.MarshalIndent(cookies, "", "  ")
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
	f.Write(buf)

	return nil
}

// findBrowser is a helper utility to get the path of a browser that ghstat
// can use for gathering the information it requires.
func findBrowser() (string, error) {
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
