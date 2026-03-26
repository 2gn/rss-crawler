run *args:
    go run main.go {{args}} all

list *args:
    go run main.go {{args}} list

update-docs:
    go run main.go update-docs
