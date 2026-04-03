package rules

import "fmt"

// Status represents the outcome of a single compliance rule evaluation.
type Status string

const (
	Pass    Status = "pass"
	Fail    Status = "fail"
	Unknown Status = "unknown" // no recent telemetry data for the metric
)

// Result is the evaluated outcome of one Rule against a live metric value.
type Result struct {
	RuleID      string  `json:"rule_id"`
	Description string  `json:"description"`
	ASMERef     string  `json:"asme_ref"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Unit        string  `json:"unit"`
	Status      Status  `json:"status"`
	Message     string  `json:"message"`
}

// Summary aggregates counts of pass/fail/unknown across all evaluated rules.
type Summary struct {
	Pass    int    `json:"pass"`
	Fail    int    `json:"fail"`
	Unknown int    `json:"unknown"`
	Overall Status `json:"overall"`
}

// Summarise tallies a Result slice into a Summary.
// Overall is "fail" if any rule failed, "unknown" if any rule is unknown
// (and none failed), and "pass" only when every rule passes.
func Summarise(results []Result) Summary {
	var s Summary
	for _, r := range results {
		switch r.Status {
		case Pass:
			s.Pass++
		case Fail:
			s.Fail++
		case Unknown:
			s.Unknown++
		}
	}
	switch {
	case s.Fail > 0:
		s.Overall = Fail
	case s.Unknown > 0:
		s.Overall = Unknown
	default:
		s.Overall = Pass
	}
	return s
}

// Rule defines a single ASME A17.1 compliance check.
type Rule struct {
	ID          string
	Description string
	ASMERef     string
	Metric      string
	Threshold   float64
	Unit        string
	PassMsg     string // human-readable pass message (no value substitution needed)
	FailMsg     string // human-readable fail message
	evaluate    func(value float64) bool
}

// Evaluate runs the rule against value and returns a fully populated Result.
func (r Rule) Evaluate(value float64) Result {
	status := Pass
	msg := fmt.Sprintf("%s (%.4g %s)", r.PassMsg, value, r.Unit)
	if !r.evaluate(value) {
		status = Fail
		msg = fmt.Sprintf("%s (%.4g %s, limit %.4g %s)", r.FailMsg, value, r.Unit, r.Threshold, r.Unit)
	}
	return Result{
		RuleID:      r.ID,
		Description: r.Description,
		ASMERef:     r.ASMERef,
		Metric:      r.Metric,
		Value:       value,
		Threshold:   r.Threshold,
		Unit:        r.Unit,
		Status:      status,
		Message:     msg,
	}
}
