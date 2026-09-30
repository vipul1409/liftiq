package report

import (
	"context"
	"fmt"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// PDFRenderer converts a resolved Inspection into a PDF byte slice.
// Abstracting behind an interface keeps the HTTP handler testable without
// requiring a live Chrome installation.
type PDFRenderer interface {
	Render(ctx context.Context, in Inspection) ([]byte, error)
}

// ChromePDFRenderer renders PDFs using a headless Chrome instance via chromedp.
// Chrome or Chromium must be installed and discoverable on PATH.
type ChromePDFRenderer struct{}

// Render renders the report HTML to PDF via chromedp.
func (r *ChromePDFRenderer) Render(ctx context.Context, in Inspection) ([]byte, error) {
	html, err := RenderHTML(in)
	if err != nil {
		return nil, fmt.Errorf("render html: %w", err)
	}

	// Write HTML to a temporary file so Chrome can load it via a file:// URL.
	// This avoids URL-length limits that affect data: URI navigation.
	tmp, err := os.CreateTemp("", "liftiq-report-*.html")
	if err != nil {
		return nil, fmt.Errorf("create temp html file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(html); err != nil {
		return nil, fmt.Errorf("write temp html file: %w", err)
	}
	tmp.Close()

	allocCtx, allocCancel := chromedp.NewContext(ctx)
	defer allocCancel()

	var pdfBuf []byte
	if err := chromedp.Run(allocCtx,
		chromedp.Navigate("file://"+tmp.Name()),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("chromedp pdf: %w", err)
	}
	return pdfBuf, nil
}
