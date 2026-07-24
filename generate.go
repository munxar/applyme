package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

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

	Job Job

	CV          CV
	CoverLetter CoverLetter
}

// Job is the position being applied for, at a given company.
type Job struct {
	Title   string
	Company CompanyInfo
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

	cfg, err := loadConfig(fs, args)
	if err != nil {
		return err
	}
	ids := fs.Args()

	if len(ids) == 0 {
		return fmt.Errorf("usage: applyme generate <id> [<id> ...]")
	}

	var failed []string
	for _, id := range ids {
		if err := generateApplication(cfg, id); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", id, err)
			failed = append(failed, id)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to generate %d application(s): %v", len(failed), failed)
	}
	return nil
}

func generateApplication(cfg Config, id string) error {
	app := loadApplication(id)

	dir := filepath.Join(cfg.ApplicationsDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	cvFileName, err := renderFileName(cfg.CVFileName, defaultCVFileName, app)
	if err != nil {
		return fmt.Errorf("rendering cv file name: %w", err)
	}
	cvPath := filepath.Join(dir, cvFileName)
	if err := renderPDF(cvTemplateHTML, app, cvPath); err != nil {
		return fmt.Errorf("rendering %s: %w", cvFileName, err)
	}
	fmt.Printf("created %s\n", cvPath)

	coverFileName, err := renderFileName(cfg.CoverFileName, defaultCoverFileName, app)
	if err != nil {
		return fmt.Errorf("rendering cover file name: %w", err)
	}
	coverPath := filepath.Join(dir, coverFileName)
	if err := renderPDF(coverTemplateHTML, app, coverPath); err != nil {
		return fmt.Errorf("rendering %s: %w", coverFileName, err)
	}
	fmt.Printf("created %s\n", coverPath)

	return nil
}

// renderFileName executes nameTemplate (a go template, e.g.
// "{{.CV.Personal.LastName}} {{.CV.Personal.FirstName}} - cv.pdf") against
// app to produce a PDF file name. If nameTemplate is empty, fallback is
// used as-is.
func renderFileName(nameTemplate, fallback string, app Application) (string, error) {
	if nameTemplate == "" {
		return fallback, nil
	}

	tmpl, err := template.New("filename").Parse(nameTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing file name template %q: %w", nameTemplate, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, app); err != nil {
		return "", fmt.Errorf("executing file name template %q: %w", nameTemplate, err)
	}
	return buf.String(), nil
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

		Job: Job{
			Title: "Solution Architect - D365 & Power Platform",
			Company: CompanyInfo{
				Name:    "Zur Rose Suisse AG",
				Street:  "Walzmühlestrasse 60",
				City:    "8500 Frauenfeld",
				Contact: "",
			},
		},

		CV: CV{
			Personal: Personal{
				FirstName: "Jane",
				LastName:  "Doe",
				Address: Address{
					Street:  "Musterstrasse 12",
					Zip:     "8000",
					City:    "Zürich",
					Country: "Schweiz",
				},
				Email: "jane.doe@example.com",
				Phone: "+41 79 123 45 67",
			},
			Summary: []string{
				"Solution architect with 8+ years of experience designing scalable CRM and low-code platform solutions.",
			},
			Experience: []Experience{
				{
					Company: "Example AG",
					Title:   "Solution Architect",
					Start:   "01/2021",
					Responsibilities: []string{
						"owns platform architecture for CRM and Power Platform",
					},
				},
				{
					Company: "Sample GmbH",
					Title:   "Software Engineer",
					Start:   "01/2017",
					End:     "12/2020",
					Responsibilities: []string{
						"built integrations and automations on Dynamics 365",
					},
				},
			},
			Education: []Education{
				{
					Institution: "ETH Zürich",
					Degree:      "BSc Computer Science",
					Start:       "09/2013",
					End:         "06/2017",
					Graduated:   true,
				},
			},
			Skills: []SkillGroup{
				{
					Category: "Technologies",
					Items: []Skill{
						{Name: "Dynamics 365", Level: 3},
						{Name: "Power Platform", Level: 3},
						{Name: "Azure Functions", Level: 2},
						{Name: "Solution Design", Level: 3},
						{Name: "Go", Level: 2},
						{Name: "TypeScript", Level: 2},
					},
				},
			},
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
  <h1>{{.CV.Personal.FirstName}} {{.CV.Personal.LastName}}</h1>
  <p class="contact">{{.CV.Personal.Email}} &middot; {{.CV.Personal.Phone}}</p>

  {{if .CV.Summary}}
  <h2>Summary</h2>
  {{range .CV.Summary}}<p>{{.}}</p>{{end}}
  {{end}}

  {{if .CV.Experience}}
  <h2>Experience</h2>
  <ul>
    {{range .CV.Experience}}<li>{{.Title}}, {{.Company}} ({{.Start}} - {{if .End}}{{.End}}{{else}}present{{end}})</li>{{end}}
  </ul>
  {{end}}

  {{if .CV.Education}}
  <h2>Education</h2>
  <ul>
    {{range .CV.Education}}<li>{{.Degree}}, {{.Institution}} ({{.Start}} - {{.End}})</li>{{end}}
  </ul>
  {{end}}

  {{if .CV.Skills}}
  <h2>Skills</h2>
  {{range .CV.Skills}}<p class="skills">{{range $i, $s := .Items}}{{if $i}} &middot; {{end}}{{$s.Name}}{{end}}</p>{{end}}
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
    {{.CV.Personal.FirstName}} {{.CV.Personal.LastName}}<br>
    {{.CV.Personal.Address.Street}}<br>
    {{.CV.Personal.Address.Zip}} {{.CV.Personal.Address.City}}<br>
    {{.CV.Personal.Email}} &middot; {{.CV.Personal.Phone}}
  </div>

  <div class="recipient">
    {{.Job.Company.Name}}<br>
    {{if .Job.Company.Contact}}{{.Job.Company.Contact}}<br>{{end}}
    {{.Job.Company.Street}}<br>
    {{.Job.Company.City}}
  </div>

  <div class="date">{{.CoverLetter.PlaceAndDate}}</div>

  <p class="subject">Application: {{.Job.Title}}</p>

  <p>{{.CoverLetter.Salutation}}</p>

  {{range .CoverLetter.Paragraphs}}<p>{{.}}</p>
  {{end}}

  <p class="closing">{{.CoverLetter.Closing}}<br>{{.CV.Personal.FirstName}} {{.CV.Personal.LastName}}</p>
</body>
</html>`
