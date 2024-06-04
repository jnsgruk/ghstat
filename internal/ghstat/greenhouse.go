package ghstat

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/go-rod/rod"
)

// Greenhouse is an internal representation of an instance of Greenhouse
type Greenhouse struct {
	browser *rod.Browser
}

// RoleTitle takes a role ID and returns the req title as a string
func (g *Greenhouse) RoleTitle(roleId int64) string {
	slog.Debug("fetching job title for role", "role", roleId)
	page := g.getCandidatesPage(roleId, map[string]string{})
	el, err := page.Timeout(2 * time.Second).Element(".nav-title")
	if err != nil {
		slog.Debug("failed to retrieve title for role", "role", roleId, "error", err.Error())
		return "-"
	}

	return el.MustText()
}

// AppReviews returns the number of outstanding application reviews for a given role as an int
func (g *Greenhouse) AppReviews(roleId int64) int {
	slog.Debug("fetching app reviews for role", "role", roleId)
	count, err := g.getCandidateCount(roleId, map[string]string{
		"in_stages[]": "Application Review",
	})

	if err != nil {
		slog.Debug("failed to retrieve app reviews for role", "role", roleId, "error", err.Error())
		return 0
	}

	return count
}

// OutstandingWIGrading reports the number of candidates for a role who are outstanding
// a full written interview grading by assigned graders in the Hold stage
func (g *Greenhouse) OutstandingWIGrading(roleId int64) int {
	slog.Debug("fetching outstanding written interview gradings for role", "role", roleId)
	count, err := g.getCandidateCount(roleId, map[string]string{
		"take_home_test_status_id[]": "9",
		"in_stages[]":                "Hold",
		"stage_status_id[]":          "2",
	})

	if err != nil {
		slog.Debug("failed to retrieve written interview grading count for role", "role", roleId, "error", err.Error())
		return 0
	}

	return count
}

// OutstandingWIScreening reports the number of candidates for a role who are outstanding
// an initial written interview screen by the Hiring Lead for education
func (g *Greenhouse) OutstandingWIScreening(roleId int64) int {
	slog.Debug("fetching outstanding written interview initial screenings for role", "role", roleId)
	count, err := g.getCandidateCount(roleId, map[string]string{
		"take_home_test_status_id[]": "9",
		"in_stages[]":                "Written Interview",
		"stage_status_id[]":          "2",
	})

	if err != nil {
		slog.Debug("failed to retrieve written interview screening count for role", "role", roleId, "error", err.Error())
		return 0
	}

	return count
}

// OutstandingDecisions reports the number of candidates for a role who have an outstanding
// decision to be made by the hiring lead
func (g *Greenhouse) OutstandingDecisions(roleId int64) int {
	slog.Debug("fetching outstanding decisions for role", "role", roleId)
	count, err := g.getCandidateCount(roleId, map[string]string{
		"needs_decision": "1",
	})

	if err != nil {
		slog.Debug("failed to retrieve decision count for role", "role", roleId, "error", err.Error())
		return 0
	}

	return count
}

// OutstandingScheduling reports the number of candidates for a role who require scheduling,
// and who have submitted their availability
func (g *Greenhouse) OutstandingScheduling(roleId int64) int {
	slog.Debug("fetching outstanding scheduling for role", "role", roleId)
	count, err := g.getCandidateCount(roleId, map[string]string{
		"interview_status_id[]": "1",
		"availability_state":    "received",
	})

	if err != nil {
		slog.Debug("failed to retrieve outstanding scheduling count for role", "role", roleId, "error", err.Error())
		return 0
	}

	return count
}

// StaleCandidates reports the number of candidates on a role who've seen no activity in 7 days
func (g *Greenhouse) StaleCandidates(roleId int64) int {
	slog.Debug("fetching stale candidates for role", "role", roleId)
	lastWeek := time.Now().AddDate(0, 0, -7)
	count, err := g.getCandidateCount(roleId, map[string]string{
		"last_activity_end": lastWeek.Format("2006/01/02"),
	})

	if err != nil {
		slog.Debug("failed to retrieve title for role", "role", roleId, "error", err.Error())
		return 0
	}

	return count
}

// getCandidateCount is a helper method for requesting Greenhouse candidate pages with
// a specified set of query parameters in the URL
func (g *Greenhouse) getCandidateCount(roleId int64, queries map[string]string) (int, error) {
	page := g.getCandidatesPage(roleId, queries)

	// If this element is present, the number of results is zero
	_, err := page.Timeout(1 * time.Second).Element(".no_results--header")
	if err == nil {
		return 0, nil
	}

	rc, err := page.Timeout(1 * time.Second).Element("#results_count")
	if err != nil {
		return -1, fmt.Errorf("failed to retrieve candidate count")
	}

	count, err := strconv.Atoi(rc.MustText())
	if err != nil {
		return -1, fmt.Errorf("failed to parse count as integer")
	}

	return count, nil
}

// getCandidatesPage is a helper method to construct and fetch the Candidates listing
// page for a given role, with a specified set of URL query parameters
func (g *Greenhouse) getCandidatesPage(roleId int64, queries map[string]string) *rod.Page {
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

	return g.browser.MustPage(pageUrl.String()).MustWaitStable()
}
