package main

import "regexp"

// jobRoomEscapedPunctuation matches the stray Markdown-style backslash
// escaping the job-room.ch API puts in front of punctuation in free-text
// fields (e.g. "80\-100%", "D365\&Power"). It's meaningless outside a
// Markdown renderer, so strip it before we write the data out.
var jobRoomEscapedPunctuation = regexp.MustCompile(`\\([!"#$%&'()*+,\-./:;<=>?@\[\]^_` + "`" + `{|}~])`)

func stripStrayBackslashEscapes(s string) string {
	return jobRoomEscapedPunctuation.ReplaceAllString(s, "$1")
}

func sanitizeJob(job *Job) {
	for i := range job.JobContent.JobDescriptions {
		d := &job.JobContent.JobDescriptions[i]
		d.Title = stripStrayBackslashEscapes(d.Title)
		d.Description = stripStrayBackslashEscapes(d.Description)
	}
}
