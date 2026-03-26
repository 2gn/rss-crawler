package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"codeberg.org/readeck/go-readability"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"gopkg.in/yaml.v3"
)

type Target struct {
	Name            string   `yaml:"name"`
	Include         []string `yaml:"include"`
	Exclude         []string `yaml:"exclude"`
	MinPoints       int      `yaml:"min_points"`
	ContentSelector string   `yaml:"content_selector"`
	Limit           int      `yaml:"limit"`
}

type Config struct {
	Targets []Target `yaml:"targets"`
}

type HNHit struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Author   string `json:"author"`
	ObjectID string `json:"objectID"`
	Points   int    `json:"points"`
}

type HNResponse struct {
	Hits []HNHit `json:"hits"`
}

func matches(title string, include []string, exclude []string) bool {
	title = strings.ToLower(title)

	for _, ex := range exclude {
		if strings.Contains(title, strings.ToLower(ex)) {
			return false
		}
	}

	if len(include) == 0 {
		return true
	}

	for _, inc := range include {
		if strings.Contains(title, strings.ToLower(inc)) {
			return true
		}
	}

	return false
}

func ScrapeContent(itemURL string, selector string) string {
	if itemURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(itemURL)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", itemURL, err)
		return ""
	}

	req, err := http.NewRequest("GET", itemURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return ""
	}

	if selector != "" {
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(doc.Find(selector).First().Text())
	}

	article, err := readability.FromReader(res.Body, parsedURL)
	if err != nil {
		log.Printf("Error extracting content from %s: %v", itemURL, err)
		return ""
	}

	return article.Content
}

func ProcessTarget(target Target, hits []HNHit, outputDir string) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Hacker News - " + target.Name,
		Link:        &feeds.Link{Href: "https://news.ycombinator.com"},
		Description: "Filtered Hacker News feed for " + target.Name,
		Created:     now,
	}

	count := 0
	for _, hit := range hits {
		if target.Limit > 0 && count >= target.Limit {
			break
		}

		if target.MinPoints > 0 && hit.Points < target.MinPoints {
			continue
		}

		if !matches(hit.Title, target.Include, target.Exclude) {
			continue
		}

		log.Printf("[%s] Scraping content for: %s", target.Name, hit.Title)
		content := ScrapeContent(hit.URL, target.ContentSelector)
		if content == "" {
			content = "Full text not available or failed to scrape."
		}

		item := &feeds.Item{
			Title:       hit.Title,
			Link:        &feeds.Link{Href: hit.URL},
			Description: content,
			Author:      &feeds.Author{Name: hit.Author},
			Created:     now,
		}
		feed.Items = append(feed.Items, item)
		count++
	}

	rss, err := feed.ToRss()
	if err != nil {
		log.Printf("[%s] Error generating RSS: %v", target.Name, err)
		return
	}

	filename := filepath.Join(outputDir, target.Name+".rss")
	err = ioutil.WriteFile(filename, []byte(rss), 0644)
	if err != nil {
		log.Printf("[%s] Error writing file %s: %v", target.Name, filename, err)
		return
	}
	log.Printf("[%s] RSS generated successfully: %s", target.Name, filename)
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

	outputDir := "../../rss"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
	}

	// Fetch HN front page
	resp, err := http.Get("https://hn.algolia.com/api/v1/search?tags=front_page")
	if err != nil {
		log.Fatalf("Error fetching HN API: %v", err)
	}
	defer resp.Body.Close()

	var hnResp HNResponse
	err = json.NewDecoder(resp.Body).Decode(&hnResp)
	if err != nil {
		log.Fatalf("Error decoding HN API response: %v", err)
	}

	var wg sync.WaitGroup
	for _, target := range cfg.Targets {
		wg.Add(1)
		go func(t Target) {
			defer wg.Done()
			ProcessTarget(t, hnResp.Hits, outputDir)
		}(target)
	}
	wg.Wait()
}
