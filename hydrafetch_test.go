package hydrafetch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type capture struct {
	requests []*http.Request
	bodies   []map[string]any
}

func newServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*Client, *capture) {
	t.Helper()
	cap := &capture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
		cap.requests = append(cap.requests, r.Clone(r.Context()))
		cap.bodies = append(cap.bodies, body)
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	c, err := New("hf_test", WithBaseURL(srv.URL), WithMaxRetries(0))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c, cap
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func TestSendsAPIKeyHeaderAndNeverBearer(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{"markdown":"# hi"}}`)
	})
	if _, err := c.Scrape(context.Background(), "https://example.com", nil); err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	got := cap.requests[0]
	if got.Header.Get("X-API-Key") != "hf_test" {
		t.Errorf("X-API-Key = %q, want hf_test", got.Header.Get("X-API-Key"))
	}
	if auth := got.Header.Get("Authorization"); auth != "" {
		t.Errorf("Authorization must not be sent, got %q", auth)
	}
}

func TestNewRequiresAKey(t *testing.T) {
	t.Setenv("HYDRAFETCH_API_KEY", "")
	if _, err := New(""); err == nil {
		t.Fatal("expected an error without a key")
	}
}

func TestNewReadsTheEnvironment(t *testing.T) {
	t.Setenv("HYDRAFETCH_API_KEY", "hf_from_env")
	c, err := New("")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.apiKey != "hf_from_env" {
		t.Errorf("apiKey = %q", c.apiKey)
	}
}

func TestUnwrapsDataEnvelope(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{"markdown":"# hi","status":200}}`)
	})
	page, err := c.Scrape(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if page.Markdown != "# hi" || page.Status != 200 {
		t.Errorf("got %+v", page)
	}
}

func TestOptionsRideAlongsideTheURL(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{}}`)
	})
	_, err := c.Scrape(context.Background(), "https://example.com", &ScrapeOptions{
		Formats:         []Format{FormatMarkdown, FormatLinks},
		PreferStructure: true,
	})
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	body := cap.bodies[0]
	if body["url"] != "https://example.com" || body["preferStructure"] != true {
		t.Errorf("body = %+v", body)
	}
	if formats, ok := body["formats"].([]any); !ok || len(formats) != 2 {
		t.Errorf("formats = %+v", body["formats"])
	}
}

func TestExtractSendsLocationAndHeaders(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{"results":[]}}`)
	})
	_, err := c.Extract(context.Background(), []string{"https://shop.example/s?q=ghee"}, &ExtractOptions{
		Prompt:   "products and prices",
		Location: &Location{Country: "in", Languages: []string{"en-IN"}},
		Headers:  map[string]string{"Cookie": "store_id=42"},
	})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	body := cap.bodies[0]
	loc, ok := body["location"].(map[string]any)
	if !ok || loc["country"] != "in" {
		t.Errorf("location = %+v", body["location"])
	}
	headers, ok := body["headers"].(map[string]any)
	if !ok || headers["Cookie"] != "store_id=42" {
		t.Errorf("headers = %+v", body["headers"])
	}
}

func TestZeroValuedOptionsAreOmitted(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{}}`)
	})
	if _, err := c.Scrape(context.Background(), "https://example.com", &ScrapeOptions{BlockAds: Bool(true)}); err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if _, present := cap.bodies[0]["timeout"]; present {
		t.Error("unset timeout must not be serialised")
	}
}

func TestADefaultTrueOptionCanBeTurnedOff(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{}}`)
	})
	if _, err := c.Scrape(context.Background(), "https://example.com", &ScrapeOptions{OnlyMainContent: Bool(false)}); err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if v, present := cap.bodies[0]["onlyMainContent"]; !present || v != false {
		t.Errorf("onlyMainContent = %v, present %v", v, present)
	}
	if _, present := cap.bodies[0]["blockAds"]; present {
		t.Error("an unset default-true option must not be serialised")
	}
}

