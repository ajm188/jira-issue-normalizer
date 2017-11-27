package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

import (
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

func getCreds(authFilePath string) (string, string, error) {
	raise := func(err error) (string, string, error) {
		return "", "", err
	}

	authFile, err := os.Open(authFilePath)
	if err != nil {
		return raise(err)
	}
	defer authFile.Close()

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
	url := flag.String("jira-url", os.Getenv("JIRA_URL"), "URL of your JIRA instance")
	authFilePath := flag.String("auth-file", os.Getenv("JIRA_AUTH_FILE"), "Path to a file with auth credentials. Must have <user> on line 1 and <pass> on line 2.")
	maxIssues := flag.Int("max-issues", 50, "Number of issues to look for labels in the project.")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <project>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	user, pass, err := getCreds(*authFilePath)
	if err != nil {
		panic(err)
	}

	client, err := jira.NewClient(nil, *url)
	if err != nil {
		panic(err)
	}
	client.Authentication.SetBasicAuth(user, pass)

	project := flag.Arg(0)
	issues, r, err := getIssuesInProject(client, project, *maxIssues)
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
