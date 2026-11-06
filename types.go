package hydrafetch

type Format string

const (
	FormatMarkdown   Format = "markdown"
	FormatHTML       Format = "html"
	FormatRawHTML    Format = "rawHtml"
	FormatLinks      Format = "links"
	FormatStructured Format = "structured"
	FormatSummary    Format = "summary"
	FormatJSON       Format = "json"
	FormatBrand      Format = "brand"
)

type Usage struct {
	CreditsUsed      int    `json:"creditsUsed"`
	CreditsRemaining int    `json:"creditsRemaining"`
	Freshness        string `json:"freshness,omitempty"`
}

type Metadata map[string]any

type Location struct {
	Country   string   `json:"country,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

type ScrapeOptions struct {
	Formats            []Format          `json:"formats,omitempty"`
	MaxAge             int               `json:"maxAge,omitempty"`
	CacheOnly          bool              `json:"cacheOnly,omitempty"`
	StoreInCache       *bool             `json:"storeInCache,omitempty"`
	PreferStructure    bool              `json:"preferStructure,omitempty"`
	OnlyMainContent    *bool             `json:"onlyMainContent,omitempty"`
	IncludeTags        []string          `json:"includeTags,omitempty"`
	ExcludeTags        []string          `json:"excludeTags,omitempty"`
	RemoveBase64Images *bool             `json:"removeBase64Images,omitempty"`
	BlockAds           *bool             `json:"blockAds,omitempty"`
	IncludeLinks       *bool             `json:"includeLinks,omitempty"`
	WaitFor            int               `json:"waitFor,omitempty"`
	Timeout            int               `json:"timeout,omitempty"`
	Location           *Location         `json:"location,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
}

type ScrapeResult struct {
	URL        string   `json:"url"`
	FinalURL   string   `json:"finalUrl,omitempty"`
	Redirected bool     `json:"redirected,omitempty"`
	Status     int      `json:"status,omitempty"`
	Cached     bool     `json:"cached,omitempty"`
	Warning    string   `json:"warning,omitempty"`
	Metadata   Metadata `json:"metadata,omitempty"`
	Usage      *Usage   `json:"usage,omitempty"`
	Markdown   string   `json:"markdown,omitempty"`
	HTML       string   `json:"html,omitempty"`
	RawHTML    string   `json:"rawHtml,omitempty"`
	Links      []string `json:"links,omitempty"`
	Structured any      `json:"structured,omitempty"`
	Summary    string   `json:"summary,omitempty"`
	JSON       any      `json:"json,omitempty"`
}

type MapOptions struct {
	IncludeLinks          bool     `json:"includeLinks,omitempty"`
	Limit                 int      `json:"limit,omitempty"`
	Search                string   `json:"search,omitempty"`
	Sitemap               string   `json:"sitemap,omitempty"`
	SitemapInclude        []string `json:"sitemapInclude,omitempty"`
	SitemapExclude        []string `json:"sitemapExclude,omitempty"`
	IncludeSubdomains     bool     `json:"includeSubdomains,omitempty"`
	IgnoreQueryParameters bool     `json:"ignoreQueryParameters,omitempty"`
}

// MapLink is one URL on the site, with when its sitemap says it last changed.
type MapLink struct {
	URL     string  `json:"url"`
	Lastmod *string `json:"lastmod"`
}

type MapResult struct {
	URL       string    `json:"url"`
	Links     []MapLink `json:"links"`
	Count     int       `json:"count"`
	Total     int       `json:"total,omitempty"`
	Truncated bool      `json:"truncated,omitempty"`
}

type SearchOptions struct {
	Limit          int            `json:"limit,omitempty"`
	ScrapeResults  bool           `json:"scrapeResults,omitempty"`
	ScrapeOptions  *ScrapeOptions `json:"scrapeOptions,omitempty"`
	TimeRange      string         `json:"timeRange,omitempty"`
	Country        string         `json:"country,omitempty"`
	IncludeDomains []string       `json:"includeDomains,omitempty"`
	ExcludeDomains []string       `json:"excludeDomains,omitempty"`
}

type SearchItem struct {
	Title   string        `json:"title"`
	URL     string        `json:"url"`
	Snippet string        `json:"snippet"`
	Rank    int           `json:"rank"`
	Data    *ScrapeResult `json:"data,omitempty"`
}

