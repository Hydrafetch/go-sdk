package hydrafetch

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func jsonKind(t reflect.Type) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice, reflect.Array:
		return "array"
	default:
		return "object"
	}
}

func jsonFields(v any) map[string]string {
	out := map[string]string{}
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			continue
		}
		out[name] = jsonKind(f.Type)
	}
	return out
}

func TestOptionsMatchTheAPI(t *testing.T) {
	cases := []struct {
		name       string
		options    any
		request    any
		positional []string
	}{
		{"ScrapeOptions", ScrapeOptions{}, ScrapeRequestDto{}, []string{"url", "async"}},
		{"MapOptions", MapOptions{}, MapRequestDto{}, []string{"url"}},
		{"SearchOptions", SearchOptions{}, SearchRequestDto{}, []string{"query"}},
		{"ExtractOptions", ExtractOptions{}, ExtractRequestDto{}, []string{"urls"}},
		{"CrawlOptions", CrawlOptions{}, CrawlRequestDto{}, []string{"url"}},
		{"BatchOptions", BatchOptions{}, BatchRequestDto{}, []string{"urls"}},
		{"ScreenshotOptions", ScreenshotOptions{}, ScreenshotRequestDto{}, []string{"url"}},
		{"RateResultInput", RateResultInput{}, ResultFeedbackDto{}, nil},
		{"SearchJobsOptions", SearchJobsOptions{}, JobsIndexSearchRequestDto{}, []string{"query"}},
		{"HiringChangesOptions", HiringChangesOptions{}, JobsIndexChangesRequestDto{}, []string{"domains", "since"}},
		{"NewsOptions", NewsOptions{}, NewsRequestDto{}, []string{"query"}},
		{"Location", Location{}, LocationDto{}, nil},
	}
	for _, c := range cases {
		mine, api := jsonFields(c.options), jsonFields(c.request)
		for _, p := range c.positional {
			delete(api, p)
		}
		var problems []string
		for name, kind := range api {
			got, ok := mine[name]
			if !ok {
				problems = append(problems, name+" is in the API but not in "+c.name)
			} else if got != kind {
				problems = append(problems, name+" is "+got+" here but "+kind+" in the API")
			}
		}
		for name := range mine {
			if _, ok := api[name]; !ok {
				problems = append(problems, name+" is in "+c.name+" but the API does not take it")
			}
		}
		sort.Strings(problems)
		for _, p := range problems {
			t.Errorf("%s: %s", c.name, p)
		}
	}
}
