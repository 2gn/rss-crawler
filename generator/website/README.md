# Website Scraper

This generator scrapes websites using CSS selectors to generate RSS feeds.

## Configuration

The configuration is defined in `config.yaml`.

Example target:

```yaml
targets:
  - name: "i-harness"
    feed:
      title: "i-harness.com"
      link: "https://i-harness.com"
      description: "RSS feed for i-harness.com"
      author: "i-harness.com"
    scrape:
      url: "https://i-harness.com"
      item_selector: ".col-md-9"
      title_selector: "a"
      link_selector: "a"
      link_attr: "href"
      description_selector: "p"
      base_url: "https://i-harness.com"
```
