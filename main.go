package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

import (
	"gopkg.in/alecthomas/kingpin.v2"
	jira "gopkg.in/andygrunwald/go-jira.v1"
)

func getIssuesInProject(client *jira.Client, project string, maxIssues int) ([]jira.Issue, *jira.Response, error) {
	queryString := fmt.Sprintf("project = %s", project)
	queryOptions := &jira.SearchOptions{
		MaxResults: maxIssues,
		Fields:     []string{"id,labels"},
	}
	return client.Issue.Search(queryString, queryOptions)
}

func extractLabels(issues []jira.Issue) []string {
	labels := make([]string, 0)
	for _, issue := range issues {
		fields := issue.Fields
		labels = append(labels, fields.Labels...)
	}

	return labels
}

func normalizeLabels(labels []string) map[string]string {
	// This is a fairly non-intuitive, multi-step process, so I will
	// document it a bit here:
	//
	// Step 1:
	// Build a map going from labels to just the letters in those labels,
	// so we can group mostly-duplicated labels together. While we're doing
	// that, we also look for the longest label that falls in a particular
	// group (this is stored in canonicalLabels). This is so that when we
	// create the final label map, "my-long-label" will be preferred over
	// both "mylonglabel" and "my-longlabel".
	//
	// Step 2:
	// Go back over the labels and map each original label to the canonical
	// label for that label's group.
	labelMap := make(map[string]string, len(labels))
	labelsToLetters := make(map[string]string, len(labels))
	canonicalLabels := make(map[string]string, len(labels))

	lettersOnly := func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		default:
			return rune(-1)
		}
	}

	for _, label := range labels {
		downcased := strings.ToLower(label)
		letters := strings.Map(lettersOnly, downcased)
		canonicalLabel, exists := canonicalLabels[letters]
		if !exists || len(canonicalLabel) < len(label) {
			canonicalLabels[letters] = downcased
		}

		labelsToLetters[label] = letters
	}

	for _, label := range labels {
		labelMap[label] = canonicalLabels[labelsToLetters[label]]
	}
	return labelMap
}

func updateIssues(client *jira.Client, issues []jira.Issue, labelMap map[string]string) error {
	for _, issue := range issues {
		// reduce labels to a unique set
		labels := make(map[string]bool, len(issue.Fields.Labels))
		for _, label := range issue.Fields.Labels {
			labels[labelMap[label]] = true
		}

		// build list from set of unique normalized labels
		labelSlice := make([]string, 0, len(labels))
		for k, _ := range labels {
			labelSlice = append(labelSlice, k)
		}

		// build update payload
		payload := map[string]interface{}{
			"fields": map[string]interface{}{
				"labels": labelSlice,
			},
		}
		resp, err := client.Issue.UpdateIssue(issue.ID, payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", resp.Response.Request)
			return err
		}
	}
	return nil
}

type Creds struct {
	User     string
	Password string
}

func getCreds(authFilePath string) (*Creds, error) {
	data, err := ioutil.ReadFile(authFilePath)
	if err != nil {
		return nil, err
	}

	var creds Creds
	err = json.Unmarshal(data, &creds)
	return &creds, err
}

func main() {
	workdir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	url := kingpin.Flag(
		"jira-url",
		"JIRA instance URL",
	).Default("http://localhost").URL()
	authFilePath := kingpin.Flag(
		"auth-file",
		"Path to JSON file with auth credentials. Must have <user> and <password>.",
	).Default(
		path.Join(workdir, "auth.json"),
	).ExistingFile()
	maxIssues := kingpin.Flag(
		"max-issues",
		"Number of issues to look at for labels in the project.",
	).Default("50").Int()
	project := kingpin.Arg(
		"project",
		"Project to normalize labels on.",
	).Required().String()

	kingpin.Parse()

	creds, err := getCreds(*authFilePath)
	if err != nil {
		panic(err)
	}

	client, err := jira.NewClient(nil, (*url).String())
	if err != nil {
		panic(err)
	}
	client.Authentication.SetBasicAuth(creds.User, creds.Password)

	issues, r, err := getIssuesInProject(client, *project, *maxIssues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", *r.Response.Request)
		panic(err)
	}
	labels := extractLabels(issues)
	normalizedLabelMap := normalizeLabels(labels)

	err = updateIssues(client, issues, normalizedLabelMap)
	if err != nil {
		panic(err)
	}
}
