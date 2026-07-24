package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos7ags/folio/document"
	"github.com/carlos7ags/folio/layout"
	"github.com/carlos7ags/folio/tmpl"
)

// Application is the view model that drives cv.pdf and cover.pdf
// generation. It merges the applicant's CV with the target job and the
// cover letter text into the single shape the templates render from.
//
// TODO: read this from {id}/application.json once that file format
// exists. For now loadApplication builds one in memory with placeholder
// data so the pdf pipeline can be exercised end to end.
type Application struct {
	ID string

	Applicant ApplicantInfo
	Company   CompanyInfo
	JobTitle  string

	CV          CV
	CoverLetter CoverLetter
}

// ApplicantInfo is the sender side of an application: the person applying.
type ApplicantInfo struct {
	Name   string
	Street string
	City   string
	Email  string
	Phone  string
}

// CompanyInfo is the recipient side of an application.
type CompanyInfo struct {
	Name    string
	Street  string
	City    string
	Contact string // e.g. "Ms. Jane Doe", empty if unknown
}

// CoverLetter holds the free-text parts of the cover letter.
type CoverLetter struct {
	PlaceAndDate string
	Salutation   string
	Paragraphs   []string
	Closing      string
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	ids := fs.Args()

	if len(ids) == 0 {
		return fmt.Errorf("usage: applyme generate <id> [<id> ...]")
	}

	var failed []string
	for _, id := range ids {
		if err := generateApplication(id); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", id, err)
			failed = append(failed, id)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to generate %d application(s): %v", len(failed), failed)
	}
	return nil
}

func generateApplication(id string) error {
	app := loadApplication(id)

	dir := filepath.Join("applications", id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	cvPath := filepath.Join(dir, "cv.pdf")
	if err := renderPDF(cvTemplateHTML, app, cvPath); err != nil {
		return fmt.Errorf("rendering cv.pdf: %w", err)
	}
	fmt.Printf("created %s\n", cvPath)

	coverPath := filepath.Join(dir, "cover.pdf")
	if err := renderPDF(coverTemplateHTML, app, coverPath); err != nil {
		return fmt.Errorf("rendering cover.pdf: %w", err)
	}
	fmt.Printf("created %s\n", coverPath)

	return nil
}

// renderPDF executes an html/template against app and writes the result
// as a PDF using Folio's HTML-to-PDF pipeline.
func renderPDF(templateHTML string, app Application, outPath string) error {
	opts := &tmpl.Options{
		PageSize: document.PageSizeA4,
		Margins:  &layout.Margins{Top: 56, Right: 56, Bottom: 56, Left: 56},
	}

	doc, err := tmpl.RenderDocument(templateHTML, app, opts)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}
	if err := doc.Save(outPath); err != nil {
		return fmt.Errorf("saving %s: %w", outPath, err)
	}
	return nil
}

// loadApplication stands in for reading {id}/application.json from disk.
// It returns placeholder data shaped like a real application so the
// generate pipeline can be built and exercised now.
func loadApplication(id string) Application {
	return Application{
		ID: id,

		Applicant: ApplicantInfo{
			Name:   "Jane Doe",
			Street: "Musterstrasse 12",
			City:   "8000 Zürich",
			Email:  "jane.doe@example.com",
			Phone:  "+41 79 123 45 67",
		},

		Company: CompanyInfo{
			Name:    "Zur Rose Suisse AG",
			Street:  "Walzmühlestrasse 60",
			City:    "8500 Frauenfeld",
			Contact: "",
		},
		JobTitle: "Solution Architect - D365 & Power Platform",

		CV: CV{
			Name:    "Jane Doe",
			Email:   "jane.doe@example.com",
			Phone:   "+41 79 123 45 67",
			Summary: "Solution architect with 8+ years of experience designing scalable CRM and low-code platform solutions.",
			Experience: []string{
				"Solution Architect, Example AG (2021-present) - owns platform architecture for CRM and Power Platform.",
				"Software Engineer, Sample GmbH (2017-2021) - built integrations and automations on Dynamics 365.",
			},
			Education: []string{
				"BSc Computer Science, ETH Zürich (2013-2017)",
			},
			Skills: []string{"Dynamics 365", "Power Platform", "Azure Functions", "Solution Design", "Go", "TypeScript"},
		},

		CoverLetter: CoverLetter{
			PlaceAndDate: "Zürich, 24. Juli 2026",
			Salutation:   "Dear Hiring Team,",
			Paragraphs: []string{
				"I am writing to apply for the Solution Architect position on your CRM and Power Platform team. Your focus on turning requirements into consistent, scalable architectures matches exactly how I like to work.",
				"In my current role I own the solution design for our CRM platform end to end - from data model to automation to integration - and I would bring that same structured, pragmatic approach to your team.",
				"I would welcome the opportunity to discuss how my experience fits your plans.",
			},
			Closing: "Kind regards,",
		},
	}
}

