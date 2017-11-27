# jira-issue-normalizer

Normalizes JIRA labels on issues in a project.

## Motivation

JIRA labels are case sensitive, meaning "my-label" and "my-LaBeL" are two different labels.
This is dumb.
`jira-issue-normalizer` combs through the labels in a project and does its best to standardize labels to their downcased and dash-separated forms ("my-label" is preferred over "mylabel", and "my-long-label" is preferred over "mylong-label" or "my-longlabel").

## How to Run This

### Setup Go

If you don't have Go installed and setup on your system, you'll need to do that first.
Follow instructions at [golang.org](https://golang.org).

### Install dependencies

Simply run `./bin/setup` on any UNIX system.
It installs all the dependencies for you.

### Running

After completing the above, running is as simple as:

```bash
$ go run main.go --jira-url <my-jira> --auth-file <path-to-creds> my-project
```

For more options, run `go run main.go --help`.
