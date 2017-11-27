package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

import (
	"gopkg.in/alecthomas/kingpin.v2"
	jira "gopkg.in/andygrunwald/go-jira.v1"
)

func getIssuesInProject(client *jira.Client, project string, maxIssues int) ([]jira.Issue, *jira.Response, error) {
	queryString := fmt.Sprintf("project = %s", project)
	queryOptions := &jira.SearchOptions{
		MaxResults: maxIssues,
		Fields: []string{"id,labels"},
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
	return nil
}

func updateIssues(client *jira.Client, issues []jira.Issue, labelMap map[string]string) error {
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
	for _, l := range labels { fmt.Println(l) }
	return
	normalizedLabelMap := normalizeLabels(labels)

	err = updateIssues(client, issues, normalizedLabelMap)
	if err != nil {
		panic(err)
	}
}