const cvTemplateHTML = `<!DOCTYPE html>
<html>
<head><style>
  body { font-family: sans-serif; font-size: 11px; color: #1a1a1a; }
  h1 { font-size: 24px; margin-bottom: 4px; }
  .contact { color: #555555; margin-bottom: 20px; }
  h2 { font-size: 14px; margin-top: 18px; margin-bottom: 6px; border-bottom: 1px solid #cccccc; padding-bottom: 2px; }
  p { margin: 0 0 8px 0; line-height: 1.4; }
  ul { margin: 0 0 8px 0; padding-left: 18px; }
  li { margin-bottom: 4px; line-height: 1.4; }
  .skills { line-height: 1.6; }
</style></head>
<body>
  <h1>{{.CV.Name}}</h1>
  <p class="contact">{{.CV.Email}} &middot; {{.CV.Phone}}</p>

  {{if .CV.Summary}}
  <h2>Summary</h2>
  <p>{{.CV.Summary}}</p>
  {{end}}

  {{if .CV.Experience}}
  <h2>Experience</h2>
  <ul>
    {{range .CV.Experience}}<li>{{.}}</li>{{end}}
  </ul>
  {{end}}

  {{if .CV.Education}}
  <h2>Education</h2>
  <ul>
    {{range .CV.Education}}<li>{{.}}</li>{{end}}
  </ul>
  {{end}}

  {{if .CV.Skills}}
  <h2>Skills</h2>
  <p class="skills">{{range $i, $s := .CV.Skills}}{{if $i}} &middot; {{end}}{{$s}}{{end}}</p>
  {{end}}
</body>
</html>`

const coverTemplateHTML = `<!DOCTYPE html>
<html>
<head><style>
  body { font-family: sans-serif; font-size: 11px; color: #1a1a1a; line-height: 1.5; }
  .sender { margin-bottom: 40px; }
  .recipient { margin-bottom: 40px; }
  .date { text-align: right; margin-bottom: 40px; }
  .subject { font-weight: bold; margin-bottom: 16px; }
  p { margin: 0 0 12px 0; }
  .closing { margin-top: 24px; }
</style></head>
<body>
  <div class="sender">
    {{.Applicant.Name}}<br>
    {{.Applicant.Street}}<br>
    {{.Applicant.City}}<br>
    {{.Applicant.Email}} &middot; {{.Applicant.Phone}}
  </div>

  <div class="recipient">
    {{.Company.Name}}<br>
    {{if .Company.Contact}}{{.Company.Contact}}<br>{{end}}
    {{.Company.Street}}<br>
    {{.Company.City}}
  </div>

  <div class="date">{{.CoverLetter.PlaceAndDate}}</div>

  <p class="subject">Application: {{.JobTitle}}</p>

  <p>{{.CoverLetter.Salutation}}</p>

  {{range .CoverLetter.Paragraphs}}<p>{{.}}</p>
  {{end}}

  <p class="closing">{{.CoverLetter.Closing}}<br>{{.Applicant.Name}}</p>
</body>
</html>`
