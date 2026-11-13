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
	DismissCookies     *bool             `json:"dismissCookies,omitempty"`
	DismissPopups      *bool             `json:"dismissPopups,omitempty"`
	Render             bool              `json:"render,omitempty"`
	IncludeLinks       *bool             `json:"includeLinks,omitempty"`
	WaitFor            int               `json:"waitFor,omitempty"`
	Timeout            int               `json:"timeout,omitempty"`
	Location           *Location         `json:"location,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Method             string            `json:"method,omitempty"`
	Body               string            `json:"body,omitempty"`
	JSONOptions        map[string]any    `json:"jsonOptions,omitempty"`
}

type MapOptions struct {
	IncludeLinks          bool   `json:"includeLinks,omitempty"`
	Limit                 int    `json:"limit,omitempty"`
	Search                string `json:"search,omitempty"`
	Sitemap               string `json:"sitemap,omitempty"`
	SitemapInclude        string `json:"sitemapInclude,omitempty"`
	SitemapExclude        string `json:"sitemapExclude,omitempty"`
	IncludeSubdomains     bool   `json:"includeSubdomains,omitempty"`
	IgnoreQueryParameters bool   `json:"ignoreQueryParameters,omitempty"`
	Order                 string `json:"order,omitempty"`
	Before                string `json:"before,omitempty"`
	Offset                int    `json:"offset,omitempty"`
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

type CrawlOptions struct {
	Limit                 int            `json:"limit,omitempty"`
	MaxDepth              int            `json:"maxDepth,omitempty"`
	IncludePaths          []string       `json:"includePaths,omitempty"`
	ExcludePaths          []string       `json:"excludePaths,omitempty"`
	AllowSubdomains       bool           `json:"allowSubdomains,omitempty"`
	AllowExternalLinks    bool           `json:"allowExternalLinks,omitempty"`
	IgnoreQueryParameters bool           `json:"ignoreQueryParameters,omitempty"`
	Sitemap               string         `json:"sitemap,omitempty"`
	Webhook               *Webhook       `json:"webhook,omitempty"`
	ScrapeOptions         *ScrapeOptions `json:"scrapeOptions,omitempty"`
}

type BatchOptions struct {
	Webhook       *Webhook       `json:"webhook,omitempty"`
	ScrapeOptions *ScrapeOptions `json:"scrapeOptions,omitempty"`
}

type LogoOptions struct {
	Theme string
	Type  string
}

type StyleguideOptions struct {
	MaxAge *int
}

type ScreenshotOptions struct {
	FullPage       bool   `json:"fullPage,omitempty"`
	WaitFor        int    `json:"waitFor,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
	MaxAge         int    `json:"maxAge,omitempty"`
	Viewport       string `json:"viewport,omitempty"`
	ViewportWidth  int    `json:"viewportWidth,omitempty"`
	ViewportHeight int    `json:"viewportHeight,omitempty"`
	DismissCookies *bool  `json:"dismissCookies,omitempty"`
	DismissPopups  *bool  `json:"dismissPopups,omitempty"`
}

type RateResultInput struct {
	RequestID string   `json:"requestId,omitempty"`
	URL       string   `json:"url,omitempty"`
	Rating    Rating   `json:"rating"`
	Reasons   []string `json:"reasons,omitempty"`
	Note      string   `json:"note,omitempty"`
}

// SearchJobsOptions narrows a jobs search. Each 10 jobs returned cost 1 credit.
type SearchJobsOptions struct {
	Location            string `json:"location,omitempty"`
	RemoteOnly          bool   `json:"remoteOnly,omitempty"`
	PostedWithinDays    int    `json:"postedWithinDays,omitempty"`
	Limit               int    `json:"limit,omitempty"`
	IncludeDescriptions bool   `json:"includeDescriptions,omitempty"`
}

type HiringChangesOptions struct {
	IncludeDescriptions bool `json:"includeDescriptions,omitempty"`
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
	MaxAge         int            `json:"maxAge,omitempty"`
}

type Usage = WebUsageDto

type Metadata = WebScrapeMetadataDto

type Location struct {
	Country   string   `json:"country,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

type ScrapeResult = WebScrapeDataDto

type MapLink = MapURLEntryDto

type MapResult = MapResultDto

type SearchItem = SearchResultItemDto

type SearchResult = SearchResultDto

type ExtractItem = ExtractItemResultDto

type ExtractResult = ExtractResultDto

type Webhook = WebhookDto

type JobPage = CrawlPageResultDto

type Job = CrawlStatusResponseDto

func (j *CrawlStatusResponseDto) Done() bool {
	switch string(j.Status) {
	case "completed", "failed", "cancelled":
		return true
	}
	return false
}

type BrandResult = WebBrandDataDto

type LogoResult = LogoDescriptionDto

type StyleguideResult = WebStyleguideDataDto

type ScreenshotResult = WebScreenshotDataDto

type PageImage = WebImageDto

type ImagesResult = WebImagesDataDto

type Rating = ResultFeedbackDataDtoRating

const (
	RatingGood Rating = "good"
	RatingBad  Rating = "bad"
)

type RateResultResponse = ResultFeedbackDataDto

type JobPay = JobsIndexPayDto

type OpenJob = JobsIndexJobDto

type HiringCounts = JobsIndexCountsDto

type HiringChangePosting = JobsIndexChangePostingDto

type CompanyHiringChanges = JobsIndexCompanyChangesDto

type NewsArticle = NewsResultItemDto

type NewsResult = NewsResultDto
