package greenhouse

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/manifoldco/promptui"
)

// Greenhouse is an internal representation of an instance of Greenhouse
type Greenhouse struct {
	browser *rod.Browser
}

func NewGreenhouse(browser *rod.Browser) *Greenhouse {
	return &Greenhouse{browser: browser}
}

// CandidateCount is a helper method for requesting Greenhouse candidate pages with
// a specified set of query parameters in the URL
func (g *Greenhouse) CandidateCount(roleId int64, queries map[string]string) (int, error) {
	page, err := g.getCandidatesPage(roleId, queries)
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve candidate page: %w", err)
	}
	defer page.MustClose()

	// If this element is present, the number of results is zero
	_, err = page.Timeout(500 * time.Millisecond).Element(".no_results--header")
	if err == nil {
		return 0, nil
	}

	rc, err := page.Timeout(500 * time.Millisecond).Element("#results_count")
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve candidate count: %w", err)
	}

	rcStr, err := rc.Text()
	if err != nil {
		return -1, fmt.Errorf("failed fetch candidate count: %w", err)
	}

	count, err := strconv.Atoi(rcStr)
	if err != nil {
		return -1, fmt.Errorf("failed to parse count as integer: %w", err)
	}

	return count, nil
}

// Login is used to login to Greenhouse through the Ubuntu One SSO page
func (g *Greenhouse) Login() error {
	page, err := g.browser.Page(proto.TargetCreateTarget{URL: "https://canonical.greenhouse.io"})
	if err != nil {
		return fmt.Errorf("failed to open url 'https://canonical.greenhouse.io': %w", err)
	}
	defer page.MustClose()

	// Wait for the page to settle
	err = page.WaitStable(300 * time.Millisecond)
	if err != nil {
		return fmt.Errorf("failed to check login status: %w", err)
	}

	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("failed to retrieve page information: %w", err)
	}

	// If redirected to the Ubuntu One login, handle the login correctly
	if info.URL == "https://login.ubuntu.com/+login?next=%2Fsaml%2Fprocess" {
		login := os.Getenv("U1_LOGIN")
		if len(login) == 0 {
			prompt := promptui.Prompt{Label: "Ubuntu One Login"}
			login, err = prompt.Run()
			if err != nil {
				return fmt.Errorf("failed to read login: %w", err)
			}
		}

		password := os.Getenv("U1_PASSWORD")
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

// getCandidatesPage is a helper method to construct and fetch the Candidates listing
// page for a given role, with a specified set of URL query parameters
func (g *Greenhouse) getCandidatesPage(roleId int64, queries map[string]string) (*rod.Page, error) {
	pageUrl := url.URL{}
	pageUrl.Scheme = "https"
	pageUrl.Host = "canonical.greenhouse.io"
	pageUrl.Path = fmt.Sprintf("plans/%d/candidates", roleId)

	fields := pageUrl.Query()
	fields.Add("hiring_plan_id[]", fmt.Sprintf("%d", roleId))
	fields.Add("job_status", "open")
	fields.Add("stage_status_id[]", "2")
	fields.Add("type", "all")

	for k, v := range queries {
		fields.Add(k, v)
	}

	pageUrl.RawQuery = fields.Encode()

	page, err := g.browser.Page(proto.TargetCreateTarget{URL: pageUrl.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to load page '%s': %w", pageUrl.String(), err)
	}

	err = page.WaitStable(300 * time.Millisecond)

	return page, err
}
