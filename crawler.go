package crawler

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

// Resource represents a URI resource.
type Resource struct {
	URL   string
	Links []string
}

// Result is used by the crawler's "out" channel to return
// a resource and any errors associated with fetching the
// the resource.
type Result struct {
	Resource
	Err error
}

// LinkGetter is a function type that takes a URL and returns
// a list of resource links.
type LinkGetter interface {
	Links(url string) ([]string, error)
}

// LinkGetterFunc is an adaptor for Getter to us to implment the
// interface as a function.
type LinkGetterFunc func(url string) ([]string, error)

// Links calls f(url).
func (f LinkGetterFunc) Links(url string) ([]string, error) {
	return f(url)
}

// LinkParser takes a URL domain and the HTTP response body.
type LinkParser func(host, body string) []string

// ParseLinksFunc uses a closure and reutrns simplistic LinkParser func type.
func ParseLinksFunc(host string) LinkParser {
	validLink := regexp.MustCompile(`(http|ftp|https)://(` + host + `)([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`)
	return func(host, body string) []string {
		return validLink.FindAllString(body, -1)
		// TODO: handle relative links and make more sophisticated
	}
}

// ResourceGetter is a simple implementation of the Getter interface.
type ResourceGetter struct {
	ParseFunc LinkParser
}

// Links method implements the LinkGetter interface.
func (r ResourceGetter) Links(u string) ([]string, error) {
	ln := make([]string, 0)
	resp, err := http.Get(u)
	if err != nil {
		return ln, err
	}
	if http.StatusOK != resp.StatusCode {
		return ln, errors.New("non-200 HTTP status code")
	}
	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return ln, err
	}
	pu, err := url.ParseRequestURI(u)
	if err != nil {
		return ln, err
	}
	return r.ParseFunc(pu.Host, string(b)), nil
}

// Crawl is used for crawling a domain. It takes a Getter func type to make requests.
func Crawl(url string, g LinkGetter, done <-chan struct{}) <-chan Result {
	visited := make(map[string]bool)
	in := make(chan string)
	out := make(chan Result)
	go func() {
		<-done
		close(out)
	}()
	doWork := func(url string) {
		links, err := g.Links(url)
		for _, link := range links {
			in <- link
		}
		out <- Result{Resource{url, links}, err}
	}
	// run a single (for thread saftey) goroutine to listen on the
	// "in" channel and check if URLs have already been visited.
	go func(in <-chan string) {
		for link := range in {
			if _, ok := visited[link]; ok {
				continue
			}
			go doWork(link)
			visited[link] = true
		}
	}(in)
	// push the initial URL into the "in" channel so
	// that we can begin processing
	in <- url
	return out
}
