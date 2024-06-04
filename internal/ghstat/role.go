package ghstat

import (
	"fmt"
)

// Role represents a given req on Greenhouse
type Role struct {
	RawID              int64  `json:"id"`
	RawName            string `json:"roleName"`
	RawLead            string `json:"hiringLead"`
	RawAppReviews      int    `json:"appReviews"`
	RawNeedsDecision   int    `json:"needsDecision"`
	RawNeedsScheduling int    `json:"needsScheduling"`
	RawWIScreening     int    `json:"wiScreening"`
	RawWIGrading       int    `json:"wiGrading"`
	RawStale           int    `json:"stale"`
}

// ID returns the role ID formatted as a string
func (r *Role) ID() string {
	return fmt.Sprintf("%d", r.RawID)
}

// Name returns the role title formatted as a string
func (r *Role) Name() string {
	return r.RawName
}

// Lead returns the role's hiring lead formatted as a string
func (r *Role) Lead() string {
	return r.RawLead
}

// AppReviews returns the number of outstanding application reviews
// for the role, formatted as as string
func (r *Role) AppReviews() string {
	return fmt.Sprintf("%d", r.RawAppReviews)
}

// NeedsDecision returns the number of outstanding decisions to be
// made about candidates on a role, formatted as a string
func (r *Role) NeedsDecision() string {
	return fmt.Sprintf("%d", r.RawNeedsDecision)
}

// NeedsScheduling returns the number of outstanding interviews
// to be scheduled for the role, formatted as as string
func (r *Role) NeedsScheduling() string {
	return fmt.Sprintf("%d", r.RawNeedsScheduling)
}

// WIScreening returns the number of outstanding written
// interview screenings for the role, formatted as as string
func (r *Role) WIScreening() string {
	return fmt.Sprintf("%d", r.RawWIScreening)
}

// WIGrading returns the number of outstanding written
// interview gradings for the role, formatted as as string
func (r *Role) WIGrading() string {
	return fmt.Sprintf("%d", r.RawWIGrading)
}

// Stale returns the number of candidates who have seen no activity
// for 7 days or longer, formatted as a string
func (r *Role) Stale() string {
	return fmt.Sprintf("%d", r.RawStale)
}
