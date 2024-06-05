package greenhouse

import (
	"encoding/json"
	"log/slog"
	"time"
)

// NumRoleFields is the number of fields that are fetched from greenhouse.
// The +1 is for the title, on top of the numeric fields map
var NumRoleFields = len(filters) + 1

// Role represents a given req on Greenhouse
type Role struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Lead   string `json:"lead"`
	fields map[string]int
}

// NewRole constructs a new Role with a given ID
func NewRole(id int64, lead string) *Role {
	return &Role{
		ID:     id,
		Lead:   lead,
		fields: make(map[string]int),
	}
}

// Type alias for a set of Greenhouse queries
type filterSet map[string]string

// filters represents the various URL query parameters required to
// filter the candidate page with to acquire the correct value
var filters map[string]filterSet = map[string]filterSet{
	"appReviews": {
		"in_stages[]": "Application Review",
	},
	"needsDecision": {
		"needs_decision": "1",
	},
	"needsScheduling": {
		"interview_status_id[]": "1",
		"availability_state":    "received",
	},
	"wiScreening": {
		"take_home_test_status_id[]": "9",
		"in_stages[]":                "Written Interview",
		"stage_status_id[]":          "2",
	},
	"wiGrading": {
		"take_home_test_status_id[]": "9",
		"in_stages[]":                "Hold",
		"stage_status_id[]":          "2",
	},
	"stale": {
		"last_activity_end": time.Now().AddDate(0, 0, -7).Format("2006/01/02"),
	},
}

// Populate is used to fetch the details of each field from Greenhouse using
// the specified filters
func (r *Role) Populate(g *Greenhouse, incProgress func(amount int64)) error {
	slog.Debug("processing role", "roleId", r.ID, "lead", r.Lead)

	r.Title = r.fetchTitle(g)
	incProgress(1)

	for k, v := range filters {
		count, err := g.CandidateCount(r.ID, v)
		if err != nil {
			slog.Debug("failed to retrieve field", "role", r.ID, "field", k, "error", err.Error())
			count = 0
		}
		r.fields[k] = count
		incProgress(1)
	}

	return nil
}

// fetchTitle is a helper method for fetching the title of the req from
// Greenhouse
func (r *Role) fetchTitle(g *Greenhouse) string {
	page, err := g.getCandidatesPage(r.ID, map[string]string{})
	if err != nil {
		return ""
	}

	el, err := page.Element(".nav-title")
	if err != nil {
		slog.Debug("failed to retrieve title for role", "role", r.ID, "error", err.Error())
		return ""
	}

	text, err := el.Text()
	if err != nil {
		return ""
	}

	return text
}

// AppReviews returns the number of outstanding application reviews
// for the role
func (r *Role) AppReviews() int {
	return r.fields["appReviews"]
}

// NeedsDecision returns the number of outstanding decisions to be
// made about candidates on a role
func (r *Role) NeedsDecision() int {
	return r.fields["needsDecision"]
}

// NeedsScheduling returns the number of outstanding interviews
// to be scheduled for the role where the candidate has
// already submitted their availability
func (r *Role) NeedsScheduling() int {
	return r.fields["needsScheduling"]
}

// WIScreening returns the number of outstanding written
// interview screenings for the role
func (r *Role) WIScreening() int {
	return r.fields["wiScreening"]
}

// WIGrading returns the number of outstanding written
// interview gradings for the role
func (r *Role) WIGrading() int {
	return r.fields["wiGrading"]
}

// Stale returns the number of candidates who have seen no activity
// for 7 days or longer
func (r *Role) Stale() int {
	return r.fields["stale"]
}

// MarshalJSON implements a custom marshaller to get the output format we want
func (r *Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID              int64  `json:"id"`
		Title           string `json:"title"`
		Lead            string `json:"lead"`
		AppReviews      int    `json:"appReviews"`
		NeedsDecision   int    `json:"needsDecision"`
		NeedsScheduling int    `json:"needsScheduling"`
		WIScreening     int    `json:"wiScreening"`
		WIGrading       int    `json:"wiGrading"`
		Stale           int    `json:"stale"`
	}{
		ID:              r.ID,
		Title:           r.Title,
		Lead:            r.Lead,
		AppReviews:      r.AppReviews(),
		NeedsDecision:   r.NeedsDecision(),
		NeedsScheduling: r.NeedsScheduling(),
		WIScreening:     r.WIScreening(),
		WIGrading:       r.WIGrading(),
		Stale:           r.Stale(),
	})
}
