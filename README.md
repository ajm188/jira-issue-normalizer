# jira-issue-normalizer

Normalizes JIRA labels on issues in a project.

## Motivation

JIRA labels are case sensitive, meaning "my-label" and "my-LaBeL" are two different labels.
This is dumb.
`jira-issue-normalizer` combs through the labels in a project and does its best to standardize labels to their downcased and dash-separated forms ("my-label" is preferred over "mylabel", and "my-long-label" is preferred over "mylong-label" or "my-longlabel").

## Usage

See `--help` for all usage and options.

First, you will need a go installation on your system.
[Follow these instructions](https://golang.org).

After installing go, you'll want to ensure any binaries you install via the go tooling are in your PATH:

```bash
$ export PATH=$GOPATH/bin:$PATH
```

### via `go install`

```bash
$ go install github.com/ajm188/jira-issue-normalizer
$ jira-issue-normalizer --help
```

### from source

You will need [glide](https://glide.sh) to install dependencies.

```bash
$ go install github.com/Masterminds/glide
```

```bash
$ git clone git@github.com:ajm188/jira-issue-normalizer && cd jira-issue-normalizer
$ glide install
$ go build .
$ ./jira-issue-normalizer --help
```
