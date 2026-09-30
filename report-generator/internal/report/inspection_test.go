package report

import (
	"errors"
	"strings"
	"testing"
)

func TestResolve_AppliesOverridesAndDerivesSummary(t *testing.T) {
	cases := []struct {
		name      string
		telemetry map[string]Status // rule ID → telemetry status; others pass
		overrides map[string]Status
		want      Summary
		wantRow   map[string]Status // rule ID → expected effective status
	}{
		{
			name: "no overrides, all pass",
			want: Summary{Pass: 20, Overall: StatusPass},
		},
		{
			name:      "technician fails a telemetry pass",
			overrides: map[string]Status{"ASME-006": StatusFail},
			want:      Summary{Pass: 19, Fail: 1, Overall: StatusFail},
			wantRow:   map[string]Status{"ASME-006": StatusFail},
		},
		{
			name:      "technician passes a telemetry fail",
			telemetry: map[string]Status{"ASME-011": StatusFail},
			overrides: map[string]Status{"ASME-011": StatusPass},
			want:      Summary{Pass: 20, Overall: StatusPass},
			wantRow:   map[string]Status{"ASME-011": StatusPass},
		},
		{
			name:      "technician resolves an unknown",
			telemetry: map[string]Status{"ASME-014": StatusUnknown, "ASME-015": StatusUnknown},
			overrides: map[string]Status{"ASME-014": StatusPass},
			want:      Summary{Pass: 19, Unknown: 1, Overall: StatusUnknown},
			wantRow:   map[string]Status{"ASME-014": StatusPass, "ASME-015": StatusUnknown},
		},
		{
			name:      "fail outranks unknown",
			telemetry: map[string]Status{"ASME-014": StatusUnknown},
			overrides: map[string]Status{"ASME-001": StatusFail},
			want:      Summary{Pass: 18, Fail: 1, Unknown: 1, Overall: StatusFail},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := Request{UnitTag: "ELV-003", Results: allPassResults()}
			for i, r := range req.Results {
				if s, ok := tc.telemetry[r.RuleID]; ok {
					req.Results[i].Status = s
				}
			}
			req.Overrides = anchored(req.Results, tc.overrides)

			in, err := Resolve(req)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if in.Summary != tc.want {
				t.Errorf("summary = %+v, want %+v", in.Summary, tc.want)
			}
			for _, row := range in.Rows {
				if want, ok := tc.wantRow[row.RuleID]; ok && row.Effective != want {
					t.Errorf("%s effective = %q, want %q", row.RuleID, row.Effective, want)
				}
				_, overridden := tc.overrides[row.RuleID]
				if row.Overridden != overridden {
					t.Errorf("%s overridden = %v, want %v", row.RuleID, row.Overridden, overridden)
				}
				if s, ok := tc.telemetry[row.RuleID]; ok && row.Status != s {
					t.Errorf("%s telemetry status changed to %q, want %q", row.RuleID, row.Status, s)
				}
			}
		})
	}
}

func TestResolve_RejectsInvalidRequests(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Request)
	}{
		{"missing unit tag", func(r *Request) { r.UnitTag = "" }},
		{"empty results", func(r *Request) { r.Results = nil }},
		{"override for unknown rule", func(r *Request) {
			r.Overrides = map[string]Override{"ASME-999": {Status: StatusPass, Against: StatusPass}}
		}},
		{"override to unknown", func(r *Request) {
			r.Overrides = map[string]Override{"ASME-001": {Status: StatusUnknown, Against: StatusPass}}
		}},
		{"override to garbage", func(r *Request) {
			r.Overrides = map[string]Override{"ASME-001": {Status: "maybe", Against: StatusPass}}
		}},
		{"override made against a different result", func(r *Request) {
			// Made while ASME-001 was unknown; the submitted result is now pass.
			r.Overrides = map[string]Override{"ASME-001": {Status: StatusFail, Against: StatusUnknown}}
		}},
		{"override missing its anchor", func(r *Request) {
			r.Overrides = map[string]Override{"ASME-001": {Status: StatusFail}}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := Request{UnitTag: "ELV-003", Results: allPassResults()}
			tc.mutate(&req)
			_, err := Resolve(req)
			var verr *ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("err = %v, want *ValidationError", err)
			}
		})
	}
}

// anchored builds Overrides made against each rule's current result status,
// i.e. overrides the technician has confirmed against this exact evidence.
func anchored(results []RuleResult, calls map[string]Status) map[string]Override {
	if calls == nil {
		return nil
	}
	current := make(map[string]Status, len(results))
	for _, r := range results {
		current[r.RuleID] = r.Status
	}
	out := make(map[string]Override, len(calls))
	for id, s := range calls {
		out[id] = Override{Status: s, Against: current[id]}
	}
	return out
}

func TestResolve_RejectsOverrideAgainstStaleResultWithReason(t *testing.T) {
	req := Request{UnitTag: "ELV-003", Results: allPassResults()}
	req.Overrides = map[string]Override{"ASME-011": {Status: StatusFail, Against: StatusUnknown}}
	_, err := Resolve(req)
	if err == nil {
		t.Fatal("want error for override made against a different result")
	}
	for _, want := range []string{"ASME-011", "unknown", "pass", "re-confirm"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}
