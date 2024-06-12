package greenhouse

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRolePopulate(t *testing.T) {
	r := NewRole(666, "Joe Bloggs")

	progress := 0
	incProgress := func(a int64) {
		progress += int(a)
	}

	g := &FakeGreenhouse{}

	err := r.Populate(g, incProgress)

	if err != nil {
		t.Errorf("error populating role: %s", err.Error())
	}

	if progress != NumRoleFields {
		t.Errorf("incorrect progress reporting when populating role, expected %d, got %d", progress, NumRoleFields)
	}

	expectedFields := map[string]int{
		"appReviews":      17,
		"needsDecision":   17,
		"needsScheduling": 17,
		"stale":           17,
		"wiGrading":       17,
		"wiScreening":     17,
	}

	if fmt.Sprint(r.fields) != fmt.Sprint(expectedFields) {
		t.Errorf("incorrect fields returned from role population")
	}
}

func TestRoleJSONMarshal(t *testing.T) {
	r := NewRole(666, "Steve Jobs")

	incProgress := func(a int64) {}
	g := &FakeGreenhouse{}
	r.Populate(g, incProgress)

	b, err := json.Marshal(r)
	if err != nil {
		t.Errorf("failed to marshal role as json: %s", err.Error())
	}

	expected := `{"id":666,"title":"Fake Role","lead":"Steve Jobs","appReviews":17,"needsDecision":17,"needsScheduling":17,"wiScreening":17,"wiGrading":17,"stale":17}`

	if string(b) != expected {
		t.Error("role marshalled incorrectly to JSON")
	}
}

type FakeGreenhouse struct{}

func (fg *FakeGreenhouse) RoleTitle(roleId int64) (string, error) {
	return "Fake Role", nil
}

func (fg *FakeGreenhouse) CandidateCount(roleId int64, query map[string]string) (int, error) {
	return 17, nil
}

func (fg *FakeGreenhouse) Login() error {
	return nil
}
