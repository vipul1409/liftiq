package report

import (
	"errors"
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
			req := Request{UnitTag: "ELV-003", Results: allPassResults(), Overrides: tc.overrides}
			for i, r := range req.Results {
				if s, ok := tc.telemetry[r.RuleID]; ok {
					req.Results[i].Status = s
				}
			}

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
			r.Overrides = map[string]Status{"ASME-999": StatusPass}
		}},
		{"override to unknown", func(r *Request) {
			r.Overrides = map[string]Status{"ASME-001": StatusUnknown}
		}},
		{"override to garbage", func(r *Request) {
			r.Overrides = map[string]Status{"ASME-001": "maybe"}
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
