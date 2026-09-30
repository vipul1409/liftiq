package report

import "time"

// Status mirrors the compliance engine's rule status values.
type Status string

const (
	StatusPass    Status = "pass"
	StatusFail    Status = "fail"
	StatusUnknown Status = "unknown"
)

// RuleResult is one ASME A17.1 rule as evaluated from telemetry by the
// compliance engine. It is evidence: Resolve never modifies it.
type RuleResult struct {
	RuleID      string  `json:"rule_id"`
	Description string  `json:"description"`
	ASMERef     string  `json:"asme_ref"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Unit        string  `json:"unit"`
	// Comparison is "at_most" (≤), "below" (<) or "equals" (=), as defined by
	// the compliance engine's rule catalogue.
	Comparison string `json:"comparison"`
	// Subsystem is the display group, e.g. "Door Operator". Empty → "Other".
	Subsystem string `json:"subsystem"`
	Status    Status `json:"status"`
	Message   string `json:"message"`
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
//
// Results carry the telemetry status exactly as the compliance engine returned
// it. Overrides maps rule ID → the technician's manual call ("pass" or "fail").
// The report's effective statuses and summary are derived by Resolve; clients
// do not send a summary.
type Request struct {
	UnitTag          string            `json:"unit_tag"`
	InspectedAt      time.Time         `json:"inspected_at"`
	Technician       string            `json:"technician"`
	Results          []RuleResult      `json:"results"`
	Overrides        map[string]Status `json:"overrides"`
	Photos           []Photo           `json:"photos"`
	SignatureDataURI string            `json:"signature_data_uri,omitempty"` // "data:image/svg+xml;base64,..."
}

// SubsystemGroup groups report rows under a subsystem heading.
type SubsystemGroup struct {
	Title   string
	Results []Row
}

// otherSubsystem is the group for rows whose result carried no subsystem, so a
// rule is never silently dropped from the report.
const otherSubsystem = "Other"

// GroupBySubsystem organises rows by their Subsystem, in the order each
// subsystem first appears (the compliance engine returns rules in catalogue
// order). Rows without a subsystem are grouped under "Other", last.
func GroupBySubsystem(rows []Row) []SubsystemGroup {
	var groups []SubsystemGroup
	index := make(map[string]int)
	var other []Row
	for _, r := range rows {
		if r.Subsystem == "" {
			other = append(other, r)
			continue
		}
		i, ok := index[r.Subsystem]
		if !ok {
			i = len(groups)
			index[r.Subsystem] = i
			groups = append(groups, SubsystemGroup{Title: r.Subsystem})
		}
		groups[i].Results = append(groups[i].Results, r)
	}
	if len(other) > 0 {
		groups = append(groups, SubsystemGroup{Title: otherSubsystem, Results: other})
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
