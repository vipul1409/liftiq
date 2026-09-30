package report

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"time"
)

// safeURL wraps a string as template.URL, bypassing Go's URL safety filter for
// data: URIs used in img src attributes.
func toTemplateURL(s string) template.URL { return template.URL(s) }

// formatFloat prints v in shortest decimal form, never in exponent notation.
func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// formatLimit renders a rule's pass condition, e.g. "≤ 135 N", "< 500000 trips",
// or "= OK" for boolean safety fields.
func formatLimit(r Row) string {
	switch r.Comparison {
	case "equals":
		if r.Threshold == 1 {
			return "= OK"
		}
		return "= " + formatFloat(r.Threshold) + " " + r.Unit
	case "below":
		return "< " + formatFloat(r.Threshold) + " " + r.Unit
	case "at_most":
		return "≤ " + formatFloat(r.Threshold) + " " + r.Unit
	default:
		return formatFloat(r.Threshold) + " " + r.Unit
	}
}

// templateData is the full data model passed to the HTML template.
type templateData struct {
	UnitTag          string
	InspectedAt      string
	GeneratedAt      string
	Technician       string
	Summary          Summary
	Groups           []SubsystemGroup
	PhotosByRule     map[string][]Photo
	HasPhotos        bool
	SignatureDataURI template.URL // pre-typed to bypass html/template URL filtering
}

// RenderHTML renders the inspection report as an HTML string.
// This is the only function that needs testing without Chrome.
func RenderHTML(in Inspection) (string, error) {
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"statusClass": func(s Status) string {
			switch s {
			case StatusPass:
				return "pass"
			case StatusFail:
				return "fail"
			default:
				return "unknown"
			}
		},
		"statusLabel": func(s Status) string {
			switch s {
			case StatusPass:
				return "PASS"
			case StatusFail:
				return "FAIL"
			default:
				return "N/A"
			}
		},
		"formatFloat": formatFloat,
		"formatLimit": formatLimit,
		"formatTime": func(t time.Time) string {
			return t.UTC().Format("2006-01-02 15:04:05 UTC")
		},
		"gpsLabel": func(lat, lon *float64) string {
			if lat == nil || lon == nil {
				return "No GPS"
			}
			return fmt.Sprintf("%.4f, %.4f", *lat, *lon)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
	}).Parse(reportHTML)
	if err != nil {
		return "", fmt.Errorf("parse report template: %w", err)
	}

	photosByRule := PhotosByRule(in.Photos)
	data := templateData{
		UnitTag:          in.UnitTag,
		InspectedAt:      in.InspectedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
		GeneratedAt:      time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		Technician:       in.Technician,
		Summary:          in.Summary,
		Groups:           GroupBySubsystem(in.Rows),
		PhotosByRule:     photosByRule,
		HasPhotos:        len(in.Photos) > 0,
		SignatureDataURI: toTemplateURL(in.SignatureDataURI),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute report template: %w", err)
	}
	return buf.String(), nil
}

const reportHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>LiftIQ Inspection Report — {{.UnitTag}}</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, Arial, sans-serif; font-size: 11pt; color: #111827; background: #fff; }

  @page { size: A4; margin: 18mm 15mm; }
  @media print { .no-print { display: none; } }

  /* ── Header ── */
  .header { display: flex; justify-content: space-between; align-items: flex-start; padding-bottom: 12px; border-bottom: 3px solid #1e3a8a; margin-bottom: 16px; }
  .brand { font-size: 22pt; font-weight: 900; color: #1e3a8a; letter-spacing: -0.5px; }
  .brand span { color: #2563eb; }
  .report-title { font-size: 10pt; color: #6b7280; text-align: right; line-height: 1.6; }

  /* ── Status banner ── */
  .banner { border-radius: 8px; padding: 16px 20px; margin-bottom: 18px; display: flex; justify-content: space-between; align-items: center; }
  .banner.pass  { background: #dcfce7; border: 1.5px solid #86efac; }
  .banner.fail  { background: #fee2e2; border: 1.5px solid #fca5a5; }
  .banner.unknown { background: #f3f4f6; border: 1.5px solid #d1d5db; }
  .banner-left h2 { font-size: 16pt; font-weight: 800; }
  .banner-left h2.pass    { color: #15803d; }
  .banner-left h2.fail    { color: #b91c1c; }
  .banner-left h2.unknown { color: #374151; }
  .banner-left p { font-size: 10pt; color: #4b5563; margin-top: 4px; }
  .banner-counts { display: flex; gap: 20px; }
  .count-item { text-align: center; }
  .count-item .num { font-size: 22pt; font-weight: 800; line-height: 1; }
  .count-item .lbl { font-size: 8pt; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; color: #6b7280; }
  .count-item.c-pass   .num { color: #15803d; }
  .count-item.c-fail   .num { color: #b91c1c; }
  .count-item.c-unknown .num { color: #6b7280; }

  /* ── Metadata table ── */
  .meta-table { width: 100%; border-collapse: collapse; margin-bottom: 20px; font-size: 10pt; }
  .meta-table td { padding: 5px 10px; border: 1px solid #e5e7eb; }
  .meta-table td:first-child { font-weight: 600; color: #374151; background: #f9fafb; width: 38%; }

  /* ── Section heading ── */
  .section-heading { font-size: 9pt; font-weight: 700; text-transform: uppercase; letter-spacing: 0.8px; color: #374151; background: #f3f4f6; padding: 6px 10px; border-top: 1px solid #e5e7eb; border-bottom: 1px solid #e5e7eb; margin-bottom: 0; }

  /* ── Rules table ── */
  table.rules { width: 100%; border-collapse: collapse; font-size: 9.5pt; margin-bottom: 4px; }
  table.rules th { background: #f9fafb; font-weight: 700; text-align: left; padding: 6px 8px; border-bottom: 2px solid #e5e7eb; color: #374151; font-size: 8.5pt; text-transform: uppercase; letter-spacing: 0.5px; }
  table.rules td { padding: 7px 8px; border-bottom: 1px solid #f3f4f6; vertical-align: top; }
  table.rules tr:last-child td { border-bottom: none; }
  table.rules tr.row-fail td { background: #fff5f5; }

  /* Status badge */
  .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 8pt; font-weight: 700; letter-spacing: 0.3px; }
  .badge.pass    { background: #dcfce7; color: #15803d; }
  .badge.fail    { background: #fee2e2; color: #b91c1c; }
  .badge.unknown { background: #f3f4f6; color: #6b7280; }

  .rule-id   { font-size: 8pt; color: #9ca3af; font-weight: 600; white-space: nowrap; }
  .rule-desc { font-weight: 600; color: #111827; }
  .rule-ref  { font-size: 8pt; color: #6b7280; font-style: italic; margin-top: 2px; }
  .rule-msg  { font-size: 8.5pt; color: #4b5563; margin-top: 2px; }
  .override-tag { font-size: 7.5pt; color: #7c3aed; margin-left: 4px; }

  /* ── Photos ── */
  .photos-section { margin-top: 20px; }
  .photos-section h3 { font-size: 11pt; font-weight: 700; color: #111827; margin-bottom: 12px; border-bottom: 2px solid #e5e7eb; padding-bottom: 6px; }
  .photo-group { margin-bottom: 16px; }
  .photo-group-title { font-size: 9pt; font-weight: 700; color: #374151; background: #f3f4f6; padding: 4px 8px; border-left: 3px solid #2563eb; margin-bottom: 8px; }
  .photo-grid { display: flex; flex-wrap: wrap; gap: 10px; }
  .photo-item { width: 120px; }
  .photo-item img { width: 120px; height: 90px; object-fit: cover; border-radius: 4px; border: 1px solid #e5e7eb; }
  .photo-meta { font-size: 7.5pt; color: #9ca3af; margin-top: 3px; line-height: 1.4; }

  /* ── Certification ── */
  .cert-section { margin-top: 20px; padding: 16px; border: 1.5px solid #e5e7eb; border-radius: 8px; }
  .cert-section h3 { font-size: 10pt; font-weight: 700; color: #374151; margin-bottom: 6px; }
  .cert-section p { font-size: 9pt; color: #6b7280; margin-bottom: 12px; }
  .sig-box { display: flex; flex-direction: column; align-items: flex-start; }
  .sig-img { height: 72px; max-width: 280px; border-bottom: 1.5px solid #374151; margin-bottom: 4px; }
  .sig-line { font-size: 8.5pt; color: #374151; font-weight: 600; }

  /* ── Footer ── */
  .footer { margin-top: 24px; padding-top: 10px; border-top: 1px solid #e5e7eb; font-size: 8pt; color: #9ca3af; display: flex; justify-content: space-between; }
</style>
</head>
<body>

<!-- Header -->
<div class="header">
  <div class="brand">Lift<span>IQ</span></div>
  <div class="report-title">
    ELEVATOR INSPECTION REPORT<br>
    ASME A17.1 Safety Code<br>
    Generated {{.GeneratedAt}}
  </div>
</div>

<!-- Status Banner -->
{{$overall := .Summary.Overall}}
<div class="banner {{statusClass $overall}}">
  <div class="banner-left">
    <h2 class="{{statusClass $overall}}">
      {{if eq $overall "pass"}}ALL CHECKS PASSED
      {{else if eq $overall "fail"}}INSPECTION FAILED
      {{else}}INCOMPLETE DATA{{end}}
    </h2>
    <p>Unit {{.UnitTag}} · Inspected {{.InspectedAt}}</p>
  </div>
  <div class="banner-counts">
    <div class="count-item c-pass">
      <div class="num">{{.Summary.Pass}}</div>
      <div class="lbl">Pass</div>
    </div>
    <div class="count-item c-fail">
      <div class="num">{{.Summary.Fail}}</div>
      <div class="lbl">Fail</div>
    </div>
    <div class="count-item c-unknown">
      <div class="num">{{.Summary.Unknown}}</div>
      <div class="lbl">N/A</div>
    </div>
  </div>
</div>

<!-- Inspection Metadata -->
<table class="meta-table">
  <tr><td>Unit Tag</td><td>{{.UnitTag}}</td></tr>
  <tr><td>Inspection Date</td><td>{{.InspectedAt}}</td></tr>
  <tr><td>Technician</td><td>{{.Technician}}</td></tr>
  <tr><td>Standard</td><td>ASME A17.1 Safety Code for Elevators and Escalators</td></tr>
  <tr><td>Overall Result</td><td><span class="badge {{statusClass $overall}}">{{statusLabel $overall}}</span></td></tr>
</table>

<!-- Compliance Checklist -->
{{range .Groups}}
<div class="section-heading">{{.Title}}</div>
<table class="rules">
  <thead>
    <tr>
      <th style="width:8%">ID</th>
      <th style="width:36%">Rule</th>
      <th style="width:10%">Status</th>
      <th style="width:14%">Value</th>
      <th style="width:14%">Limit</th>
      <th style="width:18%">Assessment</th>
    </tr>
  </thead>
  <tbody>
  {{range .Results}}
  <tr class="{{if eq .Effective "fail"}}row-fail{{end}}">
    <td class="rule-id">{{.RuleID}}</td>
    <td>
      <div class="rule-desc">{{.Description}}{{if .Overridden}}<span class="override-tag">(manual)</span>{{end}}</div>
      <div class="rule-ref">{{.ASMERef}}</div>
    </td>
    <td><span class="badge {{statusClass .Effective}}">{{statusLabel .Effective}}</span></td>
    <td>{{if eq .Status "unknown"}}—{{else}}{{formatFloat .Value}} {{.Unit}}{{end}}</td>
    <td>{{formatLimit .}}</td>
    <td class="rule-msg">{{if .Overridden}}<span class="override-tag">Technician: {{statusLabel .Effective}} · Telemetry: {{statusLabel .Status}}</span><br>{{end}}{{.Message}}</td>
  </tr>
  {{end}}
  </tbody>
</table>
{{end}}

<!-- Photo Evidence -->
{{if .HasPhotos}}
<div class="photos-section">
  <h3>Photo Evidence</h3>
  {{range .Groups}}
    {{$title := .Title}}
    {{range .Results}}
      {{$ruleID := .RuleID}}
      {{$photos := index $.PhotosByRule $ruleID}}
      {{if $photos}}
      <div class="photo-group">
        <div class="photo-group-title">{{$ruleID}} — {{.Description}}</div>
        <div class="photo-grid">
          {{range $photos}}
          <div class="photo-item">
            <img src="{{safeURL .DataURI}}" alt="Evidence for {{.RuleID}}">
            <div class="photo-meta">
              {{.Timestamp}}<br>
              {{gpsLabel .Latitude .Longitude}}
            </div>
          </div>
          {{end}}
        </div>
      </div>
      {{end}}
    {{end}}
  {{end}}
</div>
{{end}}

<!-- Technician Certification -->
{{if .SignatureDataURI}}
<div class="cert-section">
  <h3>Technician Certification</h3>
  <p>I certify that this inspection was performed in accordance with ASME A17.1 Safety Code for Elevators and Escalators.</p>
  <div class="sig-box">
    <img src="{{.SignatureDataURI}}" class="sig-img" alt="Technician signature">
    <div class="sig-line">{{.Technician}} · {{.InspectedAt}}</div>
  </div>
</div>
{{end}}

<!-- Footer -->
<div class="footer">
  <span>LiftIQ · Automated Elevator Compliance Platform</span>
  <span>ASME A17.1 Safety Code for Elevators and Escalators</span>
</div>

</body>
</html>`
