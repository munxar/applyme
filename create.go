package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const jobAdAPIBaseURL = "https://www.job-room.ch/jobadservice/api/jobAdvertisements"

// Job mirrors the JobAdvertisement schema of the job-room.ch API.
type Job struct {
	ID                         string      `json:"id"`
	CreatedTime                string      `json:"createdTime"`
	UpdatedTime                string      `json:"updatedTime"`
	Status                     string      `json:"status"`
	SourceSystem               string      `json:"sourceSystem"`
	ExternalReference          string      `json:"externalReference"`
	StellennummerEgov          string      `json:"stellennummerEgov"`
	StellennummerAvam          *string     `json:"stellennummerAvam"`
	Fingerprint                string      `json:"fingerprint"`
	ReportingObligation        bool        `json:"reportingObligation"`
	ReportingObligationEndDate *string     `json:"reportingObligationEndDate"`
	ReportToAvam               bool        `json:"reportToAvam"`
	JobCenterCode              *string     `json:"jobCenterCode"`
	JobCenterUserID            *string     `json:"jobCenterUserId"`
	ApprovalDate               *string     `json:"approvalDate"`
	RejectionDate              *string     `json:"rejectionDate"`
	RejectionCode              *string     `json:"rejectionCode"`
	RejectionReason            *string     `json:"rejectionReason"`
	CancellationDate           *string     `json:"cancellationDate"`
	CancellationCode           *string     `json:"cancellationCode"`
	JobContent                 JobContent  `json:"jobContent"`
	Publication                Publication `json:"publication"`
	Owner                      *string     `json:"owner"`
}

type JobContent struct {
	ExternalURL     string           `json:"externalUrl"`
	NumberOfJobs    *string          `json:"numberOfJobs"`
	JobDescriptions []JobDescription `json:"jobDescriptions"`
	Company         Company          `json:"company"`
	Employment      Employment       `json:"employment"`
	Location        Location         `json:"location"`
	Occupations     []Occupation     `json:"occupations"`
	LanguageSkills  []LanguageSkill  `json:"languageSkills"`
	ApplyChannel    ApplyChannel     `json:"applyChannel"`
	PublicContact   *PublicContact   `json:"publicContact"`
}

type JobDescription struct {
	LanguageIsoCode string `json:"languageIsoCode"`
	Title           string `json:"title"`
	Description     string `json:"description"`
}

type Company struct {
	Name                    string  `json:"name"`
	Street                  string  `json:"street"`
	HouseNumber             string  `json:"houseNumber"`
	PostalCode              string  `json:"postalCode"`
	City                    string  `json:"city"`
	CountryIsoCode          string  `json:"countryIsoCode"`
	PostOfficeBoxNumber     *string `json:"postOfficeBoxNumber"`
	PostOfficeBoxPostalCode *string `json:"postOfficeBoxPostalCode"`
	PostOfficeBoxCity       *string `json:"postOfficeBoxCity"`
	Phone                   *string `json:"phone"`
	Email                   *string `json:"email"`
	Website                 *string `json:"website"`
	Surrogate               bool    `json:"surrogate"`
}

type Employment struct {
	StartDate             *string  `json:"startDate"`
	EndDate               *string  `json:"endDate"`
	ShortEmployment       bool     `json:"shortEmployment"`
	Immediately           bool     `json:"immediately"`
	Permanent             bool     `json:"permanent"`
	WorkloadPercentageMin string   `json:"workloadPercentageMin"`
	WorkloadPercentageMax string   `json:"workloadPercentageMax"`
	WorkForms             []string `json:"workForms"`
}

type Location struct {
	ID             *string     `json:"id"`
	Remarks        *string     `json:"remarks"`
	City           string      `json:"city"`
	PostalCode     string      `json:"postalCode"`
	CommunalCode   string      `json:"communalCode"`
	RegionCode     string      `json:"regionCode"`
	CantonCode     string      `json:"cantonCode"`
	CountryIsoCode string      `json:"countryIsoCode"`
	Coordinates    Coordinates `json:"coordinates"`
}

type Coordinates struct {
	Lon string `json:"lon"`
	Lat string `json:"lat"`
}

type Occupation struct {
	AvamOccupationCode string  `json:"avamOccupationCode"`
	WorkExperience     *string `json:"workExperience"`
	EducationCode      *string `json:"educationCode"`
	QualificationCode  *string `json:"qualificationCode"`
}

// LanguageSkill levels are one of NONE, BASIC, INTERMEDIATE, PROFICIENT.
type LanguageSkill struct {
	LanguageIsoCode string `json:"languageIsoCode"`
	SpokenLevel     string `json:"spokenLevel"`
	WrittenLevel    string `json:"writtenLevel"`
}

type ApplyChannel struct {
	RawPostAddress *string `json:"rawPostAddress"`
	PostAddress    *string `json:"postAddress"`
	EmailAddress   *string `json:"emailAddress"`
	PhoneNumber    *string `json:"phoneNumber"`
	FormURL        string  `json:"formUrl"`
	AdditionalInfo *string `json:"additionalInfo"`
}

// PublicContact salutation is one of MR, MS.
type PublicContact struct {
	Salutation string  `json:"salutation"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Phone      *string `json:"phone"`
	Email      *string `json:"email"`
}

type Publication struct {
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
	EuresDisplay      bool   `json:"euresDisplay"`
	EuresAnonymous    bool   `json:"euresAnonymous"`
	PublicDisplay     bool   `json:"publicDisplay"`
	RestrictedDisplay bool   `json:"restrictedDisplay"`
	CompanyAnonymous  bool   `json:"companyAnonymous"`
}

func runCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing job.json files with freshly fetched data")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ids := fs.Args()

	if len(ids) == 0 {
		return fmt.Errorf("usage: applyme create [--force] <id> [<id> ...]")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	var failed []string
	// sequential on purpose: batch mode must not hammer the job-room api
	for _, id := range ids {
		if err := createApplication(client, id, *force); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", id, err)
			failed = append(failed, id)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to create %d application(s): %v", len(failed), failed)
	}
	return nil
}

func createApplication(client *http.Client, id string, force bool) error {
	// company name will be appended to the folder once we parse the fetched job data
	dir := filepath.Join("applications", id)
	jobPath := filepath.Join(dir, "job.json")

	if !force {
		if _, err := os.Stat(jobPath); err == nil {
			fmt.Printf("skipping %s (already exists)\n", jobPath)
			return nil
		}
	}

	job, err := fetchJobAdvertisement(client, id)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	body, err := marshalJSONIndent(job)
	if err != nil {
		return fmt.Errorf("marshaling job: %w", err)
	}

	if err := os.WriteFile(jobPath, body, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", jobPath, err)
	}
	fmt.Printf("created %s\n", jobPath)
	return nil
}

func fetchJobAdvertisement(client *http.Client, id string) (*Job, error) {
	url := fmt.Sprintf("%s/%s", jobAdAPIBaseURL, id)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching job ad: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	sanitizeJob(&job)

	return &job, nil
}