type SearchResult struct {
	Query   string       `json:"query"`
	Status  string       `json:"status,omitempty"`
	Results []SearchItem `json:"results"`
}

type ExtractOptions struct {
	Schema          map[string]any    `json:"schema,omitempty"`
	Prompt          string            `json:"prompt,omitempty"`
	PreferStructure bool              `json:"preferStructure,omitempty"`
	EnableWebSearch bool              `json:"enableWebSearch,omitempty"`
	ShowSources     bool              `json:"showSources,omitempty"`
	ShowConfidence  bool              `json:"showConfidence,omitempty"`
	MergeEntities   bool              `json:"mergeEntities,omitempty"`
	MaxAge          int               `json:"maxAge,omitempty"`
	Location        *Location         `json:"location,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
}

type ExtractItem struct {
	URL    string         `json:"url"`
	Data   map[string]any `json:"data,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
	Error  string         `json:"error,omitempty"`
}

type ExtractResult struct {
	Results []ExtractItem `json:"results"`
	Sources []any         `json:"sources,omitempty"`
}

type CrawlOptions struct {
	Limit           int            `json:"limit,omitempty"`
	MaxDepth        int            `json:"maxDepth,omitempty"`
	IncludePaths    []string       `json:"includePaths,omitempty"`
	ExcludePaths    []string       `json:"excludePaths,omitempty"`
	AllowSubdomains bool           `json:"allowSubdomains,omitempty"`
	Webhook         string         `json:"webhook,omitempty"`
	ScrapeOptions   *ScrapeOptions `json:"scrapeOptions,omitempty"`
}

type BatchOptions struct {
	Webhook       string         `json:"webhook,omitempty"`
	ScrapeOptions *ScrapeOptions `json:"scrapeOptions,omitempty"`
}

type JobPage struct {
	URL          string        `json:"url"`
	RequestedURL string        `json:"requestedUrl,omitempty"`
	Status       string        `json:"status"`
	Depth        int           `json:"depth,omitempty"`
	Error        string        `json:"error,omitempty"`
	ErrorCode    string        `json:"errorCode,omitempty"`
	Retryable    bool          `json:"retryable,omitempty"`
	Data         *ScrapeResult `json:"data,omitempty"`
}

type Job struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind,omitempty"`
	Status      string    `json:"status"`
	SeedURL     string    `json:"seedUrl,omitempty"`
	Total       int       `json:"total,omitempty"`
	Completed   int       `json:"completed,omitempty"`
	Failed      int       `json:"failed,omitempty"`
	CreditsUsed int       `json:"creditsUsed,omitempty"`
	Pages       []JobPage `json:"pages,omitempty"`
}

func (j *Job) Done() bool {
	switch j.Status {
	case "completed", "failed", "cancelled":
		return true
	}
	return false
}

type LogoOptions struct {
	Theme string
	Type  string
}

type StyleguideOptions struct {
	MaxAge *int
}

type BrandResult map[string]any

type ScreenshotOptions struct {
	FullPage bool `json:"fullPage,omitempty"`
	WaitFor  int  `json:"waitFor,omitempty"`
	Timeout  int  `json:"timeout,omitempty"`
	MaxAge   int  `json:"maxAge,omitempty"`
}

type ScreenshotResult struct {
	URL            string `json:"url"`
	FinalURL       string `json:"finalUrl,omitempty"`
	Status         int    `json:"status,omitempty"`
	Screenshot     string `json:"screenshot"`
	ScreenshotType string `json:"screenshotType,omitempty"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	Cached         bool   `json:"cached,omitempty"`
}

type PageImage struct {
	Src    string `json:"src"`
	Alt    string `json:"alt,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type ImagesResult struct {
	URL      string      `json:"url"`
	FinalURL string      `json:"finalUrl,omitempty"`
	Images   []PageImage `json:"images"`
}

type Rating string

const (
	RatingGood Rating = "good"
	RatingBad  Rating = "bad"
)

type RateResultInput struct {
	RequestID string   `json:"requestId,omitempty"`
	URL       string   `json:"url,omitempty"`
	Rating    Rating   `json:"rating"`
	Reasons   []string `json:"reasons,omitempty"`
	Note      string   `json:"note,omitempty"`
}

