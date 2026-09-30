package report

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func baseRequest() Request {
	lat, lon := 37.7749, -122.4194
	return Request{
		UnitTag:     "ELV-003",
		InspectedAt: time.Date(2026, 4, 20, 14, 30, 0, 0, time.UTC),
		Technician:  "J. Smith",
		Results:     allPassResults(),
		Photos: []Photo{
			{
				RuleID:    "ASME-006",
				DataURI:   "data:image/jpeg;base64,/9j/testdata",
				Timestamp: "2026-04-20T14:31:00Z",
				Latitude:  &lat,
				Longitude: &lon,
			},
		},
	}
}

// render resolves req and renders it, the same path the Chrome renderer takes.
func render(req Request) (string, error) {
	in, err := Resolve(req)
	if err != nil {
		return "", err
	}
	return RenderHTML(in)
}

// ── Output validity ───────────────────────────────────────────────────────────

func TestRenderHTML_ReturnsNonEmptyString(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if html == "" {
		t.Fatal("expected non-empty HTML output")
	}
}

func TestRenderHTML_OutputIsValidHTMLDocument(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE declaration")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("missing closing </html> tag")
	}
}

// ── Unit tag ──────────────────────────────────────────────────────────────────

func TestRenderHTML_ContainsUnitTag(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "ELV-003") {
		t.Error("unit tag ELV-003 not found in rendered HTML")
	}
}

func TestRenderHTML_DifferentUnitTagAppearsInOutput(t *testing.T) {
	req := baseRequest()
	req.UnitTag = "ELV-001"
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "ELV-001") {
		t.Error("unit tag ELV-001 not found in rendered HTML")
	}
}

// ── Technician ────────────────────────────────────────────────────────────────

func TestRenderHTML_ContainsTechnicianName(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "J. Smith") {
		t.Error("technician name not found in rendered HTML")
	}
}

// ── Summary counts ────────────────────────────────────────────────────────────