func TestGETEndpointsUseQueryParameters(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{}}`)
	})
	if _, err := c.Logo(context.Background(), "stripe.com", &LogoOptions{Theme: "dark"}); err != nil {
		t.Fatalf("Logo: %v", err)
	}
	q := cap.requests[0].URL.Query()
	if q.Get("domain") != "stripe.com" || q.Get("theme") != "dark" {
		t.Errorf("query = %v", q)
	}
}

func TestStyleguideReadsADomainOrOnePage(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{}}`)
	})
	if _, err := c.Styleguide(context.Background(), "stripe.com"); err != nil {
		t.Fatalf("Styleguide: %v", err)
	}
	fresh := 0
	if _, err := c.Styleguide(context.Background(), "https://stripe.com/pricing", StyleguideOptions{MaxAge: &fresh}); err != nil {
		t.Fatalf("Styleguide: %v", err)
	}
	if q := cap.requests[0].URL.Query(); q.Get("domain") != "stripe.com" {
		t.Errorf("domain query = %v", q)
	}
	q := cap.requests[1].URL.Query()
	if q.Get("url") != "https://stripe.com/pricing" || q.Get("maxAge") != "0" || q.Has("domain") {
		t.Errorf("page query = %v", q)
	}
}

func TestErrorCarriesCodeStatusAndRequestID(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"Invalid API key"},"meta":{"requestId":"abc"}}`)
	})
	_, err := c.Scrape(context.Background(), "https://example.com", nil)
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *Error, got %T", err)
	}
	if apiErr.Code != "UNAUTHORIZED" || apiErr.Status != 401 || apiErr.RequestID != "abc" {
		t.Errorf("got %+v", apiErr)
	}
	if !apiErr.IsAuth() {
		t.Error("401 should classify as auth")
	}
}

func TestStatusClassification(t *testing.T) {
	cases := []struct {
		status int
		check  func(*Error) bool
		name   string
	}{
		{401, (*Error).IsAuth, "auth"},
		{402, (*Error).IsOutOfCredits, "credits"},
		{422, (*Error).IsInvalidRequest, "invalid"},
		{429, (*Error).IsRetryable, "retryable"},
		{503, (*Error).IsRetryable, "retryable"},
	}
	for _, tc := range cases {
		c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, tc.status, `{"success":false,"error":{"message":"x"}}`)
		})
		_, err := c.Scrape(context.Background(), "https://example.com", nil)
		var apiErr *Error
		if !errors.As(err, &apiErr) || !tc.check(apiErr) {
			t.Errorf("status %d should classify as %s", tc.status, tc.name)
		}
	}
}

func TestDoesNotRetryValidationFailures(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, 422, `{"success":false,"error":{"message":"bad url"}}`)
	}))
	defer srv.Close()

	c, _ := New("hf_test", WithBaseURL(srv.URL), WithMaxRetries(3))
	if _, err := c.Scrape(context.Background(), "nope", nil); err == nil {
		t.Fatal("expected an error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestRetriesServerErrors(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls < 3 {
			writeJSON(w, 503, `{"success":false,"error":{"message":"upstream"}}`)
			return
		}
		writeJSON(w, 200, `{"success":true,"data":{"markdown":"ok"}}`)
	}))
	defer srv.Close()

	c, _ := New("hf_test", WithBaseURL(srv.URL), WithMaxRetries(3))
	page, err := c.Scrape(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if page.Markdown != "ok" || calls != 3 {
		t.Errorf("markdown=%q calls=%d", page.Markdown, calls)
	}
}

func TestBatchIDComesFromTopLevel(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 201, `{"batchId":"b_1","status":"pending"}`)
	})
	id, err := c.StartBatch(context.Background(), []string{"https://a.example"}, nil)
	if err != nil {
		t.Fatalf("StartBatch: %v", err)
	}
	if id != "b_1" {
		t.Errorf("id = %q", id)
	}
}

func TestPerPageOptionsNestUnderScrapeOptions(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 201, `{"batchId":"b_1"}`)
	})
	_, err := c.StartBatch(context.Background(), []string{"https://a.example"}, &BatchOptions{
		ScrapeOptions: &ScrapeOptions{Formats: []Format{FormatMarkdown}},
	})
	if err != nil {
		t.Fatalf("StartBatch: %v", err)
	}
	body := cap.bodies[0]
	if _, nested := body["scrapeOptions"]; !nested {
		t.Error("scrapeOptions must be nested")
	}
	if _, flat := body["formats"]; flat {
		t.Error("formats must not appear at the top level")
	}
}

func TestPollsUntilTerminal(t *testing.T) {
	statuses := []string{"pending", "running", "completed"}
	i := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			writeJSON(w, 201, `{"batchId":"b_1"}`)
			return
		}
		s := statuses[min(i, len(statuses)-1)]
		i++
		writeJSON(w, 200, `{"success":true,"data":{"id":"b_1","status":"`+s+`"}}`)
	}))
	defer srv.Close()

	c, _ := New("hf_test", WithBaseURL(srv.URL), WithMaxRetries(0))
	var seen []string
	job, err := c.Batch(context.Background(), []string{"https://a.example"}, nil, &WaitOptions{
		PollInterval: time.Millisecond,
		OnProgress:   func(j *Job) { seen = append(seen, j.Status) },
	})
	if err != nil {
		t.Fatalf("Batch: %v", err)
	}
	if job.Status != "completed" || len(seen) != 3 {
		t.Errorf("status=%q seen=%v", job.Status, seen)
	}
}

func TestJobPollingTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			writeJSON(w, 201, `{"batchId":"b_1"}`)
			return
		}
		writeJSON(w, 200, `{"success":true,"data":{"id":"b_1","status":"running"}}`)
	}))
	defer srv.Close()

	c, _ := New("hf_test", WithBaseURL(srv.URL), WithMaxRetries(0))
	_, err := c.Batch(context.Background(), []string{"https://a.example"}, nil, &WaitOptions{
		PollInterval: time.Millisecond,
		Timeout:      5 * time.Millisecond,
	})
	var apiErr *Error
	if !errors.As(err, &apiErr) || !apiErr.IsTimeout() {
		t.Fatalf("want a timeout error, got %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestRateResultPostsTheRatingForTheLastRequest(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/web/feedback" {
			writeJSON(w, 200, `{"success":true,"data":{"id":"f1","requestId":"r1","rating":"bad"}}`)
			return
		}
		w.Header().Set("X-Request-Id", "r1")
		writeJSON(w, 200, `{"success":true,"data":{"markdown":"# hi"}}`)
	})
	if _, err := c.Scrape(context.Background(), "https://example.com", nil); err != nil {
		t.Fatalf("Scrape: %v", err)
	}
	if c.LastRequestID() != "r1" {
		t.Fatalf("LastRequestID = %q, want r1", c.LastRequestID())
	}
	out, err := c.RateResult(context.Background(), RateResultInput{
		RequestID: c.LastRequestID(),
		Rating:    RatingBad,
		Reasons:   []string{"missing_content"},
		Note:      "no prices",
	})
	if err != nil {
		t.Fatalf("RateResult: %v", err)
	}
	if out.ID != "f1" || out.Rating != RatingBad {
		t.Errorf("unexpected response %+v", out)
	}
	body := cap.bodies[1]
	if body["requestId"] != "r1" || body["rating"] != "bad" || body["note"] != "no prices" {
		t.Errorf("unexpected body %v", body)
	}
	if _, ok := body["url"]; ok {
		t.Errorf("url should be omitted, got %v", body["url"])
	}
}

func TestRateResultNeedsATarget(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{}}`)
	})
	if _, err := c.RateResult(context.Background(), RateResultInput{Rating: RatingGood}); err == nil {
		t.Fatal("expected an error without RequestID or URL")
	}
}

