package report

import (
	"fmt"
	"time"
)

// Row is one rule as it appears in the signed report: the telemetry evidence
// plus the status the technician certifies.
type Row struct {
	RuleResult // telemetry evidence, unchanged

	// Effective is the certified status: the technician's override if one was
	// given, otherwise the telemetry status.
	Effective Status
	// Overridden is true when Effective came from a technician override.
	Overridden bool
}

// Inspection is the validated, fully derived outcome of an inspection.
// It is the only input the report renderer needs.
type Inspection struct {
	UnitTag          string
	InspectedAt      time.Time
	Technician       string
	Rows             []Row
	Summary          Summary
	Photos           []Photo
	SignatureDataURI string
}

// ValidationError reports a Request that cannot produce a report.
type ValidationError struct{ msg string }

func (e *ValidationError) Error() string { return e.msg }

func invalid(format string, args ...any) error {
	return &ValidationError{msg: fmt.Sprintf(format, args...)}
}

// Resolve applies the technician's overrides to the telemetry results and
// derives the summary. Every override must name a rule present in Results and
// be "pass" or "fail"; otherwise Resolve returns a *ValidationError.
func Resolve(req Request) (Inspection, error) {
	if req.UnitTag == "" {
		return Inspection{}, invalid("unit_tag is required")
	}
	if len(req.Results) == 0 {
		return Inspection{}, invalid("results must not be empty")
	}

	known := make(map[string]bool, len(req.Results))
	for _, r := range req.Results {
		known[r.RuleID] = true
	}
	for id, s := range req.Overrides {
		if !known[id] {
			return Inspection{}, invalid("override for unknown rule %q", id)
		}
		if s != StatusPass && s != StatusFail {
			return Inspection{}, invalid("override for %s must be \"pass\" or \"fail\", got %q", id, s)
		}
	}

	rows := make([]Row, len(req.Results))
	for i, r := range req.Results {
		row := Row{RuleResult: r, Effective: r.Status}
		if s, ok := req.Overrides[r.RuleID]; ok {
			row.Effective = s
			row.Overridden = true
		}
		rows[i] = row
	}

	return Inspection{
		UnitTag:          req.UnitTag,
		InspectedAt:      req.InspectedAt,
		Technician:       req.Technician,
		Rows:             rows,
		Summary:          summarise(rows),
		Photos:           req.Photos,
		SignatureDataURI: req.SignatureDataURI,
	}, nil
}

// summarise tallies effective statuses. Overall is "fail" if any rule failed,
// "unknown" if any rule is unknown (and none failed), and "pass" only when
// every rule passes — the same semantics as the compliance engine.
func summarise(rows []Row) Summary {
	var s Summary
	for _, r := range rows {
		switch r.Effective {
		case StatusPass:
			s.Pass++
		case StatusFail:
			s.Fail++
		default:
			s.Unknown++
		}
	}
	switch {
	case s.Fail > 0:
		s.Overall = StatusFail
	case s.Unknown > 0:
		s.Overall = StatusUnknown
	default:
		s.Overall = StatusPass
	}
	return s
}
