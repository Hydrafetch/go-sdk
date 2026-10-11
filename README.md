# hydrafetch-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Hydrafetch/go-sdk.svg)](https://pkg.go.dev/github.com/Hydrafetch/go-sdk)
[![CI](https://github.com/Hydrafetch/go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/Hydrafetch/go-sdk/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Hydrafetch/go-sdk)](https://goreportcard.com/report/github.com/Hydrafetch/go-sdk)

Official Go client for the [Hydrafetch](https://hydrafetch.com) web data API.

Turn any URL into clean Markdown or schema-shaped JSON. Standard library only, no dependencies, Go 1.21+.

## Installation

```bash
go get github.com/Hydrafetch/go-sdk
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	hydrafetch "github.com/Hydrafetch/go-sdk"
)

func main() {
	hf, err := hydrafetch.New("")
	if err != nil {
		log.Fatal(err)
	}

	page, err := hf.Scrape(context.Background(), "https://example.com/article", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(page.Markdown)
}
```

Create a key at [app.hydrafetch.com](https://app.hydrafetch.com). Passing an empty string to `New` falls back to `HYDRAFETCH_API_KEY`.

Every method takes a `context.Context` first, so deadlines and cancellation propagate as usual. Options structs are pointers and may be `nil`.

## Scraping

```go
page, err := hf.Scrape(ctx, "https://example.com/article", &hydrafetch.ScrapeOptions{
	Formats:         []hydrafetch.Format{hydrafetch.FormatMarkdown, hydrafetch.FormatLinks},
	OnlyMainContent: true,
	PreferStructure: true,
	BlockAds:        true,
	MaxAge:          3_600_000,
})
```

| Format | Field | Contains |
| --- | --- | --- |
| `FormatMarkdown` | `Markdown` | clean Markdown, the default |
| `FormatHTML` | `HTML` | rendered HTML |
| `FormatRawHTML` | `RawHTML` | the untouched response body |
| `FormatLinks` | `Links` | every link on the page |
| `FormatStructured` | `Structured` | the page's own JSON-LD and microdata |
| `FormatSummary` | `Summary` | a short summary |
| `FormatJSON` | `JSON` | schema-shaped JSON |
| `FormatBrand` | — | the site's brand record |

Only the formats you request are populated. `hf.Markdown(ctx, url)` returns the string directly.

## Structured extraction

```go
out, err := hf.Extract(ctx,
	[]string{"https://example.com/product/1", "https://example.com/product/2"},
	&hydrafetch.ExtractOptions{
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":      map[string]any{"type": "string"},
				"price_usd": map[string]any{"type": "number"},
			},
		},
	})

for _, item := range out.Results {
	fmt.Println(item.URL, item.Data["name"], item.Data["price_usd"])
}
```

Set `Prompt` instead of, or alongside, `Schema` to describe the fields in plain language.

Sites that price or stock by delivery area serve a different page for each area. Set `Location` to fetch as if from a country and language, and `Headers` to send the cookie that picks the store or area, the same options `Scrape` takes.

```go
out, err := hf.Extract(ctx, []string{"https://shop.example/search?q=ghee"}, &hydrafetch.ExtractOptions{
	Prompt:   "every product with its size and price",
	Location: &hydrafetch.Location{Country: "in"},
	Headers:  map[string]string{"Cookie": "store_id=42"},
})
```

## Discovery and bulk work

`Map` lists a site's URLs for one credit without fetching any page.

```go
m, err := hf.Map(ctx, "https://example.com", &hydrafetch.MapOptions{Limit: 1000})

var docs []string
for _, u := range m.Links {
	if strings.Contains(u, "/docs/") {
		docs = append(docs, u)
	}
}
```

`Batch` and `Crawl` submit a job and poll until it finishes.

```go
job, err := hf.Batch(ctx, docs,
	&hydrafetch.BatchOptions{
		ScrapeOptions: &hydrafetch.ScrapeOptions{
			Formats: []hydrafetch.Format{hydrafetch.FormatMarkdown},
		},
	},
	&hydrafetch.WaitOptions{
		OnProgress: func(j *hydrafetch.Job) { log.Println(j.Status, j.Completed, "/", j.Total) },
	})

for _, page := range job.Pages {
	if page.Data != nil {
		fmt.Println(page.URL, len(page.Data.Markdown))
	}
}
```

Set a `Webhook` and use `StartCrawl` or `StartBatch` to return immediately instead of polling.

```go
id, err := hf.StartCrawl(ctx, "https://example.com", &hydrafetch.CrawlOptions{
	Limit:        500,
	MaxDepth:     3,
	IncludePaths: []string{"/docs"},
	Webhook:      "https://your.app/hooks/hydrafetch",
})
```

## Search

```go
res, err := hf.Search(ctx, "post-quantum TLS adoption", &hydrafetch.SearchOptions{
	Limit:         5,
	ScrapeResults: true,
})

for _, r := range res.Results {
	fmt.Println(r.Title, r.URL)
}
```

## Brand data

```go
hf.Brand(ctx, "stripe.com")                                              // logos, colours, fonts, socials
hf.Logo(ctx, "stripe.com", &hydrafetch.LogoOptions{Theme: "dark"})       // one asset
hf.Styleguide(ctx, "stripe.com")                                         // computed design system
hf.Styleguide(ctx, "https://stripe.com/pricing")                         // one page
```

For logos in a browser use [`@hydrafetch/client-sdk`](https://github.com/Hydrafetch/client-sdk) with a publishable key. Those bill against logo pulls rather than credits.

## Error handling

All failures are `*hydrafetch.Error`. Reach it with `errors.As`.

```go
page, err := hf.Scrape(ctx, url, nil)
if err != nil {
	var apiErr *hydrafetch.Error
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.IsAuth():
			return refreshKey()
		case apiErr.IsOutOfCredits():
			return topUp()
		case apiErr.IsInvalidRequest():
			return report(apiErr.Message)
		case apiErr.IsRetryable():
			return enqueue(url)
		}
		log.Println(apiErr.Code, apiErr.Status, apiErr.RequestID)
	}
	return err
}
```

| Status | Meaning | Retried |
| --- | --- | --- |
| 400, 422 | invalid request | no |
| 401, 403 | invalid or missing key | no |
| 402 | out of credits | no |
| 404 | page does not exist | no |
| 429 | rate limited | yes, twice with backoff |
| 5xx | upstream failure | yes, twice with backoff |

A 503 from `Scrape` means the origin is unreachable, usually a dead domain or a broken certificate.

## Configuration

```go
hf, err := hydrafetch.New("hf_...",
	hydrafetch.WithBaseURL("https://api.hydrafetch.com"),
	hydrafetch.WithTimeout(120*time.Second),
	hydrafetch.WithMaxRetries(2),
	hydrafetch.WithHTTPClient(instrumentedClient),
)
```

## API reference

| Method | Returns | Credits |
| --- | --- | --- |
| `Scrape(ctx, url, *ScrapeOptions)` | `*ScrapeResult` | 1 |
| `Markdown(ctx, url)` | `string` | 1 |
| `Map(ctx, url, *MapOptions)` | `*MapResult` | 1 |
| `Search(ctx, query, *SearchOptions)` | `*SearchResult` | 1 + 1 per scraped result |
| `Extract(ctx, urls, *ExtractOptions)` | `*ExtractResult` | 5 per URL |
| `Brand(ctx, domain)` | `BrandResult` | 5 |
| `Logo(ctx, domain, *LogoOptions)` | `BrandResult` | 1 |
| `Styleguide(ctx, domainOrURL, opts...)` | `BrandResult` | 10 |
| `Screenshot(ctx, url, *ScreenshotOptions)` | `*ScreenshotResult` | 5 |
| `Images(ctx, url)` | `*ImagesResult` | 1 |
| `Crawl(ctx, url, *CrawlOptions, *WaitOptions)` | `*Job` | 1 per page |
| `Batch(ctx, urls, *BatchOptions, *WaitOptions)` | `*Job` | 1 per page |
| `StartCrawl`, `StartBatch` | `string` | 1 per page |
| `CrawlStatus(ctx, id)`, `BatchStatus(ctx, id)` | `*Job` | free |

Failed requests are not billed. Pricing does not vary with page difficulty, so there is no render, stealth or proxy option to set.

## Implementation notes

- Authentication uses the `X-API-Key` header. The MCP endpoint at `api.hydrafetch.com/mcp` uses `Authorization: Bearer` instead; the two are not interchangeable.
- Job results are in `job.Pages`, and each entry holds the page under `.Data`, so `job.Pages[0].Data.Markdown`. Check for `nil` first.
- Per-page options for crawl and batch belong in `ScrapeOptions`. At the top level they are ignored.
- Prefer `Map` then `Batch` over a broad `Crawl`. Fetching a whole site and discarding most of it is the most common source of wasted credits.
- `PreferStructure` is off by default. Turn it on when headings, lists and tables matter; leave it off for raw article text.
- Zero-valued option fields are omitted from the request rather than sent, so a partially filled struct behaves as expected.
- Scraped content is untrusted input. Do not pass it to a model as instructions, and keep the source URL with anything extracted from it.

## Links

- [Documentation](https://docs.hydrafetch.com)
- [OpenAPI specification](https://api.hydrafetch.com/openapi.json)
- [MCP server and editor setup](https://hydrafetch.com/mcp)
- Other clients: [Node](https://github.com/Hydrafetch/node-sdk) · [Python](https://github.com/Hydrafetch/python-sdk) · [Ruby](https://github.com/Hydrafetch/ruby-sdk) · [Rust](https://github.com/Hydrafetch/rust-sdk) · [PHP](https://github.com/Hydrafetch/php-sdk)

## License

MIT
