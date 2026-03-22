package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"gopkg.in/yaml.v3"
)

type Target struct {
	Name string `yaml:"name"`
	Feed struct {
		Title       string `yaml:"title"`
		Link        string `yaml:"link"`
		Description string `yaml:"description"`
		Author      string `yaml:"author"`
	} `yaml:"feed"`
	Scrape struct {
		URL                 string `yaml:"url"`
		ItemSelector        string `yaml:"item_selector"`
		TitleSelector       string `yaml:"title_selector"`
		LinkSelector        string `yaml:"link_selector"`
		LinkAttr            string `yaml:"link_attr"`
		DescriptionSelector string `yaml:"description_selector"`
		BaseURL             string `yaml:"base_url"`
	} `yaml:"scrape"`
}

type Config struct {
	Targets []Target `yaml:"targets"`
}

func ProcessTarget(target Target, outputDir string) {
	now := time.Now()

	feed := &feeds.Feed{
		Title:       target.Feed.Title,
		Link:        &feeds.Link{Href: target.Feed.Link},
		Description: target.Feed.Description,
		Created:     now,
	}

	// Request the HTML page.
	req, err := http.NewRequest("GET", target.Scrape.URL, nil)
	if err != nil {
		log.Printf("[%s] Error creating request: %v", target.Name, err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[%s] Error fetching URL: %v", target.Name, err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("[%s] Status code error: %d %s", target.Name, res.StatusCode, res.Status)
		return
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("[%s] Error parsing HTML: %v", target.Name, err)
		return
	}

	// Find entries
	doc.Find(target.Scrape.ItemSelector).Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := strings.TrimSpace(s.Find(target.Scrape.TitleSelector).First().Text())
		url, _ := s.Find(target.Scrape.LinkSelector).First().Attr(target.Scrape.LinkAttr)
		description := strings.TrimSpace(s.Find(target.Scrape.DescriptionSelector).First().Text())

		if url != "" && target.Scrape.BaseURL != "" && url[0] == '/' {
			url = target.Scrape.BaseURL + url
		}

		entry := &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: url},
			Description: description,
			Author:      &feeds.Author{Name: target.Feed.Author},
			Created:     now,
		}

		if title != "" {
			feed.Items = append(feed.Items, entry)
		}
	})

	rss, err := feed.ToRss()
	if err != nil {
		log.Printf("[%s] Error generating RSS: %v", target.Name, err)
		return
	}

	// write the result to file
	filename := filepath.Join(outputDir, target.Name+".rss")
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("[%s] Error creating file %s: %v", target.Name, filename, err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(rss)
	if err != nil {
		log.Printf("[%s] Error writing to file %s: %v", target.Name, filename, err)
		return
	}
	log.Printf("[%s] RSS generated successfully: %s", target.Name, filename)
}

func cleanup(outputDir string, activeNames []string) {
	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		log.Printf("Error reading output directory for cleanup: %v", err)
		return
	}

	activeMap := make(map[string]bool)
	for _, name := range activeNames {
		activeMap[name+".rss"] = true
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".rss" {
			if !activeMap[file.Name()] {
				path := filepath.Join(outputDir, file.Name())
				err := os.Remove(path)
				if err != nil {
					log.Printf("Error removing old feed file %s: %v", path, err)
				} else {
					log.Printf("Removed old feed file: %s", path)
				}
			}
		}
	}
}

func main() {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}

	outputDir := "rss"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
	}

	var activeNames []string
	var wg sync.WaitGroup
	for _, target := range cfg.Targets {
		activeNames = append(activeNames, target.Name)
		wg.Add(1)
		go func(t Target) {
			defer wg.Done()
			ProcessTarget(t, outputDir)
		}(target)
	}
	wg.Wait()

	cleanup(outputDir, activeNames)
}
