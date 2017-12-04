.PHONY: all
all: install

.PHONY: install
install:
	glide install
	go build .

.PHONY: clean
clean:
	rm -f jira-issue-normalizer