func TestSearchJobsPostsTheQueryAndReturnsTheList(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":[{"title":"Data Engineer","domain":"acme.com","board":"greenhouse","firstSeenAt":"2026-11-01T00:00:00Z"}]}`)
	})
	jobs, err := c.SearchJobs(context.Background(), "data engineer", &SearchJobsOptions{Location: "Berlin", Limit: 20})
	if err != nil {
		t.Fatalf("SearchJobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Title != "Data Engineer" || jobs[0].Company != nil {
		t.Errorf("unexpected jobs %+v", jobs)
	}
	if cap.requests[0].URL.Path != "/v1/web/index/jobs/search" {
		t.Errorf("path = %s", cap.requests[0].URL.Path)
	}
	body := cap.bodies[0]
	if body["query"] != "data engineer" || body["location"] != "Berlin" || body["limit"] != float64(20) {
		t.Errorf("unexpected body %v", body)
	}
	if _, ok := body["remoteOnly"]; ok {
		t.Errorf("remoteOnly should be left out when false: %v", body)
	}
}

func TestHiringChangesSendsDomainsAndAnRFC3339Time(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":[{"domain":"stripe.com","totals":{"open":4,"opened":2,"closed":1}}]}`)
	})
	since := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	out, err := c.HiringChanges(context.Background(), []string{"stripe.com"}, since, nil)
	if err != nil {
		t.Fatalf("HiringChanges: %v", err)
	}
	if len(out) != 1 || out[0].Totals.Opened != 2 {
		t.Errorf("unexpected changes %+v", out)
	}
	if cap.requests[0].URL.Path != "/v1/web/index/jobs/changes" {
		t.Errorf("path = %s", cap.requests[0].URL.Path)
	}
	if cap.bodies[0]["since"] != "2026-10-01T00:00:00Z" {
		t.Errorf("since = %v", cap.bodies[0]["since"])
	}
}

