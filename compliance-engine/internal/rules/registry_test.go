package rules_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liftiq/compliance-engine/internal/rules"
)

// knownMetrics loads the telemetry metric contract shared with the simulator
// and the telemetry-ingestor (contracts/telemetry-metrics.json).
func knownMetrics(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "contracts", "telemetry-metrics.json"))
	if err != nil {
		t.Fatalf("read metric contract: %v", err)
	}
	var metrics []string
	if err := json.Unmarshal(b, &metrics); err != nil {
		t.Fatalf("decode metric contract: %v", err)
	}
	return metrics
}

func TestASME20_HasExactly20Rules(t *testing.T) {
	if got := len(rules.ASME20); got != 20 {
		t.Errorf("ASME20 has %d rules, want 20", got)
	}
}

func TestASME20_AllIDsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, r := range rules.ASME20 {
		if seen[r.ID] {
			t.Errorf("duplicate rule ID: %q", r.ID)
		}
		seen[r.ID] = true
	}
}

func TestASME20_IDsFollowConvention(t *testing.T) {
	for i, r := range rules.ASME20 {
		want := "ASME-" + zpad(i+1)
		if r.ID != want {
			t.Errorf("rule %d: ID %q, want %q", i, r.ID, want)
		}
	}
}

func zpad(n int) string {
	s := "000"
	v := s + strings.TrimLeft(func() string {
		b := make([]byte, 3)
		b[0] = byte('0' + n/100)
		b[1] = byte('0' + (n/10)%10)
		b[2] = byte('0' + n%10)
		return string(b)
	}(), "0")
	return v[len(v)-3:]
}

func TestASME20_AllMetricsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, r := range rules.ASME20 {
		if seen[r.Metric] {
			t.Errorf("duplicate metric in registry: %q", r.Metric)
		}
		seen[r.Metric] = true
	}
}

func TestASME20_AllMetricsMatchTelemetrySchema(t *testing.T) {
	known := make(map[string]bool)
	for _, m := range knownMetrics(t) {
		known[m] = true
	}
	for _, r := range rules.ASME20 {
		if !known[r.Metric] {
			t.Errorf("rule %q references metric %q which is not in the telemetry schema", r.ID, r.Metric)
		}
	}
}

func TestASME20_AllTelemetryMetricsCovered(t *testing.T) {
	covered := make(map[string]bool)
	for _, r := range rules.ASME20 {
		covered[r.Metric] = true
	}
	for _, m := range knownMetrics(t) {
		if !covered[m] {
			t.Errorf("telemetry metric %q has no corresponding ASME rule", m)
		}
	}
}

func TestASME20_AllFieldsNonEmpty(t *testing.T) {
	for _, r := range rules.ASME20 {
		if r.ID == "" {
			t.Errorf("rule has empty ID")
		}
		if r.Description == "" {
			t.Errorf("rule %q has empty Description", r.ID)
		}
		if r.ASMERef == "" {
			t.Errorf("rule %q has empty ASMERef", r.ID)
		}
		if r.Metric == "" {
			t.Errorf("rule %q has empty Metric", r.ID)
		}
		if r.Unit == "" {
			t.Errorf("rule %q has empty Unit", r.ID)
		}
		if r.PassMsg == "" {
			t.Errorf("rule %q has empty PassMsg", r.ID)
		}
		if r.FailMsg == "" {
			t.Errorf("rule %q has empty FailMsg", r.ID)
		}
	}
}

func TestASME20_AllThresholdsPositive(t *testing.T) {
	for _, r := range rules.ASME20 {
		if r.Threshold <= 0 {
			t.Errorf("rule %q has non-positive threshold %g", r.ID, r.Threshold)
		}
	}
}

func TestASME20_AllASMERefsContainASMEA17(t *testing.T) {
	for _, r := range rules.ASME20 {
		if !strings.Contains(r.ASMERef, "ASME A17.1") {
			t.Errorf("rule %q ASMERef %q does not contain 'ASME A17.1'", r.ID, r.ASMERef)
		}
	}
}
