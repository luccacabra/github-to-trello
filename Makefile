all: dep
	go build -o aws-github-to-trello

dep:
	dep ensure

build: GOOS=linux GOARCH=amd64

build: dep
	go build -o aws-github-to-trello