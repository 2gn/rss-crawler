run-website:
    cd generator/website && go run main.go

run-hn:
    cd generator/hackernews && go run main.go

run: run-website run-hn
