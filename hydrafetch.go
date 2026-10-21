package hydrafetch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultBaseURL    = "https://api.hydrafetch.com"
	defaultTimeout    = 120 * time.Second
	defaultMaxRetries = 2
)

type Client struct {
	apiKey        string
	baseURL       string
	httpClient    *http.Client
	maxRetries    int
	lastRequestID atomic.Value
}

type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

func WithMaxRetries(n int) Option {
	return func(c *Client) { c.maxRetries = n }
}

func Bool(v bool) *bool {
	return &v
}

func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv("HYDRAFETCH_API_KEY")
	}
	if apiKey == "" {
		return nil, &Error{
			Code:    "MISSING_API_KEY",
			Status:  401,
			Message: "no API key: pass one to New or set HYDRAFETCH_API_KEY (https://app.hydrafetch.com)",
		}
	}
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		maxRetries: defaultMaxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

type envelope struct {
	Success *bool           `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details"`
	} `json:"error"`
	Meta *struct {
		RequestID string `json:"requestId"`
	} `json:"meta"`
}

func (c *Client) do(ctx context.Context, method, path string, body any, query url.Values, out any) error {
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return &Error{Code: "ENCODE", Status: 400, Message: err.Error()}
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
		if err != nil {
			return &Error{Code: "REQUEST", Status: 400, Message: err.Error()}
		}
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Accept", "application/json")
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return &Error{Code: "TIMEOUT", Status: 408, Message: ctx.Err().Error()}
			}
			lastErr = &Error{Code: "TRANSPORT", Status: 500, Message: err.Error()}
			if attempt == c.maxRetries {
				return lastErr
			}
			sleep(ctx, attempt)
			continue
		}

		if id := res.Header.Get("X-Request-Id"); id != "" {
			c.lastRequestID.Store(id)
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return &Error{Code: "READ", Status: res.StatusCode, Message: readErr.Error()}
		}

		var env envelope
		_ = json.Unmarshal(raw, &env)

		failed := res.StatusCode >= 400 || (env.Success != nil && !*env.Success)
		if failed {
			apiErr := &Error{Code: "HTTP_" + strconv.Itoa(res.StatusCode), Status: res.StatusCode}
			if env.Error != nil {
				if env.Error.Code != "" {
					apiErr.Code = env.Error.Code
				}
				apiErr.Message = env.Error.Message
				apiErr.Details = env.Error.Details
			}
			if apiErr.Message == "" {
				apiErr.Message = strings.TrimSpace(string(raw))
			}
			if env.Meta != nil {
				apiErr.RequestID = env.Meta.RequestID
			}
			if !apiErr.IsRetryable() || attempt == c.maxRetries {
				return apiErr
			}
			lastErr = apiErr
			sleep(ctx, attempt)
			continue
		}

		if out == nil {
			return nil
		}
		target := env.Data
		if len(target) == 0 {
			target = raw
		}
		if err := json.Unmarshal(target, out); err != nil {
			return &Error{Code: "DECODE", Status: res.StatusCode, Message: err.Error()}
		}
		return nil
	}
	return lastErr
}

func (c *Client) LastRequestID() string {
	id, _ := c.lastRequestID.Load().(string)
	return id
}

