package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/burrbd/crawler"
)

var (
	timeout uint
	u       string
)

func init() {
	flag.StringVar(&u, "url", "", "(required) URL to crawl (eg, https://google.com)")
	flag.UintVar(&timeout, "t", 3, "Timeout in seconds when retrieving resources")
}

func main() {
	flag.Parse()
	if u == "" {
		flag.Usage()
		os.Exit(1)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	pu, err := url.ParseRequestURI(u)
	if err != nil {
		log.Fatal(err)
	}
	linkGetter := crawler.ResourceGetter{
		ParseFunc: crawler.ParseLinksFunc(pu.Host),
	}
	done := make(chan struct{})
	out := crawler.Crawl(u, linkGetter, done)

loop:
	for {
		select {
		case res := <-out:
			fmt.Println(res.URL)
			if res.Err != nil {
				fmt.Println(res.Err)
			}
			for _, ln := range res.Links {
				fmt.Println(fmt.Sprintf("  -> %s", ln))
			}
		case <-time.Tick(time.Duration(timeout) * time.Second):
			done <- struct{}{}
			break loop
		case <-sigs:
			fmt.Println("goodbye!")
			done <- struct{}{}
			break loop
		}
	}
}
