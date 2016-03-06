metalinter:
	git submodule update --init --recursive
	git submodule foreach --recursive git clean -df
	git submodule foreach --recursive git reset --hard
	go build -o /dev/null
	gometalinter --vendor --deadline=30s -E gofmt ./...