func (c *Client) RateResult(ctx context.Context, in RateResultInput) (*RateResultResponse, error) {
	if in.RequestID == "" && in.URL == "" {
		return nil, &Error{Code: "VALIDATION_ERROR", Status: 400, Message: "pass RequestID or URL to say which result you are rating"}
	}
	var out RateResultResponse
	if err := c.do(ctx, http.MethodPost, "/v1/web/feedback", in, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func sleep(ctx context.Context, attempt int) {
	d := time.Duration(1<<attempt)*500*time.Millisecond + time.Duration(rand.Intn(250))*time.Millisecond
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

func scrapeBody(rawURL string, opts *ScrapeOptions) map[string]any {
	body := map[string]any{"url": rawURL}
	if opts == nil {
		return body
	}
	encoded, _ := json.Marshal(opts)
	var fields map[string]any
	_ = json.Unmarshal(encoded, &fields)
	for k, v := range fields {
		body[k] = v
	}
	return body
}

func (c *Client) Scrape(ctx context.Context, rawURL string, opts *ScrapeOptions) (*ScrapeResult, error) {
	var out ScrapeResult
	if err := c.do(ctx, http.MethodPost, "/v1/web/scrape", scrapeBody(rawURL, opts), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Markdown(ctx context.Context, rawURL string) (string, error) {
	page, err := c.Scrape(ctx, rawURL, &ScrapeOptions{Formats: []Format{FormatMarkdown}})
	if err != nil {
		return "", err
	}
	return page.Markdown, nil
}

func (c *Client) Map(ctx context.Context, rawURL string, opts *MapOptions) (*MapResult, error) {
	body := map[string]any{"url": rawURL}
	if opts != nil {
		encoded, _ := json.Marshal(opts)
		var fields map[string]any
		_ = json.Unmarshal(encoded, &fields)
		for k, v := range fields {
			body[k] = v
		}
	}
	var out MapResult
	if err := c.do(ctx, http.MethodPost, "/v1/web/map", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) (*SearchResult, error) {
	body := map[string]any{"query": query}
	if opts != nil {
		encoded, _ := json.Marshal(opts)
		var fields map[string]any
		_ = json.Unmarshal(encoded, &fields)
		for k, v := range fields {
			body[k] = v
		}
	}
	var out SearchResult
	if err := c.do(ctx, http.MethodPost, "/v1/web/search", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Extract(ctx context.Context, urls []string, opts *ExtractOptions) (*ExtractResult, error) {
	body := map[string]any{"urls": urls}
	if opts != nil {
		encoded, _ := json.Marshal(opts)
		var fields map[string]any
		_ = json.Unmarshal(encoded, &fields)
		for k, v := range fields {
			body[k] = v
		}
	}
	var out ExtractResult
	if err := c.do(ctx, http.MethodPost, "/v1/web/extract", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Brand(ctx context.Context, domain string) (BrandResult, error) {
	var out BrandResult
	if err := c.do(ctx, http.MethodGet, "/v1/web/brand", nil, url.Values{"domain": {domain}}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Logo(ctx context.Context, domain string, opts *LogoOptions) (BrandResult, error) {
	q := url.Values{"domain": {domain}}
	if opts != nil {
		if opts.Theme != "" {
			q.Set("theme", opts.Theme)
		}
		if opts.Type != "" {
			q.Set("type", opts.Type)
		}
	}
	var out BrandResult
	if err := c.do(ctx, http.MethodGet, "/v1/web/brand/logo", nil, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Styleguide(ctx context.Context, domainOrURL string, opts ...StyleguideOptions) (BrandResult, error) {
	q := url.Values{}
	lower := strings.ToLower(domainOrURL)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		q.Set("url", domainOrURL)
	} else {
		q.Set("domain", domainOrURL)
	}
	for _, o := range opts {
		if o.MaxAge != nil {
			q.Set("maxAge", strconv.Itoa(*o.MaxAge))
		}
	}
	var out BrandResult
	if err := c.do(ctx, http.MethodGet, "/v1/web/styleguide", nil, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Screenshot(ctx context.Context, rawURL string, opts *ScreenshotOptions) (*ScreenshotResult, error) {
	body := map[string]any{"url": rawURL}
	if opts != nil {
		encoded, _ := json.Marshal(opts)
		var fields map[string]any
		_ = json.Unmarshal(encoded, &fields)
		for k, v := range fields {
			body[k] = v
		}
	}
	var out ScreenshotResult
	if err := c.do(ctx, http.MethodPost, "/v1/web/screenshot", body, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Images(ctx context.Context, rawURL string) (*ImagesResult, error) {
	var out ImagesResult
	if err := c.do(ctx, http.MethodPost, "/v1/web/images", map[string]any{"url": rawURL}, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) StartCrawl(ctx context.Context, rawURL string, opts *CrawlOptions) (string, error) {
	body := map[string]any{"url": rawURL}
	if opts != nil {
		encoded, _ := json.Marshal(opts)
		var fields map[string]any
		_ = json.Unmarshal(encoded, &fields)
		for k, v := range fields {
			body[k] = v
		}
	}
	return c.startJob(ctx, "/v1/web/crawl", body, "crawlId")
}

func (c *Client) StartBatch(ctx context.Context, urls []string, opts *BatchOptions) (string, error) {
	body := map[string]any{"urls": urls}
	if opts != nil {
		encoded, _ := json.Marshal(opts)
		var fields map[string]any
		_ = json.Unmarshal(encoded, &fields)
		for k, v := range fields {
			body[k] = v
		}
	}
	return c.startJob(ctx, "/v1/web/batch", body, "batchId")
}

func (c *Client) startJob(ctx context.Context, path string, body map[string]any, idKey string) (string, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodPost, path, body, nil, &out); err != nil {
		return "", err
	}
	for _, key := range []string{idKey, "id"} {
		if v, ok := out[key].(string); ok && v != "" {
			return v, nil
		}
	}
	return "", &Error{Code: "NO_JOB_ID", Status: 502, Message: "job accepted but returned no id"}
}

func (c *Client) CrawlStatus(ctx context.Context, id string) (*Job, error) {
	return c.jobStatus(ctx, "/v1/web/crawl/"+url.PathEscape(id))
}

func (c *Client) BatchStatus(ctx context.Context, id string) (*Job, error) {
	return c.jobStatus(ctx, "/v1/web/batch/"+url.PathEscape(id))
}

func (c *Client) jobStatus(ctx context.Context, path string) (*Job, error) {
	var out Job
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type WaitOptions struct {
	PollInterval time.Duration
	Timeout      time.Duration
	OnProgress   func(*Job)
}

func (c *Client) Crawl(ctx context.Context, rawURL string, opts *CrawlOptions, wait *WaitOptions) (*Job, error) {
	id, err := c.StartCrawl(ctx, rawURL, opts)
	if err != nil {
		return nil, err
	}
	return c.wait(ctx, func(ctx context.Context) (*Job, error) { return c.CrawlStatus(ctx, id) }, wait)
}

func (c *Client) Batch(ctx context.Context, urls []string, opts *BatchOptions, wait *WaitOptions) (*Job, error) {
	id, err := c.StartBatch(ctx, urls, opts)
	if err != nil {
		return nil, err
	}
	return c.wait(ctx, func(ctx context.Context) (*Job, error) { return c.BatchStatus(ctx, id) }, wait)
}

func (c *Client) wait(ctx context.Context, poll func(context.Context) (*Job, error), wait *WaitOptions) (*Job, error) {
	interval := 2 * time.Second
	timeout := 5 * time.Minute
	var onProgress func(*Job)
	if wait != nil {
		if wait.PollInterval > 0 {
			interval = wait.PollInterval
		}
		if wait.Timeout > 0 {
			timeout = wait.Timeout
		}
		onProgress = wait.OnProgress
	}

	deadline := time.Now().Add(timeout)
	for {
		job, err := poll(ctx)
		if err != nil {
			return nil, err
		}
		if onProgress != nil {
			onProgress(job)
		}
		if job.Done() {
			return job, nil
		}
		if time.Now().Add(interval).After(deadline) {
			return nil, &Error{
				Code:    "TIMEOUT",
				Status:  408,
				Message: fmt.Sprintf("job did not finish within %s; it is still running server-side, so poll its status or pass a webhook", timeout),
			}
		}
		select {
		case <-ctx.Done():
			return nil, &Error{Code: "TIMEOUT", Status: 408, Message: ctx.Err().Error()}
		case <-time.After(interval):
		}
	}
}
