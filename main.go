package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	args := os.Args[1:]
	debug := false
	var commandArgs []string

	for _, arg := range args {
		if arg == "--debug" {
			debug = true
		} else {
			commandArgs = append(commandArgs, arg)
		}
	}

	if len(commandArgs) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := commandArgs[0]
	generators, err := findGenerators()
	if err != nil {
		log.Fatalf("Error finding generators: %v", err)
	}

	switch command {
	case "list":
		fmt.Println("Available generators:")
		for _, g := range generators {
			fmt.Printf("  - %s\n", g)
		}
	case "all":
		for _, g := range generators {
			runGenerator(g, debug)
		}
		updateDocs()
	case "update-docs":
		updateDocs()
	default:
		// Assume it's a generator name
		found := false
		for _, g := range generators {
			if g == command {
				runGenerator(g, debug)
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Unknown command or generator: %s\n", command)
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [--debug] list              - List all available generators")
	fmt.Println("  go run main.go [--debug] all               - Run all available generators and update docs")
	fmt.Println("  go run main.go update-docs                 - Update feeds_list.md with current RSS files")
	fmt.Println("  go run main.go [--debug] <generator_name>  - Run a specific generator")
}

func findGenerators() ([]string, error) {
	var generators []string
	entries, err := os.ReadDir("generator")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it has a main.go
			mainFile := filepath.Join("generator", entry.Name(), "main.go")
			if _, err := os.Stat(mainFile); err == nil {
				generators = append(generators, entry.Name())
			}
		}
	}
	sort.Strings(generators)
	return generators, nil
}

func runGenerator(name string, debug bool) {
	fmt.Printf("Running generator: %s\n", name)
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = filepath.Join("generator", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if debug {
		cmd.Env = append(os.Environ(), "DEBUG=true")
	}
	if err := cmd.Run(); err != nil {
		log.Printf("Error running generator %s: %v", name, err)
	}
}

func updateDocs() {
	fmt.Println("Updating feeds_list.md...")
	f, err := os.Create("feeds_list.md")
	if err != nil {
		log.Fatalf("Error creating feeds_list.md: %v", err)
	}
	defer f.Close()

	fmt.Fprintln(f, "# RSS Feeds")
	fmt.Fprintln(f, "")

	files, err := os.ReadDir("rss")
	if err != nil {
		log.Fatalf("Error reading rss directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".rss" {
			name := strings.TrimSuffix(file.Name(), ".rss")
			fmt.Fprintf(f, "* [%s](https://github.com/2gn/rss-crawler/raw/refs/heads/main/rss/%s)\n", name, file.Name())
		}
	}
	fmt.Println("feeds_list.md updated successfully.")
}
