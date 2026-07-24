# applyme

a cli tool that helps you stay sane while searching a job. currently only working for https://www.job-room.ch and my RAV process.
It mainly helps organizing your cv and cover letters as structured data (json), and generates pdfs based on your templates.

## disclaimer

this tool is highly ai generated because I have more important things to do (apply for real jobs), so use with caution.

## installation

get the binary for your platform, or build from go code

todo: go install xyz
if installed correctly type **applyme** in a terminal.

```bash
applyme
```

todo: demo of help output

## create a new project

```bash
applyme init
```

initializes a project in the current folder with a template for cv.json and a config.json for settings, both with a matching json schema for editor tooling (autocomplete/validation).
note: it'll warn you if the folder is not empty.

settings in config.json (job-room api base url, applications folder, request timeout) can be overridden per command with flags, e.g. `applyme fetch --applications-dir other-folder {id}`.

## fetch a job advertisement

```bash
applyme fetch {id}
```

fetches the json from the unofficial api https://www.job-room.ch/jobadservice/api/jobAdvertisements/{id} and stores the content in a folder under "applications/{id} company name to identify application for humans/job-advertisement.json" (id is assumed uuid v4 for now)
the job-advertisement.json should only store relevant informations not everything.

note: a batch mode with multiple ids should be possible, so better prepare for that, but keep fetching sequencial, to prevent rate limiting.

## create a cv and cover letter for an application

```bash
applyme generate {id}
```

generates your application assets for {id} out of templates and your structured data living under the specified {id}.

if `applications/{id}/application.json` doesn't exist yet, it's created first: personal/contact data is copied from `cv.json`, the job is filled in from a sibling `job-advertisement.json` if you already ran `applyme fetch {id}`, and the cover letter body is left as generic placeholder text for you to rewrite. it comes with a matching json schema (`application.schema.json`, created by `applyme init`) for editor tooling.

## license

[MIT](LICENSE)
