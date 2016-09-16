package crawler_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/burrbd/crawler"
)

var linksToParse = `
some content https://mydomain.com/
src="https://mydomain.com/path/to/content?with=query#and-heading"
href = "http://mydomain.com/path/to/content?with=query#and-heading"
rel="http://not.mydomain.com/ more content
`

func TestLinkParserFunc(t *testing.T) {
	fn := crawler.ParseLinksFunc("mydomain.com")
	exp := []string{
		"https://mydomain.com/",
		"https://mydomain.com/path/to/content?with=query#and-heading",
		"http://mydomain.com/path/to/content?with=query#and-heading",
	}
	act := fn("mydomain.com", linksToParse)
	if len(exp) != len(act) {
		t.Errorf("expected %d links, got %d", len(exp), len(act))
	}
	for _, l := range exp {
		if !inSlice(l, act) {
			t.Errorf("expected link not found: %s", l)
		}
	}
}

func TestResourceGetter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
	exp := []string{"first link", "second link"}
	rg := crawler.ResourceGetter{ParseFunc: func(host, body string) []string {
		return exp
	}}
	act, err := rg.Links(server.URL)
	if err != nil {
		t.Error(err)
	}
	if 2 != len(act) {
		t.Errorf("expected 2 links, got %d", len(act))
	}
	if exp[0] != act[0] {
		t.Errorf("expected first link %s, go %s", exp[0], act[0])
	}
	if exp[1] != act[1] {
		t.Errorf("expected second link %s, go %s", exp[1], act[1])
	}
}

func TestResourceGetterNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("some body"))
	}))
	rg := crawler.ResourceGetter{}
	act, err := rg.Links(server.URL)
	if err == nil {
		t.Error("expected error")
	}
	if 0 != len(act) {
		t.Errorf("expected length to be 0, got %d", len(act))
	}
}

func TestResourceGetterWithInvalidURL(t *testing.T) {
	rg := crawler.ResourceGetter{}
	act, err := rg.Links("invalid")
	if err == nil {
		t.Error("expected error")
	}
	if 0 != len(act) {
		t.Errorf("expected length to be 0, got %d", len(act))
	}
}

func TestLinkGetterFunc(t *testing.T) {
	exp := []string{"/my/link"}
	fn := crawler.LinkGetterFunc(func(url string) ([]string, error) {
		return exp, nil
	})
	act, _ := fn.Links("a.url")
	for _, l := range act {
		if !inSlice(l, exp) {
			t.Errorf("link not found: %s", l)
		}
	}
}

func TestDoCrawl(t *testing.T) {
	done := make(chan struct{})
	crawler.Crawl("a.link", &fakeLinkGetter{}, done)
}

type fakeLinkGetter struct {
	sync.Mutex
	incr int
}

func (g *fakeLinkGetter) Links(url string) ([]string, error) {
	return []string{"a.link/a", "a.link/b"}, nil
}

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
