package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
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

func getCreds(authFile *os.File) (string, string, error) {
	raise := func(err error) (string, string, error) {
		return "", "", err
	}

	reader := bufio.NewReader(authFile)

	user, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return raise(err)
	}
	pass, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return raise(err)
	}

	return user[:len(user)-1], pass[:len(pass)-1], err
}

func main() {
	url := kingpin.Flag("jira-url", "JIRA instance URL").URL()
	authFile := kingpin.Flag("auth-file", "Path to a file with auth credentials. Must have <user> on line 1 and <pass> on line 2.").File()
	maxIssues := kingpin.Flag("max-issues", "Number of issues to look at for labels in the project.").Default("50").Int()
	project := kingpin.Arg("project", "Project to normalize labels on.").Required().String()

	defer (*authFile).Close()
	kingpin.Parse()

	user, pass, err := getCreds(*authFile)
	if err != nil {
		panic(err)
	}

	client, err := jira.NewClient(nil, (*url).String())
	if err != nil {
		panic(err)
	}
	client.Authentication.SetBasicAuth(user, pass)

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
