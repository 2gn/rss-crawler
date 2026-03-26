# Hacker News Scraper

This generator fetches the front page of Hacker News using the Algolia API, filters stories by title, and scrapes the full text of each story to generate a full-text RSS feed.

## Configuration

The configuration is defined in `config.yaml`.

Example target:

```yaml
targets:
  - name: "hn-ai"
    filter_title: "AI"
    content_selector: "article"
    limit: 5
```

- `name`: The name of the output RSS file (e.g., `hn-ai.rss`).
- `filter_title`: Only include stories whose titles contain this string (case-insensitive).
- `content_selector`: (Optional) CSS selector to extract the full text from the original article.
- `limit`: The maximum number of stories to include in the feed.
