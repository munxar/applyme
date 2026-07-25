package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/carlos7ags/folio/document"
	"github.com/carlos7ags/folio/html"
	"github.com/carlos7ags/folio/layout"
	"github.com/carlos7ags/folio/tmpl"
)

//go:embed application.schema.json
var applicationSchemaJSON []byte

//go:embed templates/cv.html
var cvTemplateHTML string

//go:embed templates/cover.html
var coverTemplateHTML string

// Application is the view model that drives cv.pdf and cover.pdf
// generation. It merges the applicant's full CV (cv.json) with the
// application-specific data of {id}/application.json into the single
// shape the templates render from.
type Application struct {
	ID string

	Job Job

	CV          CV
	CoverLetter CoverLetter
}

// ApplicationData is the structure of {id}/application.json: the
// application-specific data generate cannot derive from cv.json alone. Its
// Personal field is a copy of cv.json's personal data taken at creation
// time, so the cover letter keeps its own sender snapshot independent of
// later cv.json edits.
type ApplicationData struct {
	Job         Job         `json:"job"`
	Personal    Personal    `json:"personal"`
	CoverLetter CoverLetter `json:"cover_letter"`
}

// Job is the position being applied for, at a given company.
type Job struct {
	Title   string      `json:"title"`
	Company CompanyInfo `json:"company"`
}

// CompanyInfo is the recipient side of an application.
type CompanyInfo struct {
	Name    string `json:"name"`
	Street  string `json:"street"`
	City    string `json:"city"`
	Contact string `json:"contact,omitempty"` // e.g. "Jane Doe", empty if unknown
}

// CoverLetter holds the free-text parts of the cover letter.
type CoverLetter struct {
	PlaceAndDate string   `json:"place_and_date"`
	Salutation   string   `json:"salutation"`
	Paragraphs   []string `json:"paragraphs"`
	Closing      string   `json:"closing"`
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
	dir := filepath.Join(cfg.ApplicationsDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	cv, err := readCVFile()
	if err != nil {
		return err
	}

	data, err := loadOrCreateApplicationData(dir, cv)
	if err != nil {
		return err
	}

	app := Application{
		ID:          id,
		Job:         data.Job,
		CV:          cv,
		CoverLetter: data.CoverLetter,
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
		ConvertOpts: &html.Options{
			AllowAbsolutePaths: true,
		},
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

// loadOrCreateApplicationData reads {dir}/application.json. If it doesn't
// exist yet, it creates one: personal contact data is copied from cv (so
// the cover letter has a sender snapshot independent of later cv.json
// edits), job is filled in from a sibling job-advertisement.json if one
// was already fetched, and the cover letter body is left as generic
// placeholder text for the user to rewrite.
func loadOrCreateApplicationData(dir string, cv CV) (ApplicationData, error) {
	path := filepath.Join(dir, "application.json")

	raw, err := os.ReadFile(path)
	if err == nil {
		var data ApplicationData
		if err := json.Unmarshal(raw, &data); err != nil {
			return ApplicationData{}, fmt.Errorf("parsing %s: %w", path, err)
		}
		return data, nil
	}
	if !os.IsNotExist(err) {
		return ApplicationData{}, fmt.Errorf("reading %s: %w", path, err)
	}

	data := newApplicationData(dir, cv)

	schemaPath, err := filepath.Rel(dir, "application.schema.json")
	if err != nil {
		return ApplicationData{}, fmt.Errorf("resolving application schema path: %w", err)
	}

	body, err := marshalJSONIndent(struct {
		Schema string `json:"$schema"`
		ApplicationData
	}{
		Schema:          filepath.ToSlash(schemaPath),
		ApplicationData: data,
	})
	if err != nil {
		return ApplicationData{}, fmt.Errorf("marshaling application data: %w", err)
	}
	if err := os.WriteFile(path, body, 0644); err != nil {
		return ApplicationData{}, fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Printf("created %s\n", path)

	return data, nil
}

// newApplicationData builds the ApplicationData written for a fresh
// application: basic (personal) data copied from cv, job filled in from a
// sibling job-advertisement.json when present, and a generic cover letter
// the user is expected to rewrite.
func newApplicationData(dir string, cv CV) ApplicationData {
	data := ApplicationData{
		Personal:    cv.Personal,
		CoverLetter: genericCoverLetter(cv),
	}

	if job, err := readJobAdvertisementFile(filepath.Join(dir, "job-advertisement.json")); err == nil {
		data.Job = jobFromAdvertisement(job)
	}

	return data
}

// jobFromAdvertisement maps a fetched JobAdvertisement onto the Job shape
// application.json stores.
func jobFromAdvertisement(job *JobAdvertisement) Job {
	var title string
	if len(job.JobContent.JobDescriptions) > 0 {
		title = job.JobContent.JobDescriptions[0].Title
	}

	company := job.JobContent.Company
	street := strings.TrimSpace(company.Street + " " + company.HouseNumber)
	city := strings.TrimSpace(company.PostalCode + " " + company.City)

	var contact string
	if pc := job.JobContent.PublicContact; pc != nil {
		contact = strings.TrimSpace(pc.FirstName + " " + pc.LastName)
	}

	return Job{
		Title: title,
		Company: CompanyInfo{
			Name:    company.Name,
			Street:  street,
			City:    city,
			Contact: contact,
		},
	}
}

// genericCoverLetter returns generic, language-neutral placeholder cover
// letter text seeded with the applicant's city and today's date.
func genericCoverLetter(cv CV) CoverLetter {
	placeAndDate := time.Now().Format("02.01.2006")
	if city := cv.Personal.Address.City; city != "" {
		placeAndDate = city + ", " + placeAndDate
	}

	return CoverLetter{
		PlaceAndDate: placeAndDate,
		Salutation:   "Dear Hiring Team,",
		Paragraphs: []string{
			"TODO: write why you are a great fit for this role.",
		},
		Closing: "Kind regards,",
	}
}
