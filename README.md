# RSS Crawler

A versatile framework for generating RSS feeds from various online resources.

## Structure

- `generator/website`: Aggressive website scraper using CSS selectors.
- `generator/hackernews`: Full-text Hacker News scraper using Algolia API.
- `rss/`: Directory where all generated RSS feeds are stored.

## Usage

Generate all feeds:
```bash
just run
```

Generate only website feeds:
```bash
just run-website
```

Generate only Hacker News feeds:
```bash
just run-hn
```
