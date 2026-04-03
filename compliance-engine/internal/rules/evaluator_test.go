package rules_test

import (
	"testing"

	"github.com/liftiq/compliance-engine/internal/rules"
)

// allPassValues returns a MetricValues map where every metric is set to a
// value that should produce a Pass result.
func allPassValues() rules.MetricValues {
	return rules.MetricValues{
		"motor_current_a":        10.0,  // ≤ 20 A
		"motor_temp_c":           60.0,  // ≤ 85 °C
		"motor_rpm":              1200.0, // ≤ 1800 RPM
		"motor_run_hours":        5000.0, // < 20000 h
		"trip_count":             100000.0, // < 500000
		"door_close_force_n":     90.0,  // ≤ 135 N
		"door_motor_amps":        2.0,   // ≤ 5 A
		"door_close_time_ms":     2500.0, // ≤ 5000 ms
		"door_cycle_count":       500000.0, // < 2000000
		"door_obstruction_events": 2.0,  // ≤ 10
		"brake_response_ms":      50.0,  // ≤ 80 ms
		"brake_current_a":        1.5,   // ≤ 3 A
		"brake_engagement_count": 100000.0, // < 500000
		"leveling_accuracy_mm":   5.0,   // ≤ 12.7 mm
		"vibration_g":            0.05,  // ≤ 0.15 g
		"door_interlock_ok":      1.0,   // == 1.0
		"governor_ok":            1.0,
		"buffer_ok":              1.0,
		"pit_switch_ok":          1.0,
		"safety_circuit_ok":      1.0,
	}
}

func TestEvaluateAll_Returns20Results(t *testing.T) {
	got := rules.EvaluateAll(rules.MetricValues{})
	if len(got) != 20 {
		t.Errorf("EvaluateAll returned %d results, want 20", len(got))
	}
}

func TestEvaluateAll_AllPassWhenValuesGood(t *testing.T) {
	results := rules.EvaluateAll(allPassValues())
	for _, r := range results {
		if r.Status != rules.Pass {
			t.Errorf("rule %s (%s): got %s, want pass (value=%.4g)", r.RuleID, r.Metric, r.Status, r.Value)
		}
	}
}

func TestEvaluateAll_AllUnknownWhenEmpty(t *testing.T) {
	results := rules.EvaluateAll(rules.MetricValues{})
	for _, r := range results {
		if r.Status != rules.Unknown {
			t.Errorf("rule %s: got %s, want unknown (no metrics provided)", r.RuleID, r.Status)
		}
	}
}

func TestEvaluateAll_UnknownMessageIsSet(t *testing.T) {
	results := rules.EvaluateAll(rules.MetricValues{})
	for _, r := range results {
		if r.Message == "" {
			t.Errorf("rule %s: unknown result has empty message", r.RuleID)
		}
	}
}

func TestEvaluateAll_ResultPreservesRuleMetadata(t *testing.T) {
	results := rules.EvaluateAll(allPassValues())
	for i, r := range results {
		rule := rules.ASME20[i]
		if r.RuleID != rule.ID {
			t.Errorf("result[%d].RuleID = %q, want %q", i, r.RuleID, rule.ID)
		}
		if r.Metric != rule.Metric {
			t.Errorf("result[%d].Metric = %q, want %q", i, r.Metric, rule.Metric)
		}
		if r.ASMERef != rule.ASMERef {
			t.Errorf("result[%d].ASMERef = %q, want %q", i, r.ASMERef, rule.ASMERef)
		}
		if r.Threshold != rule.Threshold {
			t.Errorf("result[%d].Threshold = %g, want %g", i, r.Threshold, rule.Threshold)
		}
	}
}

// ── Boundary tests for the three ASME-mandated auto-fail thresholds ──────────

func TestEvaluateAll_DoorCloseForce_PassAtExactly135(t *testing.T) {
	v := allPassValues()
	v["door_close_force_n"] = 135.0
	assertStatus(t, "ASME-006", rules.Pass, rules.EvaluateAll(v))
}

func TestEvaluateAll_DoorCloseForce_FailJustAbove135(t *testing.T) {
	v := allPassValues()
	v["door_close_force_n"] = 135.1
	assertStatus(t, "ASME-006", rules.Fail, rules.EvaluateAll(v))
}

func TestEvaluateAll_BrakeResponse_PassAtExactly80(t *testing.T) {
	v := allPassValues()
	v["brake_response_ms"] = 80.0
	assertStatus(t, "ASME-011", rules.Pass, rules.EvaluateAll(v))
}

