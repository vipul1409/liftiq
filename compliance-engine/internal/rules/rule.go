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
	RuleID      string     `json:"rule_id"`
	Description string     `json:"description"`
	ASMERef     string     `json:"asme_ref"`
	Metric      string     `json:"metric"`
	Value       float64    `json:"value"`
	Threshold   float64    `json:"threshold"`
	Unit        string     `json:"unit"`
	Comparison  Comparison `json:"comparison"`
	Subsystem   string     `json:"subsystem"`
	Status      Status     `json:"status"`
	Message     string     `json:"message"`
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

// Comparison is how a metric value is checked against a rule's Threshold.
type Comparison string

const (
	AtMost Comparison = "at_most" // pass if value ≤ threshold
	Below  Comparison = "below"   // pass if value < threshold (service due at N)
	Equals Comparison = "equals"  // pass if value == threshold (boolean safety fields)
)

// passes reports whether value satisfies the comparison against threshold.
func (c Comparison) passes(value, threshold float64) bool {
	switch c {
	case AtMost:
		return value <= threshold
	case Below:
		return value < threshold
	case Equals:
		return value == threshold
	default:
		return false // an unrecognised comparison never passes a safety check
	}
}

// Rule defines a single ASME A17.1 compliance check. A rule is pure data:
// its pass/fail behaviour is fully determined by Comparison and Threshold.
type Rule struct {
	ID          string
	Description string
	ASMERef     string
	Subsystem   string // display grouping, e.g. "Door Operator"
	Metric      string
	Comparison  Comparison
	Threshold   float64
	Unit        string
	PassMsg     string // human-readable pass message (no value substitution needed)
	FailMsg     string // human-readable fail message
}

// Evaluate runs the rule against value and returns a fully populated Result.
func (r Rule) Evaluate(value float64) Result {
	status := Pass
	msg := fmt.Sprintf("%s (%.4g %s)", r.PassMsg, value, r.Unit)
	if !r.Comparison.passes(value, r.Threshold) {
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
		Comparison:  r.Comparison,
		Subsystem:   r.Subsystem,
		Status:      status,
		Message:     msg,
	}
}
