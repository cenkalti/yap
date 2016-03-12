build:
	go build -o /dev/null

install:
	go install

metalinter: install
	gometalinter --vendor --deadline=30s -E gofmt -D dupl ./...
