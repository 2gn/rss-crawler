package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	"log"
	"net/http"
	"os"
	"time"
)

func GetEntries() {
	now := time.Now()

	feed := &feeds.Feed{
		Title:       "i-harness.com",
		Link:        &feeds.Link{Href: "https://i-harness.com"},
		Description: "",
		Created:     now,
	}

	// Request the HTML page.
	res, err := http.Get("https://i-harness.com")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find entries
	doc.Find(".col-md-9").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Find("a").Text()
		url, _ := s.Find("a").Attr("href")
		description := s.Find("p").Text()
		url = "https://i-harness.com" + url
		// fmt.Printf("%s %s \n", url, title)

		entry := &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: url},
			Description: description,
			Author:      &feeds.Author{Name: "i-harness.com"},
			Created:     now,
		}

		feed.Items = append(feed.Items, entry)

		// fmt.Printf("%s %s %s\n", url, title, description)
	})

	rss, err := feed.ToRss()

	if err != nil {
		log.Fatal(err)
	}

	// write the result to file

	f, err := os.Create("feed.rss")

	if err != nil {
		fmt.Println("Error creating file: ", err)
	}

	defer f.Close()

	f.WriteString(rss)
}

func main() {
	GetEntries()
}