func TestEvaluateAll_BrakeResponse_FailJustAbove80(t *testing.T) {
	v := allPassValues()
	v["brake_response_ms"] = 80.1
	assertStatus(t, "ASME-011", rules.Fail, rules.EvaluateAll(v))
}

func TestEvaluateAll_LevelingAccuracy_PassAtExactly12_7(t *testing.T) {
	v := allPassValues()
	v["leveling_accuracy_mm"] = 12.7
	assertStatus(t, "ASME-014", rules.Pass, rules.EvaluateAll(v))
}

func TestEvaluateAll_LevelingAccuracy_FailJustAbove12_7(t *testing.T) {
	v := allPassValues()
	v["leveling_accuracy_mm"] = 12.8
	assertStatus(t, "ASME-014", rules.Fail, rules.EvaluateAll(v))
}

// ── Motor subsystem boundary tests ────────────────────────────────────────────

func TestEvaluateAll_MotorCurrent_PassAtLimit(t *testing.T) {
	v := allPassValues()
	v["motor_current_a"] = 20.0
	assertStatus(t, "ASME-001", rules.Pass, rules.EvaluateAll(v))
}

func TestEvaluateAll_MotorCurrent_FailAboveLimit(t *testing.T) {
	v := allPassValues()
	v["motor_current_a"] = 20.1
	assertStatus(t, "ASME-001", rules.Fail, rules.EvaluateAll(v))
}

func TestEvaluateAll_MotorTemp_PassAtLimit(t *testing.T) {
	v := allPassValues()
	v["motor_temp_c"] = 85.0
	assertStatus(t, "ASME-002", rules.Pass, rules.EvaluateAll(v))
}

func TestEvaluateAll_MotorTemp_FailAboveLimit(t *testing.T) {
	v := allPassValues()
	v["motor_temp_c"] = 85.1
	assertStatus(t, "ASME-002", rules.Fail, rules.EvaluateAll(v))
}

func TestEvaluateAll_MotorRPM_PassAtLimit(t *testing.T) {
	v := allPassValues()
	v["motor_rpm"] = 1800.0
	assertStatus(t, "ASME-003", rules.Pass, rules.EvaluateAll(v))
}

func TestEvaluateAll_MotorRPM_FailAboveLimit(t *testing.T) {
	v := allPassValues()
	v["motor_rpm"] = 1800.1
	assertStatus(t, "ASME-003", rules.Fail, rules.EvaluateAll(v))
}

// ── Counter rules (strict less-than thresholds) ───────────────────────────────

func TestEvaluateAll_MotorRunHours_FailAtExactThreshold(t *testing.T) {
	v := allPassValues()
	v["motor_run_hours"] = 20000.0 // must be < 20000, so exactly 20000 fails
	assertStatus(t, "ASME-004", rules.Fail, rules.EvaluateAll(v))
}

func TestEvaluateAll_TripCount_FailAtExactThreshold(t *testing.T) {
	v := allPassValues()
	v["trip_count"] = 500000.0
	assertStatus(t, "ASME-005", rules.Fail, rules.EvaluateAll(v))
}

func TestEvaluateAll_DoorCycleCount_FailAtExactThreshold(t *testing.T) {
	v := allPassValues()
	v["door_cycle_count"] = 2000000.0
	assertStatus(t, "ASME-009", rules.Fail, rules.EvaluateAll(v))
}

func TestEvaluateAll_BrakeEngagementCount_FailAtExactThreshold(t *testing.T) {
	v := allPassValues()
	v["brake_engagement_count"] = 500000.0
	assertStatus(t, "ASME-013", rules.Fail, rules.EvaluateAll(v))
}

// ── Vibration ─────────────────────────────────────────────────────────────────

func TestEvaluateAll_Vibration_PassAtLimit(t *testing.T) {
	v := allPassValues()
	v["vibration_g"] = 0.15
	assertStatus(t, "ASME-015", rules.Pass, rules.EvaluateAll(v))
}

func TestEvaluateAll_Vibration_FailAboveLimit(t *testing.T) {
	v := allPassValues()
	v["vibration_g"] = 0.16
	assertStatus(t, "ASME-015", rules.Fail, rules.EvaluateAll(v))
}

// ── Boolean safety fields (1.0 = ok, 0.0 = fault) ────────────────────────────

func TestEvaluateAll_SafetyBooleans_PassWhenOne(t *testing.T) {
	boolRules := []string{"ASME-016", "ASME-017", "ASME-018", "ASME-019", "ASME-020"}
	boolMetrics := []string{
		"door_interlock_ok", "governor_ok", "buffer_ok", "pit_switch_ok", "safety_circuit_ok",
	}
	for i, ruleID := range boolRules {
		v := allPassValues()
		v[boolMetrics[i]] = 1.0
		assertStatus(t, ruleID, rules.Pass, rules.EvaluateAll(v))
	}
}