type RateResultResponse struct {
	ID        string `json:"id"`
	RequestID string `json:"requestId"`
	Rating    Rating `json:"rating"`
}

// SearchJobsOptions narrows a jobs search. Each 10 jobs returned cost 1 credit.
type SearchJobsOptions struct {
	Location            string `json:"location,omitempty"`
	RemoteOnly          bool   `json:"remoteOnly,omitempty"`
	PostedWithinDays    int    `json:"postedWithinDays,omitempty"`
	Limit               int    `json:"limit,omitempty"`
	IncludeDescriptions bool   `json:"includeDescriptions,omitempty"`
}

type JobPay struct {
	Min      *float64 `json:"min"`
	Max      *float64 `json:"max"`
	Currency *string  `json:"currency"`
	Interval *string  `json:"interval"`
}

// OpenJob is one open job from a company career site or hiring platform.
type OpenJob struct {
	Title          string  `json:"title"`
	Company        *string `json:"company"`
	Domain         string  `json:"domain"`
	Board          string  `json:"board"`
	Department     *string `json:"department"`
	Location       *string `json:"location"`
	Remote         *bool   `json:"remote"`
	EmploymentType *string `json:"employmentType"`
	PostedAt       *string `json:"postedAt"`
	FirstSeenAt    string  `json:"firstSeenAt"`
	URL            *string `json:"url"`
	ApplyURL       *string `json:"applyUrl"`
	Pay            *JobPay `json:"pay"`
	Description    *string `json:"description,omitempty"`
}

type HiringChangesOptions struct {
	IncludeDescriptions bool `json:"includeDescriptions,omitempty"`
}

type HiringCounts struct {
	Open   int `json:"open"`
	Opened int `json:"opened"`
	Closed int `json:"closed"`
}

type HiringChangePosting struct {
	Title       string  `json:"title"`
	Department  *string `json:"department"`
	Location    *string `json:"location"`
	Seniority   string  `json:"seniority"`
	Board       string  `json:"board"`
	URL         *string `json:"url"`
	PostedAt    *string `json:"postedAt"`
	FirstSeenAt string  `json:"firstSeenAt"`
	ClosedAt    *string `json:"closedAt"`
	Description *string `json:"description,omitempty"`
}

// CompanyHiringChanges is what changed in one company's hiring since a date.
type CompanyHiringChanges struct {
	Domain       string                  `json:"domain"`
	Company      *string                 `json:"company"`
	LastReadAt   *string                 `json:"lastReadAt"`
	Since        string                  `json:"since"`
	Totals       HiringCounts            `json:"totals"`
	ByDepartment map[string]HiringCounts `json:"byDepartment"`
	BySeniority  map[string]HiringCounts `json:"bySeniority"`
	ByLocation   map[string]HiringCounts `json:"byLocation"`
	Opened       []HiringChangePosting   `json:"opened"`
	Closed       []HiringChangePosting   `json:"closed"`
	Truncated    bool                    `json:"truncated"`
}

// NewsOptions narrows a news search. A search costs 1 credit, and one that finds nothing is free.
type NewsOptions struct {
	Limit          int            `json:"limit,omitempty"`
	TimeRange      string         `json:"timeRange,omitempty"`
	Country        string         `json:"country,omitempty"`
	IncludeDomains []string       `json:"includeDomains,omitempty"`
	ExcludeDomains []string       `json:"excludeDomains,omitempty"`
	ScrapeResults  bool           `json:"scrapeResults,omitempty"`
	ScrapeOptions  *ScrapeOptions `json:"scrapeOptions,omitempty"`
}

// NewsArticle is one article, with the publisher's own URL and when it was published.
type NewsArticle struct {
	Title       string        `json:"title"`
	URL         string        `json:"url"`
	Source      *string       `json:"source"`
	PublishedAt string        `json:"publishedAt"`
	Snippet     *string       `json:"snippet"`
	Data        *ScrapeResult `json:"data"`
}

// NewsResult holds the articles, newest first. Status is "ok" or "no_results".
type NewsResult struct {
	Query   string        `json:"query"`
	Status  string        `json:"status"`
	Results []NewsArticle `json:"results"`
}