func TestNewsPostsTheQueryAndWindow(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{"query":"acme raises","status":"ok","results":[{"title":"Acme raises $10M","url":"https://example.com/a","publishedAt":"2026-11-04T10:00:00Z"}]}}`)
	})
	out, err := c.News(context.Background(), "acme raises", &NewsOptions{TimeRange: "day", Limit: 5})
	if err != nil {
		t.Fatalf("News: %v", err)
	}
	if out.Status != "ok" || len(out.Results) != 1 || out.Results[0].Source != nil {
		t.Errorf("unexpected news %+v", out)
	}
	if cap.requests[0].URL.Path != "/v1/web/news" {
		t.Errorf("path = %s", cap.requests[0].URL.Path)
	}
	body := cap.bodies[0]
	if body["query"] != "acme raises" || body["timeRange"] != "day" || body["limit"] != float64(5) {
		t.Errorf("unexpected body %v", body)
	}
	if _, ok := body["scrapeResults"]; ok {
		t.Errorf("scrapeResults should be left out when false: %v", body)
	}
}

func TestMapReadsEachLinkWithItsLastmod(t *testing.T) {
	c, cap := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, `{"success":true,"data":{"url":"https://example.com","links":[{"url":"https://example.com/a","lastmod":"2026-10-20T00:00:00.000Z"},{"url":"https://example.com/b","lastmod":null}],"count":2,"total":53,"truncated":false}}`)
	})
	out, err := c.Map(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(out.Links) != 2 || out.Links[0].URL != "https://example.com/a" || out.Links[0].Lastmod == nil || out.Links[1].Lastmod != nil {
		t.Errorf("unexpected links %+v", out.Links)
	}
	if out.Total != 53 || cap.requests[0].URL.Path != "/v1/web/map" {
		t.Errorf("unexpected map %+v path %s", out, cap.requests[0].URL.Path)
	}
}