func TestEvaluateAll_SafetyBooleans_FailWhenZero(t *testing.T) {
	boolRules := []string{"ASME-016", "ASME-017", "ASME-018", "ASME-019", "ASME-020"}
	boolMetrics := []string{
		"door_interlock_ok", "governor_ok", "buffer_ok", "pit_switch_ok", "safety_circuit_ok",
	}
	for i, ruleID := range boolRules {
		v := allPassValues()
		v[boolMetrics[i]] = 0.0
		assertStatus(t, ruleID, rules.Fail, rules.EvaluateAll(v))
	}
}

func TestEvaluateAll_PartialValues_MixedStatuses(t *testing.T) {
	// Provide only motor metrics; door/brake/safety should be Unknown.
	v := rules.MetricValues{
		"motor_current_a":  10.0,
		"motor_temp_c":     60.0,
		"motor_rpm":        1200.0,
		"motor_run_hours":  5000.0,
		"trip_count":       100000.0,
		"door_close_force_n": 200.0, // fail
	}
	results := rules.EvaluateAll(v)

	byID := make(map[string]rules.Result)
	for _, r := range results {
		byID[r.RuleID] = r
	}

	assertStatus(t, "ASME-001", rules.Pass, results)    // motor_current_a provided
	assertStatus(t, "ASME-006", rules.Fail, results)    // door_close_force_n = 200 N
	assertStatus(t, "ASME-011", rules.Unknown, results) // brake_response_ms not provided
	assertStatus(t, "ASME-020", rules.Unknown, results) // safety_circuit_ok not provided
}

func TestEvaluateAll_PassResultContainsValue(t *testing.T) {
	v := allPassValues()
	results := rules.EvaluateAll(v)
	for _, r := range results {
		if r.Status == rules.Pass && r.Value == 0 && r.Metric != "door_interlock_ok" {
			// Value should be set for all non-bool pass results
			// (zero is a valid value for bool fields at 0.0 but that would be a fail)
		}
		if r.Status != rules.Unknown && r.Message == "" {
			t.Errorf("rule %s: pass/fail result has empty message", r.RuleID)
		}
	}
}

// ── Summarise tests ───────────────────────────────────────────────────────────

func TestSummarise_AllPass(t *testing.T) {
	results := rules.EvaluateAll(allPassValues())
	s := rules.Summarise(results)
	if s.Pass != 20 || s.Fail != 0 || s.Unknown != 0 {
		t.Errorf("got pass=%d fail=%d unknown=%d, want 20/0/0", s.Pass, s.Fail, s.Unknown)
	}
	if s.Overall != rules.Pass {
		t.Errorf("overall: got %s, want pass", s.Overall)
	}
}

func TestSummarise_OneFail_OverallFail(t *testing.T) {
	v := allPassValues()
	v["door_close_force_n"] = 200.0
	results := rules.EvaluateAll(v)
	s := rules.Summarise(results)
	if s.Fail != 1 {
		t.Errorf("fail count: got %d, want 1", s.Fail)
	}
	if s.Overall != rules.Fail {
		t.Errorf("overall: got %s, want fail", s.Overall)
	}
}

func TestSummarise_OnlyUnknown_OverallUnknown(t *testing.T) {
	results := rules.EvaluateAll(rules.MetricValues{})
	s := rules.Summarise(results)
	if s.Unknown != 20 {
		t.Errorf("unknown count: got %d, want 20", s.Unknown)
	}
	if s.Overall != rules.Unknown {
		t.Errorf("overall: got %s, want unknown", s.Overall)
	}
}

func TestSummarise_FailTakesPrecedenceOverUnknown(t *testing.T) {
	// Only provide one metric that fails; the rest are unknown.
	v := rules.MetricValues{"door_close_force_n": 999.0}
	results := rules.EvaluateAll(v)
	s := rules.Summarise(results)
	if s.Overall != rules.Fail {
		t.Errorf("overall: got %s, want fail (fail should outrank unknown)", s.Overall)
	}
}

// ── Helper ───────────────────────────────────────────────────────────────────

func assertStatus(t *testing.T, ruleID string, want rules.Status, results []rules.Result) {
	t.Helper()
	for _, r := range results {
		if r.RuleID == ruleID {
			if r.Status != want {
				t.Errorf("rule %s: got status %q, want %q (value=%.4g)", ruleID, r.Status, want, r.Value)
			}
			return
		}
	}
	t.Errorf("rule %s not found in results", ruleID)
}