func TestRenderHTML_CountsDerivedFromResults(t *testing.T) {
	req := baseRequest()
	req.Results[0].Status = StatusFail
	req.Results[1].Status = StatusUnknown
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		`<div class="num">18</div>`,
		`<div class="num">1</div>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered HTML missing %s", want)
		}
	}
}

// ── Overall status banner ─────────────────────────────────────────────────────

func TestRenderHTML_PassBannerWhenAllPass(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "ALL CHECKS PASSED") {
		t.Error("all-pass banner text not found")
	}
}

func TestRenderHTML_FailBannerWhenFailed(t *testing.T) {
	req := baseRequest()
	req.Results[0].Status = StatusFail
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "INSPECTION FAILED") {
		t.Error("fail banner text not found")
	}
}

func TestRenderHTML_FailBannerWhenOverriddenToFail(t *testing.T) {
	req := baseRequest()
	req.Overrides = map[string]Status{"ASME-006": StatusFail}
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "INSPECTION FAILED") {
		t.Error("fail banner not shown for a technician fail override")
	}
}

func TestRenderHTML_UnknownBannerWhenIncomplete(t *testing.T) {
	req := baseRequest()
	req.Results[0].Status = StatusUnknown
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "INCOMPLETE DATA") {
		t.Error("unknown banner text not found")
	}
}

// ── Rule results ──────────────────────────────────────────────────────────────

func TestRenderHTML_ContainsAllRuleIDs(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i <= 20; i++ {
		id := "ASME-" + zeroPad(i)
		if !strings.Contains(html, id) {
			t.Errorf("rule ID %s not found in rendered HTML", id)
		}
	}
}

func TestRenderHTML_FailRowHighlightedForFailedRule(t *testing.T) {
	req := baseRequest()
	req.Results[0].Status = StatusFail
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "row-fail") {
		t.Error("fail row CSS class not found for failed rule")
	}
}

func TestRenderHTML_OverrideTagShownWhenOverridden(t *testing.T) {
	req := baseRequest()
	req.Overrides = map[string]Status{"ASME-001": StatusFail}
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "(manual)") {
		t.Error("override tag not found for overridden rule")
	}
}

func TestRenderHTML_OverriddenRowShowsTechnicianAndTelemetryStatus(t *testing.T) {
	req := baseRequest()
	req.Results[0].Status = StatusFail
	req.Results[0].Message = "Motor current exceeds operational limit"
	req.Overrides = map[string]Status{"ASME-001": StatusPass}
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Technician: PASS · Telemetry: FAIL") {
		t.Error("overridden row does not show both technician and telemetry status")
	}
	if !strings.Contains(html, "Motor current exceeds operational limit") {
		t.Error("overridden row dropped the telemetry message")
	}
}

func TestRenderHTML_UnknownValueRendersDash(t *testing.T) {
	req := baseRequest()
	for i := range req.Results {
		req.Results[i].Unit = "zz"
	}
	req.Results[0].Status = StatusUnknown
	req.Results[0].Value = 0
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, "<td>0 zz</td>") {
		t.Error("unknown rule rendered a 0 reading")
	}
	if !strings.Contains(html, "<td>—</td>") {
		t.Error("unknown rule value not rendered as —")
	}
}

func TestRenderHTML_NoOverrideTagWhenNotOverridden(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, "(manual)") {
		t.Error("unexpected override tag in output with no overrides")
	}
}

// ── Subsystem groupings ───────────────────────────────────────────────────────

func TestRenderHTML_ContainsAllSubsystemHeadings(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, heading := range []string{"Motor", "Door Operator", "Brake System", "Safety Circuits"} {
		if !strings.Contains(html, heading) {
			t.Errorf("subsystem heading %q not found in rendered HTML", heading)
		}
	}
}

// ── Photos ────────────────────────────────────────────────────────────────────

func TestRenderHTML_PhotoSectionPresentWhenPhotosProvided(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Photo Evidence") {
		t.Error("photo evidence section not found")
	}
}

func TestRenderHTML_PhotoDataURIEmbeddedInOutput(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "data:image/jpeg;base64,/9j/testdata") {
		t.Error("photo data URI not found in rendered HTML")
	}
}

func TestRenderHTML_PhotoGPSCoordsInOutput(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "37.7749") {
		t.Error("GPS latitude not found in rendered HTML")
	}
}

func TestRenderHTML_NoPhotoSectionWhenNoPhotos(t *testing.T) {
	req := baseRequest()
	req.Photos = nil
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, "Photo Evidence") {
		t.Error("unexpected photo section rendered when no photos provided")
	}
}

func TestRenderHTML_NullGPSShowsNoGPSLabel(t *testing.T) {
	req := baseRequest()
	req.Photos = []Photo{
		{RuleID: "ASME-006", DataURI: "data:image/jpeg;base64,abc", Timestamp: "2026-04-20T14:31:00Z"},
	}
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "No GPS") {
		t.Error("expected 'No GPS' label when latitude/longitude are nil")
	}
}

// ── ASME A17.1 reference ──────────────────────────────────────────────────────

func TestRenderHTML_ContainsASMEReference(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "ASME A17.1") {
		t.Error("ASME A17.1 reference not found in rendered HTML")
	}
}

// ── Technician certification / signature ──────────────────────────────────────

func TestRenderHTML_NoSignatureSectionWhenAbsent(t *testing.T) {
	html, err := render(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, "Technician Certification") {
		t.Error("unexpected certification section when SignatureDataURI is empty")
	}
}

func TestRenderHTML_SignatureSectionWhenSignatureProvided(t *testing.T) {
	req := baseRequest()
	// Use PNG MIME type (no '+') so the HTML-encoded src matches the raw URI exactly.
	req.SignatureDataURI = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAE="
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Technician Certification") {
		t.Error("certification section not found when SignatureDataURI is provided")
	}
	if !strings.Contains(html, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAE=") {
		t.Error("signature data URI not embedded in certification section")
	}
}

func TestRenderHTML_SignatureSectionContainsTechnicianName(t *testing.T) {
	req := baseRequest()
	req.SignatureDataURI = "data:image/png;base64,iVBORw0K"
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// cert section includes technician name in the sig-line
	if !strings.Contains(html, "J. Smith") {
		t.Error("technician name not found in certification section")
	}
}

func TestRenderHTML_SignatureSVGPlusEncodedInHTML(t *testing.T) {
	// html/template HTML-encodes '+' as '&#43;' in attribute values.
	// Browsers decode this correctly when parsing, so PDF generation works.
	req := baseRequest()
	req.SignatureDataURI = "data:image/svg+xml;base64,PHN2Zz48L3N2Zz4="
	html, err := render(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Technician Certification") {
		t.Error("certification section not found")
	}
	// '+' encoded as '&#43;' — correct HTML for attribute values
	if !strings.Contains(html, "svg&#43;xml") {
		t.Error("expected '+' to be HTML-encoded as '&#43;' in attribute value")
	}
}

// ── helper ────────────────────────────────────────────────────────────────────

func zeroPad(n int) string {
	if n < 10 {
		return fmt.Sprintf("00%d", n)
	}
	if n < 100 {
		return fmt.Sprintf("0%d", n)
	}
	return fmt.Sprintf("%d", n)
}
