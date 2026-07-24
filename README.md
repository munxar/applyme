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

initializes a project in the current folder with a template for cv.json, a .env for settings and api keys.
note: it'll warn you if the folder is not empty.

## create a application

```bash
applyme create {id}
```

fetches the json from the unofficial api https://www.job-room.ch/jobadservice/api/jobAdvertisements/{id} and stores the content in a folder under "applications/{id} company name to identify application for humans/job.json" (id is assumed uuid v4 for now)
the job.json should only store relevant informations not everything.

note: a batch mode with multiple ids should be possible, so better prepare for that, but keep fetching sequencial, to prevent rate limiting.

## create a cv and cover letter for an application

```bash
applyme generate {id}
```

generates your application assets for {id} out of templates and your structured data living under the specified {id}.
