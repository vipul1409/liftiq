package rules_test

import (
	"testing"

	"github.com/liftiq/compliance-engine/internal/rules"
)

// ruleSpec is the independent source of truth for each rule's catalogue
// entry, written out from the ASME A17.1 mapping in the engineering plan.
// atLimit is the status when the metric equals the threshold exactly.
var ruleSpec = []struct {
	id         string
	subsystem  string
	comparison rules.Comparison
	limit      float64
	atLimit    rules.Status
	justOver   float64
}{
	{"ASME-001", "Motor", rules.AtMost, 20, rules.Pass, 20.01},
	{"ASME-002", "Motor", rules.AtMost, 85, rules.Pass, 85.1},
	{"ASME-003", "Motor", rules.AtMost, 1800, rules.Pass, 1801},
	{"ASME-004", "Motor", rules.Below, 20000, rules.Fail, 20001},
	{"ASME-005", "Trip / Usage", rules.Below, 500000, rules.Fail, 500001},
	{"ASME-006", "Door Operator", rules.AtMost, 135, rules.Pass, 135.1},
	{"ASME-007", "Door Operator", rules.AtMost, 5, rules.Pass, 5.01},
	{"ASME-008", "Door Operator", rules.AtMost, 5000, rules.Pass, 5001},
	{"ASME-009", "Door Operator", rules.Below, 2000000, rules.Fail, 2000001},
	{"ASME-010", "Door Operator", rules.AtMost, 10, rules.Pass, 11},
	{"ASME-011", "Brake System", rules.AtMost, 80, rules.Pass, 81},
	{"ASME-012", "Brake System", rules.AtMost, 3, rules.Pass, 3.01},
	{"ASME-013", "Brake System", rules.Below, 500000, rules.Fail, 500001},
	{"ASME-014", "Ride Quality", rules.AtMost, 12.7, rules.Pass, 12.71},
	{"ASME-015", "Ride Quality", rules.AtMost, 0.15, rules.Pass, 0.151},
	{"ASME-016", "Safety Circuits", rules.Equals, 1, rules.Pass, 0},
	{"ASME-017", "Safety Circuits", rules.Equals, 1, rules.Pass, 0},
	{"ASME-018", "Safety Circuits", rules.Equals, 1, rules.Pass, 0},
	{"ASME-019", "Safety Circuits", rules.Equals, 1, rules.Pass, 0},
	{"ASME-020", "Safety Circuits", rules.Equals, 1, rules.Pass, 0},
}

func ruleByID(t *testing.T, id string) rules.Rule {
	t.Helper()
	for _, r := range rules.ASME20 {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("rule %s not in catalogue", id)
	return rules.Rule{}
}

func TestCatalogue_MatchesSpec(t *testing.T) {
	if len(ruleSpec) != len(rules.ASME20) {
		t.Fatalf("spec has %d rules, catalogue has %d", len(ruleSpec), len(rules.ASME20))
	}
	for _, s := range ruleSpec {
		r := ruleByID(t, s.id)
		if r.Subsystem != s.subsystem {
			t.Errorf("%s subsystem = %q, want %q", s.id, r.Subsystem, s.subsystem)
		}
		if r.Comparison != s.comparison {
			t.Errorf("%s comparison = %q, want %q", s.id, r.Comparison, s.comparison)
		}
		if r.Threshold != s.limit {
			t.Errorf("%s threshold = %g, want %g", s.id, r.Threshold, s.limit)
		}
	}
}

func TestEvaluate_BoundaryForEveryRule(t *testing.T) {
	for _, s := range ruleSpec {
		r := ruleByID(t, s.id)
		if got := r.Evaluate(s.limit).Status; got != s.atLimit {
			t.Errorf("%s at limit %g: status %q, want %q", s.id, s.limit, got, s.atLimit)
		}
		if got := r.Evaluate(s.justOver).Status; got != rules.Fail {
			t.Errorf("%s at %g: status %q, want fail", s.id, s.justOver, got)
		}
	}
}

func TestEvaluateAll_ResultsCarrySubsystemAndComparison(t *testing.T) {
	results := rules.EvaluateAll(rules.MetricValues{"door_close_force_n": 90})
	for i, res := range results {
		s := ruleSpec[i]
		if res.Subsystem != s.subsystem || res.Comparison != s.comparison {
			t.Errorf("%s result = (%q, %q), want (%q, %q)",
				res.RuleID, res.Subsystem, res.Comparison, s.subsystem, s.comparison)
		}
	}
}
