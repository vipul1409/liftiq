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
		Summary: Summary{
			Pass: 20, Fail: 0, Unknown: 0, Overall: StatusPass,
		},
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

// ── Output validity ───────────────────────────────────────────────────────────

func TestRenderHTML_ReturnsNonEmptyString(t *testing.T) {
	html, err := RenderHTML(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if html == "" {
		t.Fatal("expected non-empty HTML output")
	}
}

func TestRenderHTML_OutputIsValidHTMLDocument(t *testing.T) {
	html, err := RenderHTML(baseRequest())
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
	html, err := RenderHTML(baseRequest())
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
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "ELV-001") {
		t.Error("unit tag ELV-001 not found in rendered HTML")
	}
}

// ── Technician ────────────────────────────────────────────────────────────────

func TestRenderHTML_ContainsTechnicianName(t *testing.T) {
	html, err := RenderHTML(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "J. Smith") {
		t.Error("technician name not found in rendered HTML")
	}
}

// ── Summary counts ────────────────────────────────────────────────────────────

func TestRenderHTML_ContainsPassCount(t *testing.T) {
	req := baseRequest()
	req.Summary = Summary{Pass: 18, Fail: 1, Unknown: 1, Overall: StatusFail}
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "18") {
		t.Error("pass count 18 not found in rendered HTML")
	}
}

func TestRenderHTML_ContainsFailCount(t *testing.T) {
	req := baseRequest()
	req.Summary = Summary{Pass: 18, Fail: 2, Unknown: 0, Overall: StatusFail}
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "2") {
		t.Error("fail count not found in rendered HTML")
	}
}

// ── Overall status banner ─────────────────────────────────────────────────────

func TestRenderHTML_PassBannerWhenAllPass(t *testing.T) {
	req := baseRequest()
	req.Summary.Overall = StatusPass
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "ALL CHECKS PASSED") {
		t.Error("all-pass banner text not found")
	}
}

func TestRenderHTML_FailBannerWhenFailed(t *testing.T) {
	req := baseRequest()
	req.Summary.Overall = StatusFail
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "INSPECTION FAILED") {
		t.Error("fail banner text not found")
	}
}

func TestRenderHTML_UnknownBannerWhenIncomplete(t *testing.T) {
	req := baseRequest()
	req.Summary.Overall = StatusUnknown
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "INCOMPLETE DATA") {
		t.Error("unknown banner text not found")
	}
}

// ── Rule results ──────────────────────────────────────────────────────────────

func TestRenderHTML_ContainsAllRuleIDs(t *testing.T) {
	html, err := RenderHTML(baseRequest())
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
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "row-fail") {
		t.Error("fail row CSS class not found for failed rule")
	}
}

func TestRenderHTML_OverrideTagShownWhenOverridden(t *testing.T) {
	req := baseRequest()
	req.Results[0].Overridden = true
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "(manual)") {
		t.Error("override tag not found for overridden rule")
	}
}

func TestRenderHTML_NoOverrideTagWhenNotOverridden(t *testing.T) {
	html, err := RenderHTML(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(html, "(manual)") {
		t.Error("unexpected override tag in output with no overrides")
	}
}

// ── Subsystem groupings ───────────────────────────────────────────────────────

func TestRenderHTML_ContainsAllSubsystemHeadings(t *testing.T) {
	html, err := RenderHTML(baseRequest())
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
	html, err := RenderHTML(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "Photo Evidence") {
		t.Error("photo evidence section not found")
	}
}

func TestRenderHTML_PhotoDataURIEmbeddedInOutput(t *testing.T) {
	html, err := RenderHTML(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "data:image/jpeg;base64,/9j/testdata") {
		t.Error("photo data URI not found in rendered HTML")
	}
}

func TestRenderHTML_PhotoGPSCoordsInOutput(t *testing.T) {
	html, err := RenderHTML(baseRequest())
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
	html, err := RenderHTML(req)
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
	html, err := RenderHTML(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "No GPS") {
		t.Error("expected 'No GPS' label when latitude/longitude are nil")
	}
}

// ── ASME A17.1 reference ──────────────────────────────────────────────────────

func TestRenderHTML_ContainsASMEReference(t *testing.T) {
	html, err := RenderHTML(baseRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(html, "ASME A17.1") {
		t.Error("ASME A17.1 reference not found in rendered HTML")
	}
}

// ── Technician certification / signature ──────────────────────────────────────

func TestRenderHTML_NoSignatureSectionWhenAbsent(t *testing.T) {
	html, err := RenderHTML(baseRequest())
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
	html, err := RenderHTML(req)
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
	html, err := RenderHTML(req)
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
	html, err := RenderHTML(req)
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
