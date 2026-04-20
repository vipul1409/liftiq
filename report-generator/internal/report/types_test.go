package report

import (
	"testing"
)

// ── GroupBySubsystem ──────────────────────────────────────────────────────────

func TestGroupBySubsystem_ReturnsExpectedSubsystems(t *testing.T) {
	results := allPassResults()
	groups := GroupBySubsystem(results)

	wantTitles := []string{"Motor", "Trip / Usage", "Door Operator", "Brake System", "Ride Quality", "Safety Circuits"}
	if len(groups) != len(wantTitles) {
		t.Fatalf("got %d groups, want %d", len(groups), len(wantTitles))
	}
	for i, g := range groups {
		if g.Title != wantTitles[i] {
			t.Errorf("group[%d].Title = %q, want %q", i, g.Title, wantTitles[i])
		}
	}
}

func TestGroupBySubsystem_CorrectRuleCountsPerSubsystem(t *testing.T) {
	results := allPassResults()
	groups := GroupBySubsystem(results)

	wantCounts := map[string]int{
		"Motor":          4,
		"Trip / Usage":   1,
		"Door Operator":  5,
		"Brake System":   3,
		"Ride Quality":   2,
		"Safety Circuits": 5,
	}
	for _, g := range groups {
		want := wantCounts[g.Title]
		if len(g.Results) != want {
			t.Errorf("group %q: got %d rules, want %d", g.Title, len(g.Results), want)
		}
	}
}

func TestGroupBySubsystem_EmptyInputReturnsNoGroups(t *testing.T) {
	groups := GroupBySubsystem(nil)
	if len(groups) != 0 {
		t.Errorf("got %d groups for empty input, want 0", len(groups))
	}
}

func TestGroupBySubsystem_IgnoresUnknownRuleIDs(t *testing.T) {
	results := []RuleResult{
		{RuleID: "ASME-001", Status: StatusPass},
		{RuleID: "ASME-999", Status: StatusPass}, // not in any subsystem
	}
	groups := GroupBySubsystem(results)
	for _, g := range groups {
		for _, r := range g.Results {
			if r.RuleID == "ASME-999" {
				t.Errorf("unexpected rule ASME-999 appeared in group %q", g.Title)
			}
		}
	}
}

func TestGroupBySubsystem_PreservesRuleOrder(t *testing.T) {
	results := allPassResults()
	groups := GroupBySubsystem(results)

	// Motor group should be ASME-001 … ASME-004 in order.
	motor := groups[0]
	wantOrder := []string{"ASME-001", "ASME-002", "ASME-003", "ASME-004"}
	for i, r := range motor.Results {
		if r.RuleID != wantOrder[i] {
			t.Errorf("motor[%d] = %q, want %q", i, r.RuleID, wantOrder[i])
		}
	}
}

// ── PhotosByRule ──────────────────────────────────────────────────────────────

func TestPhotosByRule_EmptyInputReturnsEmptyMap(t *testing.T) {
	m := PhotosByRule(nil)
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestPhotosByRule_GroupsPhotosByRuleID(t *testing.T) {
	photos := []Photo{
		{RuleID: "ASME-006", DataURI: "data:image/jpeg;base64,aaa"},
		{RuleID: "ASME-006", DataURI: "data:image/jpeg;base64,bbb"},
		{RuleID: "ASME-011", DataURI: "data:image/jpeg;base64,ccc"},
	}
	m := PhotosByRule(photos)

	if len(m["ASME-006"]) != 2 {
		t.Errorf("ASME-006: got %d photos, want 2", len(m["ASME-006"]))
	}
	if len(m["ASME-011"]) != 1 {
		t.Errorf("ASME-011: got %d photos, want 1", len(m["ASME-011"]))
	}
}

func TestPhotosByRule_PreservesPhotoOrder(t *testing.T) {
	photos := []Photo{
		{RuleID: "ASME-006", DataURI: "data:image/jpeg;base64,first"},
		{RuleID: "ASME-006", DataURI: "data:image/jpeg;base64,second"},
	}
	m := PhotosByRule(photos)

	if m["ASME-006"][0].DataURI != "data:image/jpeg;base64,first" {
		t.Errorf("first photo out of order")
	}
	if m["ASME-006"][1].DataURI != "data:image/jpeg;base64,second" {
		t.Errorf("second photo out of order")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func allPassResults() []RuleResult {
	allIDs := []string{
		"ASME-001", "ASME-002", "ASME-003", "ASME-004",
		"ASME-005",
		"ASME-006", "ASME-007", "ASME-008", "ASME-009", "ASME-010",
		"ASME-011", "ASME-012", "ASME-013",
		"ASME-014", "ASME-015",
		"ASME-016", "ASME-017", "ASME-018", "ASME-019", "ASME-020",
	}
	results := make([]RuleResult, len(allIDs))
	for i, id := range allIDs {
		results[i] = RuleResult{
			RuleID:      id,
			Description: "Test rule " + id,
			ASMERef:     "ASME A17.1 2.1",
			Metric:      "test_metric",
			Value:       1.0,
			Threshold:   100.0,
			Unit:        "unit",
			Status:      StatusPass,
			Message:     "within limit",
		}
	}
	return results
}
