package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
)

var initialURLs = [...]string{
	"https://en.wikipedia.org/wiki/Main_Page",
	"https://www.bbc.com",
	"https://stackoverflow.com",
	"https://www.imdb.com",
	"https://github.com",
	"https://www.aljazeera.com/",
}

type Crawler struct {
	collector    *colly.Collector
	pageChan     chan Page
	urlChan      chan string
	visitCounter *VisitCounter
}

func NewCrawler(pageChan chan Page, counter *VisitCounter) *Crawler {
	urlChan := make(chan string, 1000)

	c := colly.NewCollector(
		colly.MaxDepth(3), // Increase this to allow deeper crawling
		colly.Async(true),  // Enable asynchronous crawling
	)
	// Limit the number of concurrent requests
	c.Limit(&colly.LimitRule{
		DomainGlob: "*", 
		Parallelism: 2, 
		Delay: 1 * time.Second,
		RandomDelay: 2 * time.Second,
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		urlChan <- link // Send new URLs to the channel
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("ERROR %v (%v): %v\n", r.StatusCode, r.Request.URL, err.Error())
	})

	c.OnResponse(func(r *colly.Response) {
		pageChan <- Page{URL: r.Request.URL.String(), Body: string(r.Body)}
		counter.Increment()
	})

	return &Crawler{c, pageChan, urlChan, counter}
}

func (c *Crawler) Crawl(ctx context.Context) {
	file, err := os.Create("visited_urls.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, url := range initialURLs {
		c.collector.Visit(url)
		_, err = writer.WriteString(url + "\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	go func() {
		visited := make(map[string]bool)
		for url := range c.urlChan {
			if !visited[url] {
				visited[url] = true
				_, err = writer.WriteString(url + "\n")
				if err != nil {
					fmt.Println(err)
					return
				}
				go func(url string) {
					c.collector.Visit(url)
				}(url)
			}
		}
	}()

	// Wait till either context closes or wait group finishes.
	doneChan := make(chan struct{})
	go func() {
		c.collector.Wait()
		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		fmt.Println("Context Expired")
	case <-doneChan:
		fmt.Println("Collector finished")
	}

}
