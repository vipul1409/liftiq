package report

import "time"

// Status mirrors the compliance engine's rule status values.
type Status string

const (
	StatusPass    Status = "pass"
	StatusFail    Status = "fail"
	StatusUnknown Status = "unknown"
)

// RuleResult is one evaluated ASME A17.1 rule included in the report.
type RuleResult struct {
	RuleID      string  `json:"rule_id"`
	Description string  `json:"description"`
	ASMERef     string  `json:"asme_ref"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Unit        string  `json:"unit"`
	Status      Status  `json:"status"`
	Message     string  `json:"message"`
	// Overridden is true when the technician manually overrode the telemetry result.
	Overridden bool `json:"overridden"`
}

// Summary is the aggregate pass/fail/unknown count across all rules.
type Summary struct {
	Pass    int    `json:"pass"`
	Fail    int    `json:"fail"`
	Unknown int    `json:"unknown"`
	Overall Status `json:"overall"`
}

// Photo is a piece of photo evidence attached to a specific rule.
type Photo struct {
	RuleID    string   `json:"rule_id"`
	DataURI   string   `json:"data_uri"`   // "data:image/jpeg;base64,..."
	Timestamp string   `json:"timestamp"`  // ISO-8601
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// Request is the JSON body for POST /reports.
type Request struct {
	UnitTag          string       `json:"unit_tag"`
	InspectedAt      time.Time    `json:"inspected_at"`
	Technician       string       `json:"technician"`
	Results          []RuleResult `json:"results"`
	Summary          Summary      `json:"summary"`
	Photos           []Photo      `json:"photos"`
	SignatureDataURI string       `json:"signature_data_uri,omitempty"` // "data:image/svg+xml;base64,..."
}

// subsystemOrder defines the display order of rule subsystems in the report.
var subsystemOrder = []struct {
	Title   string
	RuleIDs []string
}{
	{"Motor", []string{"ASME-001", "ASME-002", "ASME-003", "ASME-004"}},
	{"Trip / Usage", []string{"ASME-005"}},
	{"Door Operator", []string{"ASME-006", "ASME-007", "ASME-008", "ASME-009", "ASME-010"}},
	{"Brake System", []string{"ASME-011", "ASME-012", "ASME-013"}},
	{"Ride Quality", []string{"ASME-014", "ASME-015"}},
	{"Safety Circuits", []string{"ASME-016", "ASME-017", "ASME-018", "ASME-019", "ASME-020"}},
}

// SubsystemGroup groups rule results under a subsystem heading.
type SubsystemGroup struct {
	Title   string
	Results []RuleResult
}

// GroupBySubsystem organises results into subsystem groups matching
// the order used in the compliance checklist.
func GroupBySubsystem(results []RuleResult) []SubsystemGroup {
	byID := make(map[string]RuleResult, len(results))
	for _, r := range results {
		byID[r.RuleID] = r
	}
	var groups []SubsystemGroup
	for _, sub := range subsystemOrder {
		var grouped []RuleResult
		for _, id := range sub.RuleIDs {
			if r, ok := byID[id]; ok {
				grouped = append(grouped, r)
			}
		}
		if len(grouped) > 0 {
			groups = append(groups, SubsystemGroup{Title: sub.Title, Results: grouped})
		}
	}
	return groups
}

// PhotosByRule returns a map of ruleID → []Photo.
func PhotosByRule(photos []Photo) map[string][]Photo {
	m := make(map[string][]Photo)
	for _, p := range photos {
		m[p.RuleID] = append(m[p.RuleID], p)
	}
	return m
}
